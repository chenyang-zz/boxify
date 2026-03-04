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

func TestListExecutableCommandsUnix(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	mustWriteFileWithMode(t, filepath.Join(dir1, "git"), 0o755)
	mustWriteFileWithMode(t, filepath.Join(dir1, "README.md"), 0o644)
	mustWriteFileWithMode(t, filepath.Join(dir2, "git"), 0o755)
	mustWriteFileWithMode(t, filepath.Join(dir2, "node"), 0o755)

	commands := listExecutableCommands(dir1+string(os.PathListSeparator)+dir2, "", "darwin", testLogger)

	if len(commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(commands))
	}

	got := map[string]string{}
	for _, cmd := range commands {
		got[cmd.Name] = cmd.Path
	}

	if got["git"] != filepath.Join(dir1, "git") {
		t.Fatalf("expected git from first PATH directory, got %q", got["git"])
	}
	if got["node"] != filepath.Join(dir2, "node") {
		t.Fatalf("expected node in second PATH directory, got %q", got["node"])
	}
	if _, ok := got["README.md"]; ok {
		t.Fatal("non-executable file should not be included")
	}
}

func TestListExecutableCommandsWindows(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	mustWriteFileWithMode(t, filepath.Join(dir1, "git.cmd"), 0o644)
	mustWriteFileWithMode(t, filepath.Join(dir1, "go.exe"), 0o644)
	mustWriteFileWithMode(t, filepath.Join(dir1, "README.md"), 0o644)
	mustWriteFileWithMode(t, filepath.Join(dir2, "git.exe"), 0o644)
	mustWriteFileWithMode(t, filepath.Join(dir2, "pnpm.bat"), 0o644)

	commands := listExecutableCommands(dir1+string(os.PathListSeparator)+dir2, ".EXE;.CMD;.BAT", "windows", testLogger)

	if len(commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(commands))
	}

	got := map[string]string{}
	for _, cmd := range commands {
		got[cmd.Name] = cmd.Path
	}

	if got["git"] != filepath.Join(dir1, "git.cmd") {
		t.Fatalf("expected git from first PATH directory, got %q", got["git"])
	}
	if got["go"] != filepath.Join(dir1, "go.exe") {
		t.Fatalf("expected go.exe from first PATH directory, got %q", got["go"])
	}
	if got["pnpm"] != filepath.Join(dir2, "pnpm.bat") {
		t.Fatalf("expected pnpm.bat in second PATH directory, got %q", got["pnpm"])
	}
	if _, ok := got["README"]; ok {
		t.Fatal("README.md should not be included on windows")
	}
}

func TestResolveShellType(t *testing.T) {
	scanner := NewPathCommandScanner(testLogger, NewShellDetector())

	resolved, err := scanner.ResolveShellType(ShellTypeAuto)
	if err != nil {
		t.Fatalf("resolve auto shell failed: %v", err)
	}
	if resolved == ShellTypeAuto {
		t.Fatal("auto should be resolved to concrete shell type")
	}

	resolved, err = scanner.ResolveShellType(ShellTypeBash)
	if err != nil {
		t.Fatalf("resolve bash shell failed: %v", err)
	}
	if resolved != ShellTypeBash {
		t.Fatalf("expected bash, got %s", resolved)
	}

	_, err = scanner.ResolveShellType(ShellType("fish"))
	if err == nil {
		t.Fatal("expected error for unsupported shell type")
	}
}

func TestGetDefaultCommands(t *testing.T) {
	scanner := NewPathCommandScanner(testLogger, NewShellDetector())

	unixDefaults := scanner.GetDefaultCommands(ShellTypeBash)
	if len(unixDefaults) == 0 {
		t.Fatal("bash defaults should not be empty")
	}

	cmdDefaults := scanner.GetDefaultCommands(ShellTypeCmd)
	if len(cmdDefaults) == 0 {
		t.Fatal("cmd defaults should not be empty")
	}

	autoDefaults := scanner.GetDefaultCommands(ShellTypeAuto)
	if len(autoDefaults) != 0 {
		t.Fatal("auto defaults should be empty before resolving")
	}
}

func TestResolveShellTypeWithEmptyValue(t *testing.T) {
	scanner := NewPathCommandScanner(testLogger, NewShellDetector())
	resolved, err := scanner.ResolveShellType("")
	if err != nil {
		t.Fatalf("resolve empty shell failed: %v", err)
	}

	if resolved == "" || resolved == ShellTypeAuto {
		t.Fatalf("resolved shell should be concrete, got %s", resolved)
	}

	if runtime.GOOS == "windows" {
		if resolved != ShellTypeCmd && resolved != ShellTypePowershell && resolved != ShellTypePwsh {
			t.Fatalf("unexpected resolved shell on windows: %s", resolved)
		}
	}
}

func mustWriteFileWithMode(t *testing.T, path string, mode os.FileMode) {
	t.Helper()

	if err := os.WriteFile(path, []byte("echo"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("chmod file failed: %v", err)
	}
}
