package service

import (
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"serverManage/internal/model"
	"serverManage/internal/storage"

	"github.com/google/uuid"
)

type Manager struct {
	store   *storage.Storage
	process map[string]*exec.Cmd
	mu      sync.RWMutex
}

func NewManager(store *storage.Storage) *Manager {
	return &Manager{
		store:   store,
		process: make(map[string]*exec.Cmd),
	}
}

func (m *Manager) GetAll() ([]model.Service, error) {
	list, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	// 更新每个服务的实际状态
	for i := range list.Services {
		m.updateStatus(&list.Services[i])
	}

	return list.Services, nil
}

func (m *Manager) Get(id string) (*model.Service, error) {
	list, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	for i := range list.Services {
		if list.Services[i].ID == id {
			m.updateStatus(&list.Services[i])
			return &list.Services[i], nil
		}
	}

	return nil, fmt.Errorf("service not found: %s", id)
}

func (m *Manager) Create(req model.ServiceRequest) (*model.Service, error) {
	list, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	// 检查名称是否重复
	for _, s := range list.Services {
		if s.Name == req.Name {
			return nil, fmt.Errorf("service already exists: %s", req.Name)
		}
	}

	svc := model.Service{
		ID:            uuid.New().String(),
		Name:          req.Name,
		StartScript:   req.StartScript,
		StopScript:    req.StopScript,
		RestartScript: req.RestartScript,
		WorkDir:       req.WorkDir,
		Port:          req.Port,
		AutoStart:     req.AutoStart,
		Status:        model.StatusStopped,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	list.Services = append(list.Services, svc)
	if err := m.store.Save(list); err != nil {
		return nil, err
	}

	return &svc, nil
}

func (m *Manager) Update(id string, req model.ServiceRequest) (*model.Service, error) {
	list, err := m.store.Load()
	if err != nil {
		return nil, err
	}

	for i := range list.Services {
		if list.Services[i].ID == id {
			list.Services[i].Name = req.Name
			list.Services[i].StartScript = req.StartScript
			list.Services[i].StopScript = req.StopScript
			list.Services[i].RestartScript = req.RestartScript
			list.Services[i].WorkDir = req.WorkDir
			list.Services[i].Port = req.Port
			list.Services[i].AutoStart = req.AutoStart
			list.Services[i].UpdatedAt = time.Now()

			if err := m.store.Save(list); err != nil {
				return nil, err
			}

			return &list.Services[i], nil
		}
	}

	return nil, fmt.Errorf("service not found: %s", id)
}

func (m *Manager) Delete(id string) error {
	list, err := m.store.Load()
	if err != nil {
		return err
	}

	for i := range list.Services {
		if list.Services[i].ID == id {
			list.Services = append(list.Services[:i], list.Services[i+1:]...)
			return m.store.Save(list)
		}
	}

	return fmt.Errorf("service not found: %s", id)
}

func (m *Manager) Start(id string) error {
	svc, err := m.Get(id)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已在运行
	if m.isRunning(svc.Name) {
		return fmt.Errorf("service already running: %s", svc.Name)
	}

	cmd := exec.Command("bash", "-c", svc.StartScript)
	if svc.WorkDir != "" {
		cmd.Dir = svc.WorkDir
	}

	// 设置进程组
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	m.process[svc.Name] = cmd

	// 更新状态
	list, _ := m.store.Load()
	for i := range list.Services {
		if list.Services[i].ID == id {
			list.Services[i].Status = model.StatusRunning
			list.Services[i].PID = cmd.Process.Pid
			list.Services[i].UpdatedAt = time.Now()
			m.store.Save(list)
			break
		}
	}

	return nil
}

func (m *Manager) Stop(id string) error {
	svc, err := m.Get(id)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning(svc.Name) {
		return fmt.Errorf("service not running: %s", svc.Name)
	}

	// 执行停止脚本
	cmd := exec.Command("bash", "-c", svc.StopScript)
	if svc.WorkDir != "" {
		cmd.Dir = svc.WorkDir
	}
	cmd.Run()

	// 杀死进程组
	if proc, ok := m.process[svc.Name]; ok && proc.Process != nil {
		pgid, _ := syscall.Getpgid(proc.Process.Pid)
		syscall.Kill(-pgid, syscall.SIGTERM)
	}

	delete(m.process, svc.Name)

	// 更新状态
	list, _ := m.store.Load()
	for i := range list.Services {
		if list.Services[i].ID == id {
			list.Services[i].Status = model.StatusStopped
			list.Services[i].PID = 0
			list.Services[i].UpdatedAt = time.Now()
			m.store.Save(list)
			break
		}
	}

	return nil
}

func (m *Manager) Restart(id string) error {
	if err := m.Stop(id); err != nil {
		// 忽略停止错误，继续启动
	}
	return m.Start(id)
}

func (m *Manager) isRunning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if cmd, ok := m.process[name]; ok {
		if cmd.Process != nil {
			err := cmd.Process.Signal(syscall.Signal(0))
			return err == nil
		}
	}
	return false
}

func (m *Manager) updateStatus(svc *model.Service) {
	if m.isRunning(svc.Name) {
		svc.Status = model.StatusRunning
	} else {
		svc.Status = model.StatusStopped
	}
}