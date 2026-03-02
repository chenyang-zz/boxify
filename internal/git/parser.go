package git

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	boxtypes "github.com/chenyang-zz/boxify/internal/types"
)

// StatusParser 负责解析 git porcelain v2 输出。
type StatusParser struct {
	logger *slog.Logger
}

// NewStatusParser 创建状态解析器。
func NewStatusParser(logger *slog.Logger) *StatusParser {
	if logger == nil {
		logger = slog.Default()
	}
	return &StatusParser{logger: logger}
}

// ParsePorcelainV2 解析 git status --porcelain=v2 --branch 输出。
func (p *StatusParser) ParsePorcelainV2(lines []string) (*boxtypes.GitRepoStatus, error) {
	status := &boxtypes.GitRepoStatus{Files: make([]boxtypes.GitFileStatus, 0, 16)}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "# "):
			p.parseBranchLine(status, strings.TrimPrefix(line, "# "))
		case strings.HasPrefix(line, "1 "):
			file, err := p.parseChangedLine(line)
			if err != nil {
				continue
			}
			status.Files = append(status.Files, file)
			status.StagedCount += boolToInt(file.IndexStatus != ".")
			status.UnstagedCount += boolToInt(file.WorkTreeStatus != ".")
		case strings.HasPrefix(line, "2 "):
			file, err := p.parseRenamedLine(line)
			if err != nil {
				continue
			}
			status.Files = append(status.Files, file)
			status.StagedCount += boolToInt(file.IndexStatus != ".")
			status.UnstagedCount += boolToInt(file.WorkTreeStatus != ".")
		case strings.HasPrefix(line, "u "):
			file, err := p.parseUnmergedLine(line)
			if err != nil {
				continue
			}
			status.Files = append(status.Files, file)
			status.ConflictCount++
		case strings.HasPrefix(line, "? "):
			path := strings.TrimSpace(strings.TrimPrefix(line, "? "))
			status.Files = append(status.Files, boxtypes.GitFileStatus{Path: path, Kind: "untracked", IndexStatus: "?", WorkTreeStatus: "?"})
			status.UntrackedCount++
		}
	}

	if status.Head == "" {
		status.Head = "HEAD"
	}

	return status, nil
}

// parseBranchLine 解析分支元信息行（# branch.*）。
func (p *StatusParser) parseBranchLine(status *boxtypes.GitRepoStatus, line string) {
	switch {
	case strings.HasPrefix(line, "branch.oid "):
		status.Oid = strings.TrimSpace(strings.TrimPrefix(line, "branch.oid "))
	case strings.HasPrefix(line, "branch.head "):
		head := strings.TrimSpace(strings.TrimPrefix(line, "branch.head "))
		if head == "(detached)" {
			status.Detached = true
			status.Head = "HEAD"
			return
		}
		status.Head = head
	case strings.HasPrefix(line, "branch.upstream "):
		status.Upstream = strings.TrimSpace(strings.TrimPrefix(line, "branch.upstream "))
	case strings.HasPrefix(line, "branch.ab "):
		parts := strings.Fields(strings.TrimPrefix(line, "branch.ab "))
		for _, item := range parts {
			if strings.HasPrefix(item, "+") {
				status.Ahead, _ = strconv.Atoi(strings.TrimPrefix(item, "+"))
			} else if strings.HasPrefix(item, "-") {
				status.Behind, _ = strconv.Atoi(strings.TrimPrefix(item, "-"))
			}
		}
	}
}

// parseChangedLine 解析普通变更记录（前缀 1）。
func (p *StatusParser) parseChangedLine(line string) (boxtypes.GitFileStatus, error) {
	parts := strings.Fields(line)
	if len(parts) < 9 {
		return boxtypes.GitFileStatus{}, fmt.Errorf("invalid changed line")
	}
	xy := parts[1]
	if len(xy) < 2 {
		return boxtypes.GitFileStatus{}, fmt.Errorf("invalid changed line")
	}
	return boxtypes.GitFileStatus{
		Path:           parts[len(parts)-1],
		IndexStatus:    string(xy[0]),
		WorkTreeStatus: string(xy[1]),
		Kind:           "changed",
	}, nil
}

// parseRenamedLine 解析重命名/复制记录（前缀 2）。
func (p *StatusParser) parseRenamedLine(line string) (boxtypes.GitFileStatus, error) {
	chunks := strings.Split(line, "\t")
	if len(chunks) < 2 {
		return boxtypes.GitFileStatus{}, fmt.Errorf("invalid renamed line")
	}
	meta := strings.Fields(chunks[0])
	if len(meta) < 10 {
		return boxtypes.GitFileStatus{}, fmt.Errorf("invalid renamed line")
	}
	xy := meta[1]
	if len(xy) < 2 {
		return boxtypes.GitFileStatus{}, fmt.Errorf("invalid renamed line")
	}
	return boxtypes.GitFileStatus{
		Path:           strings.TrimSpace(chunks[1]),
		OriginalPath:   strings.TrimSpace(chunks[len(chunks)-1]),
		IndexStatus:    string(xy[0]),
		WorkTreeStatus: string(xy[1]),
		Kind:           "renamed",
	}, nil
}

// parseUnmergedLine 解析冲突记录（前缀 u）。
func (p *StatusParser) parseUnmergedLine(line string) (boxtypes.GitFileStatus, error) {
	parts := strings.Fields(line)
	if len(parts) < 11 {
		return boxtypes.GitFileStatus{}, fmt.Errorf("invalid unmerged line")
	}
	xy := parts[1]
	if len(xy) < 2 {
		return boxtypes.GitFileStatus{}, fmt.Errorf("invalid unmerged line")
	}
	return boxtypes.GitFileStatus{
		Path:           parts[len(parts)-1],
		IndexStatus:    string(xy[0]),
		WorkTreeStatus: string(xy[1]),
		Kind:           "unmerged",
	}, nil
}

// boolToInt 将布尔值转换为 0/1 计数值。
func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
