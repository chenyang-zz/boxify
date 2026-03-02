// Copyright 2026 chenyang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/chenyang-zz/boxify/internal/logger"
	"github.com/chenyang-zz/boxify/internal/terminal"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestExample(t *testing.T) {
	ctx := context.Background()
	logger.Init(slog.LevelInfo)
	service := NewTerminalService(NewServiceDeps(application.New(application.Options{
		LogLevel: slog.LevelDebug,
	}), nil))
	service.ServiceStartup(ctx, application.ServiceOptions{})
	service.Create(terminal.TerminalConfig{
		ID:             "123",
		Shell:          terminal.ShellTypeZsh,
		Rows:           0,
		Cols:           0,
		WorkPath:       "/Users/sheepzhao/WorkSpace/Boxify",
		InitialCommand: "",
	})
	time.Sleep(10 * time.Second)
	service.ServiceShutdown()
}

func TestGitService(t *testing.T) {
	ctx := context.Background()
	logger.Init(slog.LevelInfo)
	service := NewGitService(NewServiceDeps(application.New(application.Options{
		LogLevel: slog.LevelDebug,
	}), nil))
	service.ServiceStartup(ctx, application.ServiceOptions{})
	repoPath := "/Users/sheepzhao/WorkSpace/Boxify"
	service.RegisterRepo(repoPath, repoPath)
	service.StartRepoWatch(repoPath, 200)
	time.Sleep(10 * time.Second)
	service.ServiceShutdown()
}
