package model

import (
	"time"
)

type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusStopped ServiceStatus = "stopped"
	StatusError   ServiceStatus = "error"
)

type Service struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	StartScript   string        `json:"startScript"`
	StopScript    string        `json:"stopScript"`
	RestartScript string        `json:"restartScript"`
	WorkDir       string        `json:"workDir"`
	Port          int           `json:"port"`
	AutoStart     bool          `json:"autoStart"`
	Status        ServiceStatus `json:"status"`
	PID           int           `json:"pid,omitempty"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type ServiceList struct {
	Services []Service `json:"services"`
}

type ServiceRequest struct {
	Name          string `json:"name"`
	StartScript   string `json:"startScript"`
	StopScript    string `json:"stopScript"`
	RestartScript string `json:"restartScript"`
	WorkDir       string `json:"workDir"`
	Port          int    `json:"port"`
	AutoStart     bool   `json:"autoStart"`
}

type DiscoverRequest struct {
	Dirs []string `json:"dirs"`
}

type DiscoverResult struct {
	Services []ServiceRequest `json:"services"`
}