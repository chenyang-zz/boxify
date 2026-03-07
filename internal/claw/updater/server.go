package updater

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	UpdaterPort        = 19528                                                                // 独立更新服务监听端口
	TokenValidDuration = 5 * time.Minute                                                      // 更新页访问令牌有效期
	TokenSecret        = "clawpanel-updater-secret-2026"                                      // 访问令牌签名密钥
	GitHubReleaseAPI   = "https://api.github.com/repos/zhaoxinyi02/ClawPanel/releases/latest" // GitHub 最新版本接口
	AccelUpdateJSON    = "http://39.102.53.188:16198/clawpanel/update.json"                   // 加速线路版本信息接口
	GitHubDownloadBase = "https://github.com/zhaoxinyi02/ClawPanel/releases/download"         // GitHub 二进制下载前缀
	AccelDownloadBase  = "http://39.102.53.188:16198/clawpanel/releases"                      // 加速线路下载前缀
)

// UpdateStep 描述更新流程中的单个步骤状态。
type UpdateStep struct {
	Name    string `json:"name"`              // 步骤名称
	Status  string `json:"status"`            // 步骤状态：pending/running/done/error/skipped
	Message string `json:"message,omitempty"` // 步骤补充说明
}

// UpdateState 描述完整的更新状态机快照。
type UpdateState struct {
	Phase      string       `json:"phase"`                 // 流程阶段：idle/validating/checking/stopping/downloading/backing_up/replacing/starting/done/error/rolled_back
	Steps      []UpdateStep `json:"steps"`                 // 分步状态集合
	Progress   int          `json:"progress"`              // 整体进度百分比（0-100）
	Message    string       `json:"message"`               // 当前阶段提示文案
	Log        []string     `json:"log"`                   // 更新日志流
	Error      string       `json:"error,omitempty"`       // 错误详情
	StartedAt  string       `json:"started_at,omitempty"`  // 流程开始时间（RFC3339）
	FinishedAt string       `json:"finished_at,omitempty"` // 流程结束时间（RFC3339）
	Source     string       `json:"source,omitempty"`      // 更新来源：github/accel/upload
	FromVer    string       `json:"from_ver,omitempty"`    // 更新前版本
	ToVer      string       `json:"to_ver,omitempty"`      // 更新目标版本
}

// VersionInfo 描述远端版本信息。
type VersionInfo struct {
	LatestVersion string            `json:"latest_version"`           // 最新版本号
	ReleaseTime   string            `json:"release_time"`             // 发布时间（通常为 RFC3339）
	ReleaseNote   string            `json:"release_note"`             // 发布说明
	DownloadURLs  map[string]string `json:"download_urls"`            // 各平台下载地址（key: 平台标识）
	SHA256        map[string]string `json:"sha256"`                   // 各平台包摘要（key: 平台标识）
	MajorChange   bool              `json:"major_change,omitempty"`   // 是否为重大变更
	ChangeWarning string            `json:"change_warning,omitempty"` // 重大变更提示文案
}

// Server 提供独立更新服务的 HTTP 实现。
type Server struct {
	currentVersion string       // 当前 ClawPanel 版本号
	dataDir        string       // 更新服务运行所需的数据目录
	openClawDir    string       // OpenClaw 配置目录
	panelBin       string       // ClawPanel 二进制路径
	panelPort      int          // ClawPanel HTTP 服务端口
	logger         *slog.Logger // 更新服务日志器
	mu             sync.Mutex   // 更新状态并发保护锁
	state          UpdateState  // ClawPanel 更新状态
	ocState        UpdateState  // OpenClaw 更新状态
	srv            *http.Server // 独立更新 HTTP 服务实例
	running        bool         // 独立更新服务是否已拉起
}

// NewServer 创建独立更新服务实例。
func NewServer(currentVersion, dataDir, openClawDir string, panelPort int, logger *slog.Logger) *Server {
	return NewServerWithLogger(currentVersion, dataDir, openClawDir, panelPort, logger)
}

// NewServerWithLogger 创建独立更新服务实例，并注入日志器。
func NewServerWithLogger(currentVersion, dataDir, openClawDir string, panelPort int, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("module", "claw.updater")

	bin, _ := os.Executable()
	bin, _ = filepath.EvalSymlinks(bin)
	return &Server{
		currentVersion: currentVersion,
		dataDir:        dataDir,
		openClawDir:    openClawDir,
		panelBin:       bin,
		panelPort:      panelPort,
		logger:         logger,
		state: UpdateState{
			Phase: "idle",
			Steps: defaultSteps(),
			Log:   []string{},
		},
		ocState: UpdateState{
			Phase: "idle",
			Steps: defaultOCSteps(),
			Log:   []string{},
		},
	}
}

// defaultSteps 返回 ClawPanel 更新流程默认步骤。
func defaultSteps() []UpdateStep {
	return []UpdateStep{
		{Name: "验证授权", Status: "pending"},
		{Name: "检测版本", Status: "pending"},
		{Name: "停止服务", Status: "pending"},
		{Name: "下载更新", Status: "pending"},
		{Name: "备份文件", Status: "pending"},
		{Name: "替换文件", Status: "pending"},
		{Name: "启动服务", Status: "pending"},
	}
}

// defaultOCSteps 返回 OpenClaw 更新流程默认步骤。
func defaultOCSteps() []UpdateStep {
	return []UpdateStep{
		{Name: "验证授权", Status: "pending"},
		{Name: "检测版本", Status: "pending"},
		{Name: "执行更新", Status: "pending"},
		{Name: "重启服务", Status: "pending"},
	}
}

// GenerateToken 生成更新页临时访问令牌。
func GenerateToken(panelPort int) string {
	now := time.Now().Unix()
	payload := fmt.Sprintf("%d:%d", panelPort, now)
	mac := hmac.New(sha256.New, []byte(TokenSecret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%d", sig, now)
}

// ValidateToken 校验访问令牌合法性与时效。
func ValidateToken(token string, panelPort int) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	sig := parts[0]
	tsStr := parts[1]
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return false
	}
	// 校验过期时间
	if time.Since(time.Unix(ts, 0)) > TokenValidDuration {
		return false
	}
	// 校验签名
	payload := fmt.Sprintf("%d:%d", panelPort, ts)
	mac := hmac.New(sha256.New, []byte(TokenSecret))
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

// Start 以独立子进程方式启动更新服务。
// 在 systemd 环境下优先使用 systemd-run --scope 脱离父 cgroup，避免停服务时被连带终止。
func (s *Server) Start() {
	// 清理上次遗留的独立更新进程
	s.killStandaloneUpdater()
	time.Sleep(500 * time.Millisecond)

	bin := s.panelBin
	logFile := filepath.Join(s.dataDir, "updater.log")

	if runtime.GOOS != "windows" {
		// 使用 systemd-run --scope 脱离父进程 cgroup，并使用时间戳 unit 避免重名冲突。
		unitName := fmt.Sprintf("clawpanel-updater-%d", time.Now().Unix())
		cmd := exec.Command("systemd-run", "--scope", "--unit="+unitName,
			"/bin/bash", "-c",
			fmt.Sprintf("%s --updater-standalone %s %s %d %s >%s 2>&1",
				bin, s.currentVersion, s.dataDir, s.panelPort, s.openClawDir, logFile),
		)
		cmd.SysProcAttr = sysProcAttr()
		cmd.Dir = filepath.Dir(bin)
		if err := cmd.Start(); err != nil {
			if isLikelySystemdServiceProcess() {
				s.logger.Warn(fmt.Sprintf("systemd-run 启动失败: %v，当前运行在 systemd service 中，已拒绝 direct 模式以避免停服后更新器被连带终止", err))
				return
			}
			s.logger.Warn(fmt.Sprintf("systemd-run 启动失败: %v，尝试 direct 模式启动", err))
			s.startDirectChild(bin, logFile)
			return
		}
		// 短暂等待并检查 systemd-run 是否快速失败（例如 unit 名冲突）
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			// systemd-run 立即退出，表示内部命令未正常拉起
			if isLikelySystemdServiceProcess() {
				s.logger.Warn(fmt.Sprintf("systemd-run 立即退出 (err=%v)，当前运行在 systemd service 中，已拒绝 direct 模式以避免停服后更新器被连带终止", err))
				return
			}
			s.logger.Warn(fmt.Sprintf("systemd-run 立即退出 (err=%v)，尝试 direct 模式启动", err))
			s.startDirectChild(bin, logFile)
			return
		case <-time.After(800 * time.Millisecond):
			// 800ms 后仍在运行，视为成功拉起子进程
			cmd.Process.Release()
		}
	} else {
		s.startDirectChild(bin, logFile)
		return
	}

	s.running = true
	s.logger.Info(fmt.Sprintf("独立更新子进程已启动 (systemd-run scope) → http://0.0.0.0:%d/updater", UpdaterPort))
}

// isLikelySystemdServiceProcess 判断当前进程是否运行在 systemd service 环境。
func isLikelySystemdServiceProcess() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	if os.Getenv("INVOCATION_ID") != "" {
		return true
	}
	if data, err := os.ReadFile("/proc/self/cgroup"); err == nil {
		text := string(data)
		if strings.Contains(text, ".service") || strings.Contains(text, "system.slice") {
			return true
		}
	}
	return false
}

// startDirectChild 直接以脱离子进程方式启动更新服务（非 systemd 或 Windows 回退路径）。
func (s *Server) startDirectChild(bin, logFile string) {
	cmd := exec.Command(bin,
		"--updater-standalone",
		s.currentVersion,
		s.dataDir,
		fmt.Sprintf("%d", s.panelPort),
		s.openClawDir,
	)
	cmd.SysProcAttr = sysProcAttr()
	cmd.Dir = filepath.Dir(bin)
	lf, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		cmd.Stdout = lf
		cmd.Stderr = lf
	}
	if err := cmd.Start(); err != nil {
		s.logger.Error(fmt.Sprintf("启动独立更新子进程失败: %v", err))
		return
	}
	cmd.Process.Release()
	s.running = true
	s.logger.Info(fmt.Sprintf("独立更新子进程已启动 (direct) → http://0.0.0.0:%d/updater", UpdaterPort))
}

// Stop 停止独立更新子进程。
func (s *Server) Stop() {
	s.killStandaloneUpdater()
}

// killStandaloneUpdater 终止当前机器上的独立更新进程。
func (s *Server) killStandaloneUpdater() {
	if runtime.GOOS == "windows" {
		return
	}
	// 通过 shell 包装避免 --updater-standalone 被当作 pgrep 参数解析
	out, _ := exec.Command("sh", "-c", "pgrep -f 'updater-standalone'").Output()
	pids := strings.Fields(strings.TrimSpace(string(out)))
	myPid := fmt.Sprintf("%d", os.Getpid())
	for _, pid := range pids {
		if pid == myPid {
			continue
		}
		exec.Command("kill", pid).Run()
	}
}

// IsRunning 返回更新服务是否运行中。
func (s *Server) IsRunning() bool {
	return s.running
}

// RunStandalone 以独立模式运行更新服务（由 --updater-standalone 触发）。
// 该方法为阻塞调用，仅在服务退出时返回。
func (s *Server) RunStandalone() {
	mux := http.NewServeMux()
	mux.HandleFunc("/updater", s.handlePage)
	mux.HandleFunc("/updater/", s.handlePage)
	mux.HandleFunc("/updater/api/validate", s.handleValidate)
	mux.HandleFunc("/updater/api/check-version", s.handleCheckVersion)
	mux.HandleFunc("/updater/api/start-update", s.handleStartUpdate)
	mux.HandleFunc("/updater/api/upload-update", s.handleUploadUpdate)
	mux.HandleFunc("/updater/api/progress", s.handleProgress)
	// OpenClaw 更新接口
	mux.HandleFunc("/updater/api/check-openclaw-version", s.handleCheckOCVersion)
	mux.HandleFunc("/updater/api/start-openclaw-update", s.handleStartOCUpdate)
	mux.HandleFunc("/updater/api/openclaw-progress", s.handleOCProgress)

	s.srv = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", UpdaterPort),
		Handler: mux,
	}
	s.running = true

	// 自动退出：空闲或更新完成后达到阈值则关闭服务
	go func() {
		for {
			time.Sleep(10 * time.Second)
			s.mu.Lock()
			phase := s.state.Phase
			finishedAt := s.state.FinishedAt
			ocPhase := s.ocState.Phase
			ocFinishedAt := s.ocState.FinishedAt
			s.mu.Unlock()

			// 任一更新仍在执行时不退出
			panelDone := phase == "idle" || phase == "done" || phase == "error" || phase == "rolled_back"
			ocDone := ocPhase == "idle" || ocPhase == "done" || ocPhase == "error" || ocPhase == "rolled_back"
			if !panelDone || !ocDone {
				continue
			}

			// 至少有一次更新完成且超过阈值后退出
			latestFinish := finishedAt
			if ocFinishedAt > latestFinish {
				latestFinish = ocFinishedAt
			}
			if latestFinish != "" {
				if t, err := time.Parse(time.RFC3339, latestFinish); err == nil {
					if time.Since(t) > 5*time.Minute {
						s.logger.Info("更新完成超过5分钟，自动退出")
						s.srv.Close()
						return
					}
				}
			}
		}
	}()

	s.logger.Info(fmt.Sprintf("独立更新服务已启动 → http://0.0.0.0:%d/updater", UpdaterPort))
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error(fmt.Sprintf("独立更新服务启动失败: %v", err))
	}
	s.logger.Info("独立更新服务已退出")
}

// 以下为 HTTP 处理器。

// handlePage 返回更新页面 HTML，并校验访问令牌。
func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	// 仅允许携带有效 token 的请求访问页面
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "⛔ 禁止直接访问更新页面。请从 ClawPanel 面板的「版本管理」页面点击「前往更新」进入。", http.StatusForbidden)
		return
	}
	if !ValidateToken(token, s.panelPort) {
		http.Error(w, "⛔ 授权令牌已失效或无效。请返回 ClawPanel 面板重新点击「前往更新」。", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(updaterHTML(s.currentVersion, token, s.panelPort)))
}

// handleValidate 校验令牌有效性，供前端初始化流程调用。
func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	s.setCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	token := r.URL.Query().Get("token")
	valid := ValidateToken(token, s.panelPort)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    valid,
		"error": ternary(!valid, "授权令牌无效或已过期", ""),
	})
}

// handleCheckVersion 获取 ClawPanel 最新版本信息。
func (s *Server) handleCheckVersion(w http.ResponseWriter, r *http.Request) {
	s.setCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if !s.checkToken(w, r) {
		return
	}

	info, source, err := s.fetchLatestVersion()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false, "error": err.Error(),
		})
		return
	}

	hasUpdate := info.LatestVersion != "" && info.LatestVersion != s.currentVersion &&
		isNewerVersion(info.LatestVersion, s.currentVersion)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":             true,
		"currentVersion": s.currentVersion,
		"latestVersion":  info.LatestVersion,
		"releaseTime":    info.ReleaseTime,
		"releaseNote":    info.ReleaseNote,
		"hasUpdate":      hasUpdate,
		"source":         source,
		"majorChange":    info.MajorChange,
		"changeWarning":  info.ChangeWarning,
	})
}

// handleStartUpdate 触发 ClawPanel 在线更新流程。
func (s *Server) handleStartUpdate(w http.ResponseWriter, r *http.Request) {
	s.setCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.checkToken(w, r) {
		return
	}

	s.mu.Lock()
	if s.state.Phase != "idle" && s.state.Phase != "done" && s.state.Phase != "error" && s.state.Phase != "rolled_back" {
		s.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false, "error": "更新正在进行中",
		})
		return
	}
	s.state = UpdateState{
		Phase:     "validating",
		Steps:     defaultSteps(),
		Log:       []string{},
		StartedAt: time.Now().Format(time.RFC3339),
	}
	s.mu.Unlock()

	go s.doUpdate("")

	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// handleUploadUpdate 接收离线更新文件并触发替换流程。
func (s *Server) handleUploadUpdate(w http.ResponseWriter, r *http.Request) {
	s.setCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.checkToken(w, r) {
		return
	}

	s.mu.Lock()
	if s.state.Phase != "idle" && s.state.Phase != "done" && s.state.Phase != "error" && s.state.Phase != "rolled_back" {
		s.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false, "error": "更新正在进行中",
		})
		return
	}
	s.mu.Unlock()

	// 解析上传表单（最大 200MB）
	r.ParseMultipartForm(200 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false, "error": "读取上传文件失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 保存上传文件到临时目录
	tmpDir := filepath.Join(s.dataDir, "update-tmp")
	os.MkdirAll(tmpDir, 0755)
	tmpFile := filepath.Join(tmpDir, "clawpanel-upload")
	if runtime.GOOS == "windows" {
		tmpFile += ".exe"
	}

	out, err := os.Create(tmpFile)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false, "error": "保存文件失败: " + err.Error(),
		})
		return
	}
	io.Copy(out, file)
	out.Close()
	os.Chmod(tmpFile, 0755)

	s.mu.Lock()
	s.state = UpdateState{
		Phase:     "validating",
		Steps:     defaultSteps(),
		Log:       []string{},
		StartedAt: time.Now().Format(time.RFC3339),
		Source:    "upload",
	}
	// 上传更新会跳过下载步骤
	s.state.Steps[3].Status = "skipped"
	s.state.Steps[3].Message = "使用本地上传文件"
	s.mu.Unlock()

	go s.doUpdateWithFile(tmpFile)

	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// handleProgress 返回 ClawPanel 更新状态快照。
func (s *Server) handleProgress(w http.ResponseWriter, r *http.Request) {
	s.setCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	// 进度接口不强制 token，避免长时间更新时 token 过期导致轮询中断。
	s.mu.Lock()
	state := s.state
	logCopy := make([]string, len(s.state.Log))
	copy(logCopy, s.state.Log)
	state.Log = logCopy
	stepsCopy := make([]UpdateStep, len(s.state.Steps))
	copy(stepsCopy, s.state.Steps)
	state.Steps = stepsCopy
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    true,
		"state": state,
	})
}

// 以下为 OpenClaw 更新相关处理器。

// handleCheckOCVersion 获取 OpenClaw 当前与最新版本信息。
func (s *Server) handleCheckOCVersion(w http.ResponseWriter, r *http.Request) {
	s.setCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if !s.checkToken(w, r) {
		return
	}

	// 通过 openclaw --version 获取当前版本
	currentVersion := "unknown"
	if verOut, verErr := exec.Command("openclaw", "--version").Output(); verErr == nil {
		currentVersion = strings.TrimSpace(string(verOut))
	}
	// 兜底从 openclaw.json 读取版本信息
	if currentVersion == "unknown" {
		cfgPath := filepath.Join(s.openClawDir, "openclaw.json")
		if data, err := os.ReadFile(cfgPath); err == nil {
			var cfg map[string]interface{}
			if json.Unmarshal(data, &cfg) == nil {
				if meta, ok := cfg["meta"].(map[string]interface{}); ok {
					if v, ok := meta["lastTouchedVersion"].(string); ok {
						currentVersion = v
					}
				}
			}
		}
	}

	// 通过 npm 查询最新版本
	latestVersion := ""
	cmd := exec.Command("npm", "view", "openclaw", "version")
	cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH")+":/usr/local/bin:/usr/bin:/bin:/snap/bin")
	if out, err := cmd.Output(); err == nil {
		latestVersion = strings.TrimSpace(string(out))
	}

	hasUpdate := latestVersion != "" && latestVersion != currentVersion && latestVersion > currentVersion

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":             true,
		"currentVersion": currentVersion,
		"latestVersion":  latestVersion,
		"hasUpdate":      hasUpdate,
	})
}

// handleStartOCUpdate 触发 OpenClaw 更新流程。
func (s *Server) handleStartOCUpdate(w http.ResponseWriter, r *http.Request) {
	s.setCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	if !s.checkToken(w, r) {
		return
	}

	s.mu.Lock()
	if s.ocState.Phase != "idle" && s.ocState.Phase != "done" && s.ocState.Phase != "error" && s.ocState.Phase != "rolled_back" {
		s.mu.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false, "error": "OpenClaw 更新正在进行中",
		})
		return
	}
	s.ocState = UpdateState{
		Phase:     "validating",
		Steps:     defaultOCSteps(),
		Log:       []string{},
		StartedAt: time.Now().Format(time.RFC3339),
	}
	s.mu.Unlock()

	go s.doOCUpdate()

	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

// handleOCProgress 返回 OpenClaw 更新状态快照。
func (s *Server) handleOCProgress(w http.ResponseWriter, r *http.Request) {
	s.setCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	s.mu.Lock()
	state := s.ocState
	logCopy := make([]string, len(s.ocState.Log))
	copy(logCopy, s.ocState.Log)
	state.Log = logCopy
	stepsCopy := make([]UpdateStep, len(s.ocState.Steps))
	copy(stepsCopy, s.ocState.Steps)
	state.Steps = stepsCopy
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    true,
		"state": state,
	})
}

// 以下为 OpenClaw 更新流程实现。

// doOCUpdate 执行 OpenClaw 更新全流程。
func (s *Server) doOCUpdate() {
	// 步骤 1：校验
	s.setOCStep(0, "running", "验证授权中...")
	s.ocLog("🔐 验证更新授权...")
	s.setOCStep(0, "done", "授权验证通过")
	s.setOCProgress(10)

	// 步骤 2：检查版本
	s.setOCStep(1, "running", "正在检测 OpenClaw 版本...")
	s.ocLog("🔍 检测 OpenClaw 版本...")

	currentVersion := "unknown"
	// 使用 openclaw --version 获取真实安装版本
	if verOut, verErr := exec.Command("openclaw", "--version").Output(); verErr == nil {
		currentVersion = strings.TrimSpace(string(verOut))
	}
	// 兜底从 openclaw.json 的 meta.lastTouchedVersion 获取版本
	cfgPath := filepath.Join(s.openClawDir, "openclaw.json")
	if currentVersion == "unknown" {
		if data, err := os.ReadFile(cfgPath); err == nil {
			var cfg map[string]interface{}
			if json.Unmarshal(data, &cfg) == nil {
				if meta, ok := cfg["meta"].(map[string]interface{}); ok {
					if v, ok := meta["lastTouchedVersion"].(string); ok {
						currentVersion = v
					}
				}
			}
		}
	}
	s.mu.Lock()
	s.ocState.FromVer = currentVersion
	s.mu.Unlock()

	latestVersion := ""
	cmd := exec.Command("npm", "view", "openclaw", "version")
	cmd.Env = append(os.Environ(), "PATH="+os.Getenv("PATH")+":/usr/local/bin:/usr/bin:/bin:/snap/bin")
	if out, err := cmd.Output(); err == nil {
		latestVersion = strings.TrimSpace(string(out))
	}

	if latestVersion == "" {
		s.ocLog("⚠️ 无法获取最新版本号，继续执行更新...")
	} else {
		s.ocLog("📦 当前版本: %s → 最新版本: %s", currentVersion, latestVersion)
		s.mu.Lock()
		s.ocState.ToVer = latestVersion
		s.mu.Unlock()
	}
	s.setOCStep(1, "done", fmt.Sprintf("当前: %s → 最新: %s", currentVersion, ternary(latestVersion == "", "未知", latestVersion)))
	s.setOCProgress(20)

	// 步骤 3：执行 npm install -g openclaw@latest。
	// 这里使用 npm 而非 openclaw update，避免更新到错误安装路径。
	s.setOCStep(2, "running", "正在更新 OpenClaw ...")

	// 尽量定位与 PATH 中 openclaw 同源的 npm
	npmBin := "npm"
	envPath := os.Getenv("PATH") + ":/usr/local/bin:/usr/bin:/bin:/snap/bin"
	// 优先使用 openclaw 同目录下的 npm
	if ocBin, err := exec.LookPath("openclaw"); err == nil {
		ocDir := filepath.Dir(ocBin)
		candidate := filepath.Join(ocDir, "npm")
		if _, err := os.Stat(candidate); err == nil {
			npmBin = candidate
			s.ocLog("� 使用 npm: %s (与 openclaw 同目录)", npmBin)
		}
	}

	targetVersion := "latest"
	if latestVersion != "" {
		targetVersion = latestVersion
	}
	s.ocLog("🚀 执行 %s install -g openclaw@%s ...", npmBin, targetVersion)

	updateCmd := exec.Command(npmBin, "install", "-g", "openclaw@"+targetVersion)
	updateCmd.Env = append(os.Environ(), "PATH="+envPath)
	updateCmd.Dir = filepath.Dir(s.openClawDir)

	// 实时读取 stdout/stderr 输出
	stdout, err := updateCmd.StdoutPipe()
	if err != nil {
		s.setOCStepError(2, "创建输出管道失败: "+err.Error())
		s.setOCError("创建输出管道失败: " + err.Error())
		return
	}
	updateCmd.Stderr = updateCmd.Stdout // 合并 stderr 到 stdout

	if err := updateCmd.Start(); err != nil {
		s.setOCStepError(2, "启动 npm install 失败: "+err.Error())
		s.setOCError("启动 npm install 失败: " + err.Error())
		return
	}

	// 逐行收集输出，便于后续分析
	var allOutput []string
	var outputMu sync.Mutex
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				lines := strings.Split(string(buf[:n]), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						s.ocLog("%s", line)
						outputMu.Lock()
						allOutput = append(allOutput, line)
						outputMu.Unlock()
						s.mu.Lock()
						pct := s.ocState.Progress
						s.mu.Unlock()
						if pct < 80 {
							s.setOCProgress(pct + 2)
						}
					}
				}
			}
			if err != nil {
				break
			}
		}
	}()

	err = updateCmd.Wait()
	if err != nil {
		exitErr := ""
		if e, ok := err.(*exec.ExitError); ok {
			exitErr = fmt.Sprintf("退出码: %d", e.ExitCode())
		} else {
			exitErr = err.Error()
		}
		s.ocLog("❌ npm install 失败: %s", exitErr)
		s.setOCStepError(2, "更新失败: "+exitErr)
		s.setOCError("npm install -g openclaw 失败: " + exitErr)
		return
	}

	// 校验更新后版本
	verCmd := exec.Command("openclaw", "--version")
	verCmd.Env = append(os.Environ(), "PATH="+envPath)
	if verOut, verErr := verCmd.Output(); verErr == nil {
		newVer := strings.TrimSpace(string(verOut))
		s.ocLog("📦 更新后版本: %s", newVer)
		s.mu.Lock()
		s.ocState.ToVer = newVer
		s.mu.Unlock()
	}

	s.setOCStep(2, "done", "OpenClaw 更新完成")
	s.ocLog("✅ OpenClaw 更新完成")
	s.setOCProgress(80)

	// 步骤 4：重启网关守护进程。
	// 终止旧网关后由 ClawPanel 的监控逻辑自动拉起新版本。
	s.setOCStep(3, "running", "正在重启 OpenClaw 网关...")
	s.ocLog("🔄 终止旧网关守护进程...")

	// 终止 openclaw-gateway 进程
	killCmd := exec.Command("pkill", "-f", "openclaw-gateway")
	killCmd.Run() // 忽略错误（进程可能本就未运行）

	s.ocLog("⏳ 等待 ClawPanel 自动重启网关...")
	// 等待 ClawPanel 监控检测并完成拉起
	time.Sleep(15 * time.Second)

	s.setOCStep(3, "done", "网关已重启")
	s.ocLog("✅ 网关重启完成")
	s.setOCProgress(100)

	// 再次通过命令确认版本
	newVersion := ""
	finalVerCmd := exec.Command("openclaw", "--version")
	finalVerCmd.Env = append(os.Environ(), "PATH="+envPath)
	if out, err := finalVerCmd.Output(); err == nil {
		newVersion = strings.TrimSpace(string(out))
	}
	if newVersion != "" && newVersion != currentVersion {
		s.mu.Lock()
		s.ocState.ToVer = newVersion
		s.mu.Unlock()
		s.ocLog("🎉 OpenClaw 更新完成！%s → %s", currentVersion, newVersion)
	} else {
		s.ocLog("🎉 OpenClaw 更新完成！")
	}

	s.mu.Lock()
	s.ocState.Phase = "done"
	s.ocState.Message = "OpenClaw 更新完成！"
	s.ocState.FinishedAt = time.Now().Format(time.RFC3339)
	s.mu.Unlock()
}

// 以下为 OpenClaw 状态辅助方法。

// setOCStep 更新 OpenClaw 指定步骤的状态与提示信息。
func (s *Server) setOCStep(idx int, status, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < len(s.ocState.Steps) {
		s.ocState.Steps[idx].Status = status
		s.ocState.Steps[idx].Message = message
	}
}

// setOCStepError 将 OpenClaw 指定步骤标记为失败，并终结流程状态。
func (s *Server) setOCStepError(idx int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < len(s.ocState.Steps) {
		s.ocState.Steps[idx].Status = "error"
		s.ocState.Steps[idx].Message = message
	}
	s.ocState.Phase = "error"
	s.ocState.FinishedAt = time.Now().Format(time.RFC3339)
}

// setOCProgress 更新 OpenClaw 更新进度百分比。
func (s *Server) setOCProgress(pct int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ocState.Progress = pct
}

// setOCError 写入 OpenClaw 更新失败信息与结束时间。
func (s *Server) setOCError(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ocState.Phase = "error"
	s.ocState.Error = msg
	s.ocState.Message = "更新失败"
	s.ocState.FinishedAt = time.Now().Format(time.RFC3339)
	s.ocState.Log = append(s.ocState.Log, "❌ "+msg)
}

// ocLog 写入 OpenClaw 更新日志并同步到状态缓存。
func (s *Server) ocLog(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	s.logger.Info(msg, "scope", "openclaw")
	s.mu.Lock()
	s.ocState.Log = append(s.ocState.Log, msg)
	s.mu.Unlock()
}

// 以下为 ClawPanel 更新流程实现。

// doUpdate 执行 ClawPanel 在线更新流程。
func (s *Server) doUpdate(preferredSource string) {
	// 步骤 1：校验
	s.setStep(0, "running", "验证授权中...")
	s.logMsg("🔐 验证更新授权...")
	s.setStep(0, "done", "授权验证通过")
	s.setProgress(5)

	// 步骤 2：检查版本
	s.setStep(1, "running", "正在检测最新版本...")
	s.logMsg("🔍 检测最新版本...")
	info, source, err := s.fetchLatestVersion()
	if err != nil {
		s.setStepError(1, "检测版本失败: "+err.Error())
		s.setError("检测版本失败: " + err.Error())
		return
	}
	if !isNewerVersion(info.LatestVersion, s.currentVersion) {
		s.setStep(1, "done", "当前已是最新版本")
		s.logMsg("✅ 当前版本 %s 已是最新", s.currentVersion)
		s.setPhase("done")
		return
	}
	s.mu.Lock()
	s.state.FromVer = s.currentVersion
	s.state.ToVer = info.LatestVersion
	s.state.Source = source
	s.mu.Unlock()
	s.setStep(1, "done", fmt.Sprintf("发现新版本 %s → %s (线路: %s)", s.currentVersion, info.LatestVersion, source))
	s.logMsg("📦 %s → %s (线路: %s)", s.currentVersion, info.LatestVersion, source)
	s.setProgress(15)

	// 步骤 3：停止服务
	s.setStep(2, "running", "正在停止 ClawPanel 服务...")
	s.logMsg("⏹️ 停止 ClawPanel 服务...")
	if err := s.stopPanel(); err != nil {
		s.logMsg("⚠️ 停止服务出错: %v (继续更新)", err)
	}
	s.setStep(2, "done", "ClawPanel 服务已停止")
	s.setProgress(25)

	// 步骤 4：下载更新包
	s.setStep(3, "running", "正在下载更新包...")
	platformKey := getPlatformKey()
	downloadURL := ""

	// 优先 GitHub，失败后回退加速线路
	if source == "github" {
		tag := info.LatestVersion
		if !strings.HasPrefix(tag, "v") {
			tag = "v" + tag
		}
		suffix := "clawpanel-" + strings.Replace(platformKey, "_", "-", -1)
		if runtime.GOOS == "windows" {
			suffix += ".exe"
		}
		downloadURL = fmt.Sprintf("%s/%s/%s", GitHubDownloadBase, tag, suffix)
	} else if urls, ok := info.DownloadURLs[platformKey]; ok {
		downloadURL = urls
	}

	if downloadURL == "" {
		s.setStepError(3, "未找到适用于当前平台的下载链接: "+platformKey)
		s.setError("未找到适用于当前平台的下载链接")
		return
	}

	tmpDir := filepath.Join(s.dataDir, "update-tmp")
	os.MkdirAll(tmpDir, 0755)
	tmpFile := filepath.Join(tmpDir, "clawpanel-new")
	if runtime.GOOS == "windows" {
		tmpFile += ".exe"
	}

	s.logMsg("📥 下载: %s", downloadURL)
	if err := s.downloadFile(downloadURL, tmpFile, source); err != nil {
		// 回退到备用线路
		s.logMsg("⚠️ %s 线路下载失败: %v, 尝试备用线路...", source, err)
		fallbackURL := ""
		fallbackSource := ""
		if source == "github" {
			// 尝试加速线路
			if urls, ok := info.DownloadURLs[platformKey]; ok {
				fallbackURL = urls
				fallbackSource = "accel"
			}
		} else {
			// 尝试 GitHub 线路
			tag := info.LatestVersion
			if !strings.HasPrefix(tag, "v") {
				tag = "v" + tag
			}
			suffix := "clawpanel-" + strings.Replace(platformKey, "_", "-", -1)
			if runtime.GOOS == "windows" {
				suffix += ".exe"
			}
			fallbackURL = fmt.Sprintf("%s/%s/%s", GitHubDownloadBase, tag, suffix)
			fallbackSource = "github"
		}
		if fallbackURL != "" {
			s.logMsg("📥 切换至 %s 线路下载...", fallbackSource)
			if err2 := s.downloadFile(fallbackURL, tmpFile, fallbackSource); err2 != nil {
				s.setStepError(3, fmt.Sprintf("两条线路均下载失败"))
				s.setError(fmt.Sprintf("下载失败: %s 线路: %v; %s 线路: %v", source, err, fallbackSource, err2))
				return
			}
			s.mu.Lock()
			s.state.Source = fallbackSource
			s.mu.Unlock()
		} else {
			s.setStepError(3, "下载失败: "+err.Error())
			s.setError("下载失败: " + err.Error())
			return
		}
	}
	s.setStep(3, "done", "下载完成")
	s.logMsg("✅ 下载完成")
	s.setProgress(60)

	// 校验 SHA256
	expectedSHA := ""
	if info.SHA256 != nil {
		expectedSHA = info.SHA256[platformKey]
	}
	if expectedSHA != "" {
		s.logMsg("🔒 校验文件 SHA256...")
		actualSHA, err := fileSHA256(tmpFile)
		if err != nil {
			s.setStepError(3, "校验失败: "+err.Error())
			s.setError("SHA256 校验失败: " + err.Error())
			return
		}
		if !strings.EqualFold(actualSHA, expectedSHA) {
			s.setStepError(3, "SHA256 不匹配，文件可能损坏")
			s.setError(fmt.Sprintf("SHA256 校验失败: 期望 %s..., 实际 %s...", expectedSHA[:16], actualSHA[:16]))
			os.Remove(tmpFile)
			return
		}
		s.logMsg("✅ SHA256 校验通过")
	} else {
		s.logMsg("⚠️ 远程未提供 SHA256，跳过校验")
	}

	// 继续执行文件替换流程
	s.doReplace(tmpFile)
}

// doUpdateWithFile 执行离线上传文件更新流程。
func (s *Server) doUpdateWithFile(tmpFile string) {
	// 步骤 1：校验
	s.setStep(0, "running", "验证授权中...")
	s.logMsg("🔐 验证更新授权...")
	s.setStep(0, "done", "授权验证通过")
	s.setProgress(5)

	// 步骤 2：校验上传文件
	s.setStep(1, "running", "检测上传文件...")
	fi, err := os.Stat(tmpFile)
	if err != nil || fi.Size() < 1024 {
		s.setStepError(1, "上传文件无效")
		s.setError("上传文件无效或过小")
		return
	}
	s.mu.Lock()
	s.state.FromVer = s.currentVersion
	s.state.ToVer = "离线上传"
	s.state.Source = "upload"
	s.mu.Unlock()
	s.setStep(1, "done", fmt.Sprintf("上传文件有效 (%.1f MB)", float64(fi.Size())/1048576))
	s.logMsg("📦 上传文件: %.1f MB", float64(fi.Size())/1048576)
	s.setProgress(15)

	// 步骤 3：停止服务
	s.setStep(2, "running", "正在停止 ClawPanel 服务...")
	s.logMsg("⏹️ 停止 ClawPanel 服务...")
	if err := s.stopPanel(); err != nil {
		s.logMsg("⚠️ 停止服务出错: %v (继续更新)", err)
	}
	s.setStep(2, "done", "ClawPanel 服务已停止")
	s.setProgress(25)

	// 步骤 4：跳过下载（已使用上传文件）
	s.setProgress(60)

	// 继续执行文件替换流程
	s.doReplace(tmpFile)
}

// doReplace 执行备份、替换、重启与回滚逻辑。
func (s *Server) doReplace(tmpFile string) {
	// 步骤 5：备份
	s.setStep(4, "running", "正在备份当前程序...")
	s.logMsg("💾 备份当前程序...")
	backupPath := s.panelBin + ".bak"
	if err := copyFile(s.panelBin, backupPath); err != nil {
		s.logMsg("⚠️ 备份失败: %v (继续更新)", err)
		s.setStep(4, "done", "备份跳过 ("+err.Error()+")")
	} else {
		s.setStep(4, "done", "已备份至 "+filepath.Base(backupPath))
		s.logMsg("✅ 已备份至 %s", backupPath)
	}
	s.setProgress(70)

	// 步骤 6：替换
	s.setStep(5, "running", "正在替换程序文件...")
	s.logMsg("🔄 替换程序文件...")

	if runtime.GOOS == "windows" {
		// Windows：重命名旧文件后再复制新文件
		os.Remove(s.panelBin + ".old")
		os.Rename(s.panelBin, s.panelBin+".old")
		if err := copyFile(tmpFile, s.panelBin); err != nil {
			// 替换失败后立即回滚
			s.logMsg("❌ 替换失败，回滚...")
			os.Rename(s.panelBin+".old", s.panelBin)
			s.setStepError(5, "替换失败，已回滚: "+err.Error())
			s.setError("替换失败，已回滚: " + err.Error())
			s.startPanel()
			return
		}
		os.Remove(s.panelBin + ".old")
	} else {
		// Linux/macOS：先删除旧文件再复制新文件
		if err := os.Remove(s.panelBin); err != nil {
			s.logMsg("⚠️ 删除旧文件失败: %v, 尝试覆盖写入...", err)
		}
		if err := copyFile(tmpFile, s.panelBin); err != nil {
			// 替换失败后尝试回滚备份
			s.logMsg("❌ 替换失败，回滚...")
			if _, berr := os.Stat(backupPath); berr == nil {
				copyFile(backupPath, s.panelBin)
			}
			os.Chmod(s.panelBin, 0755)
			s.setStepError(5, "替换失败，已回滚: "+err.Error())
			s.setError("替换失败，已回滚: " + err.Error())
			s.startPanel()
			return
		}
	}
	os.Chmod(s.panelBin, 0755)
	s.setStep(5, "done", "程序替换完成")
	s.logMsg("✅ 程序替换完成")
	s.setProgress(85)

	// 清理临时文件
	os.Remove(tmpFile)
	os.RemoveAll(filepath.Join(s.dataDir, "update-tmp"))

	// 步骤 7：启动服务
	s.setStep(6, "running", "正在启动 ClawPanel 服务...")
	s.logMsg("🚀 启动 ClawPanel 服务...")
	if err := s.startPanel(); err != nil {
		// 尝试回滚
		s.logMsg("❌ 启动失败: %v, 尝试回滚...", err)
		if _, berr := os.Stat(backupPath); berr == nil {
			os.Remove(s.panelBin)
			copyFile(backupPath, s.panelBin)
			os.Chmod(s.panelBin, 0755)
			s.logMsg("🔄 已回滚至备份文件，尝试重新启动...")
			if err2 := s.startPanel(); err2 != nil {
				s.setStepError(6, "启动失败且回滚后仍无法启动: "+err2.Error())
				s.setError("启动失败且回滚后仍无法启动，请手动处理")
				s.setPhase("rolled_back")
				return
			}
		} else {
			s.setStepError(6, "启动失败: "+err.Error())
			s.setError("启动失败: " + err.Error())
			return
		}
		s.setStep(6, "done", "已回滚并启动旧版本")
		s.logMsg("⚠️ 已回滚并启动旧版本")
		s.setPhase("rolled_back")
		return
	}

	// 校验服务是否真正启动成功
	time.Sleep(3 * time.Second)
	if !s.isPanelRunning() {
		s.logMsg("⚠️ 服务似乎未成功启动，等待更长时间...")
		time.Sleep(5 * time.Second)
		if !s.isPanelRunning() {
			s.logMsg("❌ 服务启动失败，尝试回滚...")
			if _, berr := os.Stat(backupPath); berr == nil {
				exec.Command("systemctl", "stop", "clawpanel").Run()
				time.Sleep(1 * time.Second)
				os.Remove(s.panelBin)
				copyFile(backupPath, s.panelBin)
				os.Chmod(s.panelBin, 0755)
				s.startPanel()
			}
			s.setStepError(6, "新版本启动失败，已回滚")
			s.setPhase("rolled_back")
			return
		}
	}

	s.setStep(6, "done", "ClawPanel 服务已启动")
	s.logMsg("✅ ClawPanel 服务已启动")
	s.setProgress(100)

	s.mu.Lock()
	s.state.Phase = "done"
	s.state.Message = "更新完成！"
	s.state.FinishedAt = time.Now().Format(time.RFC3339)
	s.mu.Unlock()
	s.logMsg("🎉 更新完成！")

	// 记录更新历史
	s.recordUpdateLog()
}

// stopPanel 停止 ClawPanel 服务，按平台选择对应停服策略。
func (s *Server) stopPanel() error {
	if runtime.GOOS == "windows" {
		// 优先通过系统服务停止
		exec.Command("net", "stop", "ClawPanel").Run()
		time.Sleep(2 * time.Second)

		// 通过 PID 精准结束 clawpanel.exe，排除更新器自身与父进程。
		// taskkill /IM 会误杀更新器自身，因此此处改为 PID 模式。
		selfPID := os.Getpid()
		parentPID := os.Getppid()

		// 使用 WMIC 枚举 clawpanel.exe 进程 PID
		out, err := exec.Command("wmic", "process", "where", "name='clawpanel.exe'", "get", "processid", "/value").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(strings.ToUpper(line), "PROCESSID=") {
					continue
				}
				pidStr := strings.TrimPrefix(strings.TrimPrefix(line, "ProcessId="), "PROCESSID=")
				pidStr = strings.TrimSpace(pidStr)
				pid, err := strconv.Atoi(pidStr)
				if err != nil || pid == 0 {
					continue
				}
				// 跳过更新器进程与其父进程
				if pid == selfPID || pid == parentPID {
					continue
				}
				exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid)).Run()
			}
		}
		time.Sleep(1 * time.Second)
	} else {
		if runtime.GOOS == "darwin" {
			_ = exec.Command("launchctl", "stop", "com.clawpanel.service").Run()
			_ = exec.Command("launchctl", "bootout", "system", "/Library/LaunchDaemons/com.clawpanel.service.plist").Run()
			time.Sleep(2 * time.Second)
		} else {
			if commandExists("systemctl") {
				exec.Command("systemctl", "stop", "clawpanel").Run()
				time.Sleep(3 * time.Second)
				// 若 systemd 仍显示 active，则继续等待短时间
				for i := 0; i < 5; i++ {
					out, _ := exec.Command("systemctl", "is-active", "clawpanel").Output()
					if strings.TrimSpace(string(out)) != "active" {
						break
					}
					time.Sleep(1 * time.Second)
				}
			} else {
				// 非 systemd 的 Linux 回退路径：结束面板进程但保留更新器子进程
				_ = killPanelProcessesExceptUpdater(os.Getpid(), os.Getppid())
				time.Sleep(1 * time.Second)
			}
		}
	}
	return nil
}

// startPanel 启动 ClawPanel 服务，并在失败时回退直接启动。
func (s *Server) startPanel() error {
	if runtime.GOOS == "windows" {
		err := exec.Command("net", "start", "ClawPanel").Run()
		if err != nil {
			// 服务方式启动失败时回退到直接启动
			cmd := exec.Command(s.panelBin)
			cmd.Dir = filepath.Dir(s.panelBin)
			return cmd.Start()
		}
		return nil
	}
	err := exec.Command("systemctl", "start", "clawpanel").Run()
	if runtime.GOOS == "darwin" {
		if err := exec.Command("launchctl", "kickstart", "-k", "system/com.clawpanel.service").Run(); err == nil {
			return nil
		}
		_ = exec.Command("launchctl", "load", "-w", "/Library/LaunchDaemons/com.clawpanel.service.plist").Run()
		if err := exec.Command("launchctl", "kickstart", "-k", "system/com.clawpanel.service").Run(); err == nil {
			return nil
		}
		cmd := exec.Command("bash", "-c", fmt.Sprintf("nohup %s >/dev/null 2>&1 &", s.panelBin))
		return cmd.Run()
	}
	if err != nil {
		// systemd 启动失败时回退到后台直接启动
		cmd := exec.Command("bash", "-c", fmt.Sprintf("nohup %s >/dev/null 2>&1 &", s.panelBin))
		return cmd.Run()
	}
	return nil
}

// isPanelRunning 检查 ClawPanel 服务当前是否存活。
func (s *Server) isPanelRunning() bool {
	if runtime.GOOS == "windows" {
		out, _ := exec.Command("tasklist", "/FI", "IMAGENAME eq clawpanel.exe", "/NH").Output()
		return strings.Contains(string(out), "clawpanel")
	}
	if runtime.GOOS == "darwin" {
		return isPortOpen(s.panelPort)
	}
	if commandExists("systemctl") {
		// 使用 systemctl is-active，避免 pgrep 误匹配更新器子进程
		out, _ := exec.Command("systemctl", "is-active", "clawpanel").Output()
		return strings.TrimSpace(string(out)) == "active"
	}
	return isPortOpen(s.panelPort)
}

// commandExists 判断命令是否存在于当前执行环境。
func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// isPortOpen 检查目标端口是否可建立 TCP 连接。
func isPortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 1200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// killPanelProcessesExceptUpdater 结束 ClawPanel 进程并排除更新器相关进程。
func killPanelProcessesExceptUpdater(selfPID, parentPID int) error {
	out, err := exec.Command("pgrep", "-f", "clawpanel").Output()
	if err != nil {
		return nil
	}
	for _, pidStr := range strings.Fields(string(out)) {
		pid, err := strconv.Atoi(strings.TrimSpace(pidStr))
		if err != nil || pid <= 0 {
			continue
		}
		if pid == selfPID || pid == parentPID {
			continue
		}
		cmdlineRaw, _ := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
		cmdline := strings.ReplaceAll(string(cmdlineRaw), "\x00", " ")
		if strings.Contains(cmdline, "--updater-standalone") {
			continue
		}
		_ = exec.Command("kill", "-TERM", strconv.Itoa(pid)).Run()
	}
	return nil
}

// fetchLatestVersion 拉取最新版本信息，并在多线路间自动回退。
func (s *Server) fetchLatestVersion() (*VersionInfo, string, error) {
	// 优先尝试加速线路（国内网络通常更稳定）
	info, err := s.fetchFromAccel()
	if err == nil {
		return info, "accel", nil
	}
	s.logger.Warn(fmt.Sprintf("加速服务器请求失败: %v，尝试 GitHub...", err))

	// 失败后回退到 GitHub 线路
	info, err = s.fetchFromGitHub()
	if err == nil {
		return info, "github", nil
	}
	return nil, "", fmt.Errorf("所有线路均失败: %v", err)
}

// fetchFromAccel 从加速更新源拉取版本信息。
func (s *Server) fetchFromAccel() (*VersionInfo, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(AccelUpdateJSON)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var info VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// fetchFromGitHub 从 GitHub Release 接口拉取版本信息。
func (s *Server) fetchFromGitHub() (*VersionInfo, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(GitHubReleaseAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var release struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
		PubAt   string `json:"published_at"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	info := &VersionInfo{
		LatestVersion: strings.TrimPrefix(release.TagName, "v"),
		ReleaseTime:   release.PubAt,
		ReleaseNote:   release.Body,
		DownloadURLs:  map[string]string{},
		SHA256:        map[string]string{},
	}
	for _, a := range release.Assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, "linux") && strings.Contains(name, "amd64") && !strings.Contains(name, "setup") {
			info.DownloadURLs["linux_amd64"] = a.URL
		} else if strings.Contains(name, "linux") && strings.Contains(name, "arm64") {
			info.DownloadURLs["linux_arm64"] = a.URL
		} else if strings.Contains(name, "darwin") && strings.Contains(name, "amd64") {
			info.DownloadURLs["darwin_amd64"] = a.URL
		} else if strings.Contains(name, "darwin") && strings.Contains(name, "arm64") {
			info.DownloadURLs["darwin_arm64"] = a.URL
		} else if strings.Contains(name, "windows") && strings.Contains(name, "amd64") && !strings.Contains(name, "setup") {
			info.DownloadURLs["windows_amd64"] = a.URL
		}
	}
	return info, nil
}

// downloadFile 下载更新文件并实时回写状态进度。
func (s *Server) downloadFile(url, dest, source string) error {
	client := &http.Client{Timeout: 10 * time.Minute}
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
	buf := make([]byte, 64*1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			downloaded += int64(n)
			if totalSize > 0 {
				pct := int(float64(downloaded)/float64(totalSize)*35) + 25 // 25-60%
				s.setProgress(pct)
				s.setStep(3, "running", fmt.Sprintf("下载中 %.1f MB / %.1f MB (%d%%)",
					float64(downloaded)/1048576, float64(totalSize)/1048576,
					int(float64(downloaded)/float64(totalSize)*100)))
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

// recordUpdateLog 记录更新历史，并保留最近 50 条记录。
func (s *Server) recordUpdateLog() {
	s.mu.Lock()
	state := s.state
	s.mu.Unlock()

	logEntry := map[string]interface{}{
		"time":        time.Now().Format(time.RFC3339),
		"from":        state.FromVer,
		"to":          state.ToVer,
		"source":      state.Source,
		"result":      state.Phase,
		"started_at":  state.StartedAt,
		"finished_at": state.FinishedAt,
	}

	logFile := filepath.Join(s.dataDir, "update_history.json")
	var history []map[string]interface{}
	if data, err := os.ReadFile(logFile); err == nil {
		json.Unmarshal(data, &history)
	}
	history = append(history, logEntry)
	// 仅保留最近 50 条历史
	if len(history) > 50 {
		history = history[len(history)-50:]
	}
	data, _ := json.MarshalIndent(history, "", "  ")
	os.WriteFile(logFile, data, 0644)
}

// 以下为状态辅助方法。

// setStep 更新 ClawPanel 指定步骤的状态与提示信息。
func (s *Server) setStep(idx int, status, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < len(s.state.Steps) {
		s.state.Steps[idx].Status = status
		s.state.Steps[idx].Message = message
	}
}

// setStepError 将 ClawPanel 指定步骤标记为失败，并终结流程状态。
func (s *Server) setStepError(idx int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < len(s.state.Steps) {
		s.state.Steps[idx].Status = "error"
		s.state.Steps[idx].Message = message
	}
	s.state.Phase = "error"
	s.state.FinishedAt = time.Now().Format(time.RFC3339)
}

// setProgress 更新 ClawPanel 更新进度百分比。
func (s *Server) setProgress(pct int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Progress = pct
}

// setPhase 更新 ClawPanel 流程阶段，并在结束阶段写入完成时间。
func (s *Server) setPhase(phase string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Phase = phase
	if phase == "done" || phase == "error" || phase == "rolled_back" {
		s.state.FinishedAt = time.Now().Format(time.RFC3339)
	}
}

// setError 写入 ClawPanel 更新失败信息与结束时间。
func (s *Server) setError(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Phase = "error"
	s.state.Error = msg
	s.state.Message = "更新失败"
	s.state.FinishedAt = time.Now().Format(time.RFC3339)
	s.state.Log = append(s.state.Log, "❌ "+msg)
}

// logMsg 写入 ClawPanel 更新日志并同步到状态缓存。
func (s *Server) logMsg(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	s.logger.Info(msg, "scope", "clawpanel")
	s.mu.Lock()
	s.state.Log = append(s.state.Log, msg)
	s.mu.Unlock()
}

// setCORS 设置更新接口通用跨域响应头。
func (s *Server) setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
}

// checkToken 校验请求中的更新令牌，失败时直接返回 403。
func (s *Server) checkToken(w http.ResponseWriter, r *http.Request) bool {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("X-Update-Token")
	}
	if !ValidateToken(token, s.panelPort) {
		w.WriteHeader(403)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false, "error": "授权令牌无效或已过期",
		})
		return false
	}
	return true
}

// 以下为通用工具方法。

// getPlatformKey 生成当前运行平台标识（如 linux_amd64）。
func getPlatformKey() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	return goos + "_" + goarch
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

// fileSHA256 计算指定文件的 SHA256 摘要。
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

// copyFile 将源文件内容复制到目标文件。
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

// ternary 在字符串场景下模拟三元表达式。
func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
