package process

import "time"

// GetLogs 获取最近日志。
func (m *Manager) GetLogs(n int) []string {
	m.logMu.RLock()
	defer m.logMu.RUnlock()

	if n <= 0 || n > len(m.logLines) {
		n = len(m.logLines)
	}
	start := len(m.logLines) - n
	if start < 0 {
		start = 0
	}
	result := make([]string, n)
	copy(result, m.logLines[start:])
	return result
}

// StreamLogs 将增量日志推送给调用方回调。
func (m *Manager) StreamLogs(writeLine func(line string)) {
	if writeLine == nil {
		return
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	lastIdx := 0
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.logMu.RLock()
			newLines := m.logLines[lastIdx:]
			lastIdx = len(m.logLines)
			m.logMu.RUnlock()
			for _, line := range newLines {
				writeLine(line)
			}
		}
	}
}

// addLogLine 追加日志并执行固定窗口裁剪，防止内存无限增长。
func (m *Manager) addLogLine(line string) {
	m.logMu.Lock()
	defer m.logMu.Unlock()

	m.logLines = append(m.logLines, line)
	if len(m.logLines) > m.maxLog {
		m.logLines = m.logLines[len(m.logLines)-m.maxLog:]
	}
}
