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

package terminal

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestApplyShellEnvConfigUnix(t *testing.T) {
	loader := NewShellEnvironmentLoader(testLogger, "darwin")
	env := map[string]string{
		"HOME": "/tmp/home",
		"PATH": "/usr/bin",
	}

	content := []byte(`
export PATH="$HOME/.local/bin:$PATH"
FOO=bar
`)
	loader.applyShellEnvConfig(content, ShellTypeBash, env)

	if got := loader.GetEnv(env, "PATH"); got != "/tmp/home/.local/bin:/usr/bin" {
		t.Fatalf("unexpected PATH: %s", got)
	}
	if got := loader.GetEnv(env, "FOO"); got != "bar" {
		t.Fatalf("unexpected FOO: %s", got)
	}
}

func TestApplyShellEnvConfigPowerShell(t *testing.T) {
	loader := NewShellEnvironmentLoader(testLogger, "windows")
	env := map[string]string{
		"PATH": "C:\\Windows\\System32",
	}

	content := []byte(`
$env:Path = "C:\Tools;$env:Path"
`)
	loader.applyShellEnvConfig(content, ShellTypePwsh, env)

	if got := loader.GetEnv(env, "PATH"); got != "C:\\Tools;C:\\Windows\\System32" {
		t.Fatalf("unexpected PATH: %s", got)
	}
}

func TestListExecutableCommandsForShellUsesShellConfigPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this test validates unix shell config loading")
	}

	homeDir := t.TempDir()
	baseDir := t.TempDir()
	customDir := filepath.Join(homeDir, "custom", "bin")
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		t.Fatalf("create custom dir failed: %v", err)
	}

	mustWriteFileWithMode(t, filepath.Join(baseDir, "basecmd"), 0o755)
	mustWriteFileWithMode(t, filepath.Join(customDir, "customcmd"), 0o755)

	rcPath := filepath.Join(homeDir, ".bashrc")
	if err := os.WriteFile(rcPath, []byte(`export PATH="$HOME/custom/bin:$PATH"`), 0o644); err != nil {
		t.Fatalf("write .bashrc failed: %v", err)
	}

	t.Setenv("HOME", homeDir)
	t.Setenv("PATH", baseDir)

	scanner := NewPathCommandScanner(testLogger, NewShellDetector())
	commands, resolvedShell, err := scanner.ListExecutableCommandsForShell(ShellTypeBash)
	if err != nil {
		t.Fatalf("list commands failed: %v", err)
	}
	if resolvedShell != ShellTypeBash {
		t.Fatalf("expected resolved shell bash, got %s", resolvedShell)
	}

	got := map[string]string{}
	for _, cmd := range commands {
		got[cmd.Name] = cmd.Path
	}

	if got["customcmd"] != filepath.Join(customDir, "customcmd") {
		t.Fatalf("expected customcmd from shell config PATH, got %q", got["customcmd"])
	}
	if got["basecmd"] != filepath.Join(baseDir, "basecmd") {
		t.Fatalf("expected basecmd from process PATH, got %q", got["basecmd"])
	}
}
