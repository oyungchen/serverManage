package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"serverManage/internal/model"
)

type Storage struct {
	dataFile string
}

func New(dataFile string) (*Storage, error) {
	// 确保目录存在
	dir := filepath.Dir(dataFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 如果文件不存在，创建空文件
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		empty := model.ServiceList{Services: []model.Service{}}
		data, _ := json.MarshalIndent(empty, "", "  ")
		os.WriteFile(dataFile, data, 0644)
	}

	return &Storage{dataFile: dataFile}, nil
}

func (s *Storage) Load() (*model.ServiceList, error) {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return nil, err
	}

	var list model.ServiceList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	return &list, nil
}

func (s *Storage) Save(list *model.ServiceList) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataFile, data, 0644)
}