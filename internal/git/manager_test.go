package git

import (
	"context"
	"log/slog"
	"testing"
	"time"

	boxtypes "github.com/chenyang-zz/boxify/internal/types"
)

func TestManagerRepoLifecycle(t *testing.T) {
	repoDir := mustInitGitRepo(t)
	eventCh := make(chan boxtypes.GitStatusChangedEvent, 8)

	manager := NewManager(context.Background(), slog.Default(), func(event boxtypes.GitStatusChangedEvent) {
		eventCh <- event
	})
	defer manager.Shutdown()

	info, err := manager.RegisterRepo("repo1", repoDir)
	if err != nil {
		t.Fatalf("register repo failed: %v", err)
	}
	if info.RepoKey != "repo1" {
		t.Fatalf("unexpected repo key: %s", info.RepoKey)
	}

	if _, err := manager.SetActiveRepo("repo1", false, false); err != nil {
		t.Fatalf("set active repo failed: %v", err)
	}
	if manager.ActiveRepoKey() != "repo1" {
		t.Fatalf("unexpected active repo: %s", manager.ActiveRepoKey())
	}

	watchInfo, err := manager.StartWatch("repo1", 300)
	if err != nil {
		t.Fatalf("start watch failed: %v", err)
	}
	if !watchInfo.Watching {
		t.Fatal("watch should be running")
	}
	if watchInfo.IntervalMs < 800 {
		t.Fatalf("watch interval should be clamped, got %d", watchInfo.IntervalMs)
	}

	select {
	case evt := <-eventCh:
		if evt.RepoKey != "repo1" {
			t.Fatalf("unexpected event repo key: %s", evt.RepoKey)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("did not receive startup status event")
	}

	stoppedInfo, err := manager.StopWatch("repo1")
	if err != nil {
		t.Fatalf("stop watch failed: %v", err)
	}
	if stoppedInfo.Watching {
		t.Fatal("watch should be stopped")
	}

	if err := manager.RemoveRepo("repo1"); err != nil {
		t.Fatalf("remove repo failed: %v", err)
	}
	if got := manager.ListRepos(); len(got) != 0 {
		t.Fatalf("expected empty repo list, got %d", len(got))
	}
}
