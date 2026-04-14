package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const launchAgentsDir = "Library/LaunchAgents"

type Launcher struct {
	homeDir string
}

func New() *Launcher {
	home := os.Getenv("HOME")
	return &Launcher{homeDir: home}
}

func (l *Launcher) GetPlistPath(serviceName string) string {
	return filepath.Join(l.homeDir, launchAgentsDir, fmt.Sprintf("com.servermanage.%s.plist", serviceName))
}

func (l *Launcher) EnableAutoStart(serviceName, workDir, startScript string) error {
	// 确保目录存在
	dir := filepath.Join(l.homeDir, launchAgentsDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	plistPath := l.GetPlistPath(serviceName)

	tmpl := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.servermanage.{{.Name}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>/bin/bash</string>
        <string>-c</string>
        <string>cd {{.WorkDir}} && {{.StartScript}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
    <key>StandardOutPath</key>
    <string>{{.LogDir}}/{{.Name}}.out.log</string>
    <key>StandardErrorPath</key>
    <string>{{.LogDir}}/{{.Name}}.err.log</string>
</dict>
</plist>`

	data := struct {
		Name        string
		WorkDir     string
		StartScript string
		LogDir      string
	}{
		Name:        serviceName,
		WorkDir:     workDir,
		StartScript: startScript,
		LogDir:      filepath.Join(l.homeDir, ".serverManage", "logs"),
	}

	t, err := template.New("plist").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(plistPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := t.Execute(f, data); err != nil {
		return err
	}

	return nil
}

func (l *Launcher) DisableAutoStart(serviceName string) error {
	plistPath := l.GetPlistPath(serviceName)

	// 先 unload
	exec.Command("launchctl", "unload", plistPath).Run()

	// 删除文件
	if _, err := os.Stat(plistPath); err == nil {
		return os.Remove(plistPath)
	}

	return nil
}

func (l *Launcher) IsAutoStartEnabled(serviceName string) bool {
	_, err := os.Stat(l.GetPlistPath(serviceName))
	return err == nil
}

func (l *Launcher) GetAllManagedServices() ([]string, error) {
	dir := filepath.Join(l.homeDir, launchAgentsDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var services []string
	prefix := "com.servermanage."
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".plist") {
			services = append(services, strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".plist"))
		}
	}

	return services, nil
}