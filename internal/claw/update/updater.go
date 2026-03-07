package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	// UpdateCheckURL 是更新信息地址。
	UpdateCheckURL  = "http://39.102.53.188:16198/clawpanel/update.json"
	httpTimeout     = 30 * time.Second
	downloadTimeout = 300 * time.Second
)

// UpdateInfo 描述服务端 update.json 信息。
type UpdateInfo struct {
	LatestVersion string            `json:"latest_version"` // 服务端发布的最新版本号。
	ReleaseTime   string            `json:"release_time"`   // 版本发布时间（由服务端透传）。
	ReleaseNote   string            `json:"release_note"`   // 版本更新说明。
	DownloadURLs  map[string]string `json:"download_urls"`  // 各平台安装包下载地址，key 为平台标识。
	SHA256        map[string]string `json:"sha256"`         // 各平台安装包 SHA256 校验值，key 为平台标识。
}

// UpdatePopup 描述更新完成后的弹窗提示。
type UpdatePopup struct {
	Show        bool   `json:"show"`              // 是否展示更新完成弹窗。
	Version     string `json:"version"`           // 已安装完成的版本号。
	ReleaseNote string `json:"release_note"`      // 本次安装版本的更新说明。
	ShownAt     string `json:"shown_at,omitempty"` // 弹窗确认时间（RFC3339），未确认时为空。
}

// UpdateProgress 描述更新任务当前进度。
type UpdateProgress struct {
	Status     string   `json:"status"`              // 当前阶段：idle/checking/downloading/verifying/replacing/restarting/done/error。
	Progress   int      `json:"progress"`            // 进度百分比（0-100）。
	Message    string   `json:"message"`             // 当前阶段的人类可读提示。
	Log        []string `json:"log"`                 // 更新任务日志快照。
	Error      string   `json:"error,omitempty"`     // 失败时的错误信息。
	StartedAt  string   `json:"started_at,omitempty"` // 任务开始时间（RFC3339）。
	FinishedAt string   `json:"finished_at,omitempty"` // 任务结束时间（RFC3339）。
}

// Updater 负责应用自更新流程。
type Updater struct {
	currentVersion string       // 当前运行版本号。
	dataDir        string       // 更新缓存与弹窗信息存储目录。
	logger         *slog.Logger // 更新流程日志器。
	mu             sync.Mutex   // 更新状态互斥锁。
	progress       UpdateProgress
}

// NewUpdater 创建更新器实例。
func NewUpdater(currentVersion, dataDir string, logger *slog.Logger) *Updater {
	return NewUpdaterWithLogger(currentVersion, dataDir, logger)
}

// NewUpdaterWithLogger 创建带日志器的更新器实例。
func NewUpdaterWithLogger(currentVersion, dataDir string, logger *slog.Logger) *Updater {
	if logger == nil {
		logger = slog.Default()
	}

	logger = logger.With("module", "claw.update")

	return &Updater{
		currentVersion: currentVersion,
		dataDir:        dataDir,
		logger:         logger,
		progress: UpdateProgress{
			Status: "idle",
			Log:    []string{},
		},
	}
}

// getPlatformKey 返回当前平台下载键。
func getPlatformKey() string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	switch {
	case os == "linux" && arch == "amd64":
		return "linux_amd64"
	case os == "linux" && arch == "arm64":
		return "linux_arm64"
	case os == "windows" && arch == "amd64":
		return "windows_amd64"
	case os == "darwin" && arch == "amd64":
		return "darwin_amd64"
	case os == "darwin" && arch == "arm64":
		return "darwin_arm64"
	default:
		return os + "_" + arch
	}
}

// CheckUpdate 检查可用更新。
func (u *Updater) CheckUpdate() (*UpdateInfo, bool, error) {
	u.logger.Info("开始检查更新", "current_version", u.currentVersion, "url", UpdateCheckURL)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(UpdateCheckURL)
	if err != nil {
		u.logger.Error("请求更新服务器失败", "error", err)
		return nil, false, fmt.Errorf("请求更新服务器失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		u.logger.Warn("更新服务器返回非预期状态码", "status", resp.StatusCode)
		return nil, false, fmt.Errorf("更新服务器返回错误: HTTP %d", resp.StatusCode)
	}

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		u.logger.Error("解析更新信息失败", "error", err)
		return nil, false, fmt.Errorf("解析更新信息失败: %v", err)
	}

	hasUpdate := info.LatestVersion != "" && info.LatestVersion != u.currentVersion && isNewerVersion(info.LatestVersion, u.currentVersion)
	u.logger.Info("更新检查完成", "latest_version", info.LatestVersion, "has_update", hasUpdate)

	return &info, hasUpdate, nil
}

// GetProgress 返回更新进度快照。
func (u *Updater) GetProgress() UpdateProgress {
	u.mu.Lock()
	defer u.mu.Unlock()
	p := u.progress
	logCopy := make([]string, len(u.progress.Log))
	copy(logCopy, u.progress.Log)
	p.Log = logCopy
	return p
}

// DoUpdate 启动异步自更新流程。
func (u *Updater) DoUpdate(info *UpdateInfo) {
	if info == nil {
		u.setError("更新信息不能为空")
		return
	}

	u.mu.Lock()
	currentStatus := u.progress.Status
	if currentStatus == "downloading" || currentStatus == "verifying" || currentStatus == "replacing" {
		u.mu.Unlock()
		u.logger.Warn("检测到已有更新任务执行中，忽略重复请求", "status", currentStatus)
		return
	}
	u.progress = UpdateProgress{
		Status:    "downloading",
		Progress:  0,
		Message:   "准备下载更新...",
		Log:       []string{},
		StartedAt: time.Now().Format(time.RFC3339),
	}
	u.mu.Unlock()
	u.logger.Info("更新任务已启动", "target_version", info.LatestVersion)

	go u.doUpdateAsync(info)
}

// doUpdateAsync 执行异步更新主流程（下载、校验、替换、重启）。
func (u *Updater) doUpdateAsync(info *UpdateInfo) {
	u.log("检测平台: %s/%s", runtime.GOOS, runtime.GOARCH)

	platformKey := getPlatformKey()
	downloadURL, ok := info.DownloadURLs[platformKey]
	if !ok {
		u.setError("不支持的平台: %s", platformKey)
		return
	}
	expectedSHA, _ := info.SHA256[platformKey]

	u.log("开始下载更新: %s -> %s", info.LatestVersion, downloadURL)
	u.setStatus("downloading", 10, "正在下载更新包...")

	// 下载到临时文件
	tmpDir := filepath.Join(u.dataDir, "update-tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		u.setError("创建更新临时目录失败: %v", err)
		return
	}
	tmpFile := filepath.Join(tmpDir, "clawpanel-new")
	if runtime.GOOS == "windows" {
		tmpFile += ".exe"
	}

	if err := u.downloadFile(downloadURL, tmpFile); err != nil {
		u.setError("下载失败: %v", err)
		return
	}
	u.log("下载完成")

	// SHA256 校验
	u.setStatus("verifying", 60, "正在校验文件完整性...")
	if expectedSHA != "" {
		actualSHA, err := fileSHA256(tmpFile)
		if err != nil {
			u.setError("校验失败: %v", err)
			return
		}
		if !strings.EqualFold(actualSHA, expectedSHA) {
			u.setError("SHA256 校验失败: 期望 %s, 实际 %s\n更新包可能损坏，请重新尝试", shortHash(expectedSHA), shortHash(actualSHA))
			os.Remove(tmpFile)
			return
		}
		u.log("SHA256 校验通过")
	} else {
		u.logger.Warn("未提供 SHA256 校验值，跳过校验", "platform", platformKey)
		u.log("未提供 SHA256 校验值，已跳过校验")
	}

	// 替换二进制
	u.setStatus("replacing", 80, "正在替换程序...")
	currentBin, err := os.Executable()
	if err != nil {
		u.setError("获取当前程序路径失败: %v", err)
		return
	}
	currentBin, _ = filepath.EvalSymlinks(currentBin)
	u.log("当前程序路径: %s", currentBin)

	// 替换前先保存弹窗信息，避免重启中断后丢失
	u.saveUpdatePopup(info)
	u.log("更新信息已保存")

	// Linux/macOS 下生成外部脚本执行停服、替换与拉起，避免进程自覆盖竞态。
	if runtime.GOOS != "windows" {
		scriptPath := filepath.Join(tmpDir, "do-update.sh")
		script := fmt.Sprintf(`#!/bin/bash
set -e
echo "[ClawPanel Updater] 开始更新..."

# Stop service
echo "[ClawPanel Updater] 停止 ClawPanel 服务..."
systemctl stop clawpanel 2>/dev/null || true
sleep 1

# Wait for process to exit (up to 10s)
for i in $(seq 1 10); do
  if ! pgrep -x clawpanel >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

# Kill if still running
if pgrep -x clawpanel >/dev/null 2>&1; then
  echo "[ClawPanel Updater] 强制停止旧进程..."
  pkill -9 -x clawpanel 2>/dev/null || true
  sleep 1
fi

# Backup old binary
if [ -f "%s" ]; then
  cp -f "%s" "%s.bak" 2>/dev/null || true
  echo "[ClawPanel Updater] 已备份旧程序"
fi

# Replace: remove old then copy new
rm -f "%s"
cp -f "%s" "%s"
chmod +x "%s"
echo "[ClawPanel Updater] 程序替换完成"

# Start service
echo "[ClawPanel Updater] 启动 ClawPanel 服务..."
systemctl start clawpanel 2>/dev/null || ( echo "[ClawPanel Updater] systemctl 启动失败，尝试直接启动..." && nohup "%s" >/dev/null 2>&1 & )
echo "[ClawPanel Updater] 更新完成!"

# Clean up
rm -f "%s"
rm -rf "%s"
`, currentBin, currentBin, currentBin, currentBin, tmpFile, currentBin, currentBin, currentBin, scriptPath, tmpDir)

		if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
			u.setError("写入更新脚本失败: %v", err)
			return
		}
		u.log("更新脚本已生成: %s", scriptPath)

		u.setStatus("restarting", 95, "即将停止服务并替换程序...")
		u.log("即将停止服务并由更新脚本接管")

		u.mu.Lock()
		u.progress.Status = "done"
		u.progress.Progress = 100
		u.progress.Message = "更新完成，正在重启..."
		u.progress.FinishedAt = time.Now().Format(time.RFC3339)
		u.mu.Unlock()

		// 异步拉起外部脚本（脱离当前进程）
		go func() {
			time.Sleep(1 * time.Second)
			cmd := exec.Command("bash", "-c", "setsid bash "+scriptPath+" </dev/null >/dev/null 2>&1 &")
			cmd.Stdout = nil
			cmd.Stderr = nil
			if err := cmd.Start(); err != nil {
				u.logger.Warn("启动更新脚本失败，尝试直接替换", "error", err)
				// 兜底：直接替换并重启服务
				os.Remove(currentBin)
				copyFile(tmpFile, currentBin)
				os.Chmod(currentBin, 0755)
				execCmd("systemctl", "restart", "clawpanel")
				return
			}
			u.logger.Info("更新脚本已启动，等待接管", "pid", cmd.Process.Pid)
			// 释放子进程句柄，避免僵尸进程
			cmd.Process.Release()
		}()
		return
	}

	// Windows 使用重命名备份后覆盖
	backupPath := currentBin + ".bak"
	os.Remove(backupPath)
	if err := os.Rename(currentBin, backupPath); err != nil {
		u.setError("备份旧程序失败: %v", err)
		return
	}
	u.log("已备份旧程序: %s", backupPath)

	if err := copyFile(tmpFile, currentBin); err != nil {
		os.Rename(backupPath, currentBin)
		u.setError("替换程序失败: %v", err)
		return
	}
	os.Chmod(currentBin, 0755)
	u.log("程序替换完成")

	os.Remove(tmpFile)
	os.RemoveAll(tmpDir)

	u.setStatus("restarting", 95, "即将重启 ClawPanel...")
	u.log("ClawPanel 即将重启，请等待")

	u.mu.Lock()
	u.progress.Status = "done"
	u.progress.Progress = 100
	u.progress.Message = "更新完成，正在重启..."
	u.progress.FinishedAt = time.Now().Format(time.RFC3339)
	u.mu.Unlock()

	go func() {
		time.Sleep(1 * time.Second)
		if err := execCmd("net", "stop", "ClawPanel"); err == nil {
			execCmd("net", "start", "ClawPanel")
			return
		}
		u.logger.Warn("Windows 服务重启失败，准备退出当前进程")
		os.Exit(0)
	}()
}

// downloadFile 下载更新包并同步进度。
func (u *Updater) downloadFile(url, dest string) error {
	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	totalSize := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			downloaded += int64(n)
			if totalSize > 0 {
				pct := int(float64(downloaded)/float64(totalSize)*50) + 10 // 10-60%
				u.setStatus("downloading", pct, fmt.Sprintf("正在下载... %.1f MB / %.1f MB", float64(downloaded)/1048576, float64(totalSize)/1048576))
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// saveUpdatePopup 持久化更新完成弹窗数据。
func (u *Updater) saveUpdatePopup(info *UpdateInfo) {
	popup := UpdatePopup{
		Show:        true,
		Version:     info.LatestVersion,
		ReleaseNote: info.ReleaseNote,
	}
	data, err := json.MarshalIndent(popup, "", "  ")
	if err != nil {
		u.logger.Warn("序列化更新弹窗信息失败", "error", err)
		return
	}

	popupPath := filepath.Join(u.dataDir, "update_popup.json")
	if err := os.WriteFile(popupPath, data, 0o644); err != nil {
		u.logger.Warn("写入更新弹窗信息失败", "path", popupPath, "error", err)
	}
}

// GetUpdatePopup 读取更新弹窗信息。
func (u *Updater) GetUpdatePopup() *UpdatePopup {
	data, err := os.ReadFile(filepath.Join(u.dataDir, "update_popup.json"))
	if err != nil {
		return nil
	}
	var popup UpdatePopup
	if err := json.Unmarshal(data, &popup); err != nil {
		return nil
	}
	return &popup
}

// MarkPopupShown 标记弹窗已展示。
func (u *Updater) MarkPopupShown() {
	popup := u.GetUpdatePopup()
	if popup == nil {
		return
	}
	popup.Show = false
	popup.ShownAt = time.Now().Format(time.RFC3339)
	data, err := json.MarshalIndent(popup, "", "  ")
	if err != nil {
		u.logger.Warn("序列化弹窗展示状态失败", "error", err)
		return
	}

	popupPath := filepath.Join(u.dataDir, "update_popup.json")
	if err := os.WriteFile(popupPath, data, 0o644); err != nil {
		u.logger.Warn("写入弹窗展示状态失败", "path", popupPath, "error", err)
	}
}

// log 记录更新过程日志并写入进度日志列表。
func (u *Updater) log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	u.logger.Info(msg)
	u.mu.Lock()
	u.progress.Log = append(u.progress.Log, msg)
	u.mu.Unlock()
}

// setStatus 更新当前进度状态。
func (u *Updater) setStatus(status string, progress int, message string) {
	u.mu.Lock()
	u.progress.Status = status
	u.progress.Progress = progress
	u.progress.Message = message
	u.mu.Unlock()
}

// setError 写入失败状态并记录错误日志。
func (u *Updater) setError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	u.logger.Error("更新失败", "error", msg)
	u.mu.Lock()
	u.progress.Status = "error"
	u.progress.Error = msg
	u.progress.Message = "更新失败"
	u.progress.Log = append(u.progress.Log, "错误: "+msg)
	u.progress.FinishedAt = time.Now().Format(time.RFC3339)
	u.mu.Unlock()
}

// isNewerVersion 判断 latest 是否高于 current。
func isNewerVersion(latest, current string) bool {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")
	lp := strings.Split(latest, ".")
	cp := strings.Split(current, ".")
	for i := 0; i < len(lp) && i < len(cp); i++ {
		lv := 0
		cv := 0
		fmt.Sscanf(lp[i], "%d", &lv)
		fmt.Sscanf(cp[i], "%d", &cv)
		if lv > cv {
			return true
		}
		if lv < cv {
			return false
		}
	}
	return len(lp) > len(cp)
}

// fileSHA256 计算文件的 SHA256 摘要。
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// shortHash 返回用于日志展示的摘要前缀，避免输出过长。
func shortHash(value string) string {
	if len(value) <= 16 {
		return value
	}
	return value[:16] + "..."
}

// copyFile 覆盖复制文件内容。
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// execCmd 执行外部命令并等待结束。
func execCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
