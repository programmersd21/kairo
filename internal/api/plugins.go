package api

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/programmersd21/kairo/internal/config"
	"github.com/programmersd21/kairo/internal/util"
)

func getPluginDir() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	dir := strings.TrimSpace(cfg.Plugins.Dir)
	if dir == "" {
		stateDir, err := util.AppStateDir("kairo")
		if err != nil {
			return "", err
		}
		dir = filepath.Join(stateDir, "plugins")
	}
	_ = os.MkdirAll(dir, 0755)
	return dir, nil
}

func (api *TaskAPI) handlePluginList(ctx context.Context) Response {
	dir, err := getPluginDir()
	if err != nil {
		return Response{Success: false, Error: err.Error()}
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return Response{Success: false, Error: err.Error()}
	}
	var plugins []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".lua") {
			plugins = append(plugins, f.Name())
		}
	}
	return Response{Success: true, Data: plugins}
}

func (api *TaskAPI) handlePluginGet(ctx context.Context, payload json.RawMessage) Response {
	type P struct {
		Name string `json:"name"`
	}
	var p P
	_ = json.Unmarshal(payload, &p)
	if p.Name == "" {
		return Response{Success: false, Error: "missing plugin name"}
	}
	if !strings.HasSuffix(p.Name, ".lua") {
		p.Name += ".lua"
	}

	dir, err := getPluginDir()
	if err != nil {
		return Response{Success: false, Error: err.Error()}
	}
	content, err := os.ReadFile(filepath.Join(dir, filepath.Base(p.Name)))
	if err != nil {
		return Response{Success: false, Error: err.Error()}
	}
	return Response{Success: true, Data: string(content)}
}

func (api *TaskAPI) handlePluginWrite(ctx context.Context, payload json.RawMessage) Response {
	type P struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	var p P
	_ = json.Unmarshal(payload, &p)
	if p.Name == "" {
		return Response{Success: false, Error: "missing plugin name"}
	}
	if !strings.HasSuffix(p.Name, ".lua") {
		p.Name += ".lua"
	}

	dir, err := getPluginDir()
	if err != nil {
		return Response{Success: false, Error: err.Error()}
	}
	err = os.WriteFile(filepath.Join(dir, filepath.Base(p.Name)), []byte(p.Content), 0644)
	if err != nil {
		return Response{Success: false, Error: err.Error()}
	}
	return Response{Success: true, Data: "plugin saved"}
}

func (api *TaskAPI) handlePluginDelete(ctx context.Context, payload json.RawMessage) Response {
	type P struct {
		Name string `json:"name"`
	}
	var p P
	_ = json.Unmarshal(payload, &p)
	if p.Name == "" {
		return Response{Success: false, Error: "missing plugin name"}
	}
	if !strings.HasSuffix(p.Name, ".lua") {
		p.Name += ".lua"
	}

	dir, err := getPluginDir()
	if err != nil {
		return Response{Success: false, Error: err.Error()}
	}
	err = os.Remove(filepath.Join(dir, filepath.Base(p.Name)))
	if err != nil {
		return Response{Success: false, Error: err.Error()}
	}
	return Response{Success: true, Data: "plugin deleted"}
}
