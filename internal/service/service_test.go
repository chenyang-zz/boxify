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
	app := application.New(application.Options{
		LogLevel: slog.LevelInfo,
	})
	ctx := context.Background()
	logger.Init(slog.LevelInfo)
	ts := NewTerminalService(NewServiceDeps(app, nil))
	gs := NewGitService(NewServiceDeps(app, nil))

	ts.ServiceStartup(ctx, application.ServiceOptions{})
	gs.ServiceStartup(ctx, application.ServiceOptions{})

	time.Sleep(2 * time.Second)

	ts.Create(terminal.TerminalConfig{
		ID:             "123",
		Shell:          terminal.ShellTypeZsh,
		Rows:           0,
		Cols:           0,
		WorkPath:       "/Users/sheepzhao/WorkSpace/Boxify",
		InitialCommand: "fnm use 16",
	})

	repoPath := "/Users/sheepzhao/WorkSpace/Boxify"
	gs.RegisterRepo(repoPath, repoPath)
	gs.StartRepoWatch(repoPath, 200)

	time.Sleep(3 * time.Second)
	ts.WriteCommand("123", "echo Hello, World!\n")
	time.Sleep(3 * time.Second)
	ts.ServiceShutdown()
	gs.ServiceShutdown()
}

func TestGitService(t *testing.T) {
	ctx := context.Background()
	logger.Init(slog.LevelInfo)
	gs := NewGitService(NewServiceDeps(application.New(application.Options{
		LogLevel: slog.LevelDebug,
	}), nil))
	gs.ServiceStartup(ctx, application.ServiceOptions{})
	repoPath := "/Users/sheepzhao/WorkSpace/Boxify"
	gs.RegisterRepo(repoPath, repoPath)
	gs.StartRepoWatch(repoPath, 200)
	time.Sleep(10 * time.Second)
	gs.ServiceShutdown()
}
