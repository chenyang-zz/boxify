//go:build windows

package monitor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modKernel32                     = windows.NewLazySystemDLL("kernel32.dll")
	modAdvapi32                     = windows.NewLazySystemDLL("advapi32.dll")
	modWtsapi32                     = windows.NewLazySystemDLL("wtsapi32.dll")
	procOpenProcessToken            = modAdvapi32.NewProc("OpenProcessToken")
	procDuplicateTokenEx            = modAdvapi32.NewProc("DuplicateTokenEx")
	procSetTokenInformation         = modAdvapi32.NewProc("SetTokenInformation")
	procCreateProcessAsUserW        = modAdvapi32.NewProc("CreateProcessAsUserW")
	procWTSQuerySessionInformationW = modWtsapi32.NewProc("WTSQuerySessionInformationW")
	procWTSFreeMemory               = modWtsapi32.NewProc("WTSFreeMemory")
	modUserenv                      = windows.NewLazySystemDLL("userenv.dll")
	procCreateEnvironmentBlock      = modUserenv.NewProc("CreateEnvironmentBlock")
	procDestroyEnvironmentBlock     = modUserenv.NewProc("DestroyEnvironmentBlock")
)

const (
	TOKEN_DUPLICATE            = 0x0002
	TOKEN_QUERY                = 0x0008
	TOKEN_ASSIGN_PRIMARY       = 0x0001
	TOKEN_ADJUST_PRIVILEGES    = 0x0020
	TOKEN_ADJUST_SESSIONID     = 0x0100
	TOKEN_ADJUST_DEFAULT       = 0x0080
	SecurityImpersonation      = 2
	TokenPrimary               = 1
	TokenSessionId             = 12
	CREATE_UNICODE_ENVIRONMENT = 0x00000400
	NORMAL_PRIORITY_CLASS      = 0x00000020
	CREATE_NEW_CONSOLE         = 0x00000010
	CREATE_NO_WINDOW           = 0x08000000
)

// STARTUPINFOW 对应 Windows CreateProcess 系列 API 的启动参数结构。
type STARTUPINFOW struct {
	Cb              uint32         // 结构体大小
	LpReserved      *uint16        // 保留字段
	LpDesktop       *uint16        // 桌面名称
	LpTitle         *uint16        // 窗口标题
	DwX             uint32         // 窗口 X 坐标
	DwY             uint32         // 窗口 Y 坐标
	DwXSize         uint32         // 窗口宽度
	DwYSize         uint32         // 窗口高度
	DwXCountChars   uint32         // 控制台宽度（字符）
	DwYCountChars   uint32         // 控制台高度（字符）
	DwFillAttribute uint32         // 控制台属性
	DwFlags         uint32         // 启动标志
	WShowWindow     uint16         // 窗口显示方式
	CbReserved2     uint16         // 保留字段长度
	LpReserved2     *byte          // 保留字段指针
	HStdInput       windows.Handle // 标准输入句柄
	HStdOutput      windows.Handle // 标准输出句柄
	HStdError       windows.Handle // 标准错误句柄
}

// PROCESS_INFORMATION 对应 Windows CreateProcess 系列 API 的进程输出结构。
type PROCESS_INFORMATION struct {
	HProcess    windows.Handle // 进程句柄
	HThread     windows.Handle // 主线程句柄
	DwProcessId uint32         // 进程 ID
	DwThreadId  uint32         // 主线程 ID
}

// launchAsInteractiveUser 通过 explorer.exe 令牌在交互会话（session 1）中启动进程。
// cmdLine 为完整命令行，extraEnv 为附加环境变量。
func launchAsInteractiveUser(cmdLine, workDir string, extraEnv []string) error {
	// 查找 session 1 中的 explorer.exe。
	explorerPID, err := findExplorerPID()
	if err != nil {
		return fmt.Errorf("find explorer.exe: %w", err)
	}
	// 打开 explorer 进程。
	hProcess, err := windows.OpenProcess(
		windows.PROCESS_QUERY_INFORMATION,
		false,
		explorerPID,
	)
	if err != nil {
		return fmt.Errorf("OpenProcess explorer: %w", err)
	}
	defer windows.CloseHandle(hProcess)

	// 打开进程令牌。
	var hToken windows.Token
	r, _, e := procOpenProcessToken.Call(
		uintptr(hProcess),
		TOKEN_DUPLICATE|TOKEN_QUERY,
		uintptr(unsafe.Pointer(&hToken)),
	)
	if r == 0 {
		return fmt.Errorf("OpenProcessToken: %w", e)
	}
	defer hToken.Close()

	// 复制为主令牌。
	var hDupToken windows.Token
	r, _, e = procDuplicateTokenEx.Call(
		uintptr(hToken),
		TOKEN_ASSIGN_PRIMARY|TOKEN_DUPLICATE|TOKEN_QUERY|TOKEN_ADJUST_PRIVILEGES|TOKEN_ADJUST_SESSIONID|TOKEN_ADJUST_DEFAULT,
		0,
		SecurityImpersonation,
		TokenPrimary,
		uintptr(unsafe.Pointer(&hDupToken)),
	)
	if r == 0 {
		return fmt.Errorf("DuplicateTokenEx: %w", e)
	}
	defer hDupToken.Close()

	// 强制令牌会话为 1，确保进程运行在交互桌面。
	sessionID := uint32(1)
	procSetTokenInformation.Call(
		uintptr(hDupToken),
		TokenSessionId,
		uintptr(unsafe.Pointer(&sessionID)),
		uintptr(unsafe.Sizeof(sessionID)),
	)

	// 为用户令牌创建环境变量块。
	var rawEnvBlock *uint16
	r, _, e = procCreateEnvironmentBlock.Call(
		uintptr(unsafe.Pointer(&rawEnvBlock)),
		uintptr(hDupToken),
		0,
	)
	createEnvOK := r != 0 && rawEnvBlock != nil

	// 构建最终环境变量块（用户环境 + 额外覆盖项）。
	var finalEnvBlock uintptr
	var mergedEnv []uint16
	if len(extraEnv) > 0 {
		if createEnvOK {
			mergedEnv = mergeEnvBlock(rawEnvBlock, extraEnv)
			procDestroyEnvironmentBlock.Call(uintptr(unsafe.Pointer(rawEnvBlock)))
		} else {
			mergedEnv = mergeEnvPairs(os.Environ(), extraEnv)
		}
		finalEnvBlock = uintptr(unsafe.Pointer(&mergedEnv[0]))
	} else {
		finalEnvBlock = uintptr(unsafe.Pointer(rawEnvBlock))
		if createEnvOK {
			defer procDestroyEnvironmentBlock.Call(uintptr(unsafe.Pointer(rawEnvBlock)))
		}
	}

	// 构造 UTF-16 命令行与工作目录指针。
	cmdLinePtr, err := windows.UTF16PtrFromString(cmdLine)
	if err != nil {
		return fmt.Errorf("UTF16PtrFromString cmdLine: %w", err)
	}
	wd, err := windows.UTF16PtrFromString(workDir)
	if err != nil {
		return fmt.Errorf("UTF16PtrFromString workDir: %w", err)
	}

	si := STARTUPINFOW{
		Cb: uint32(unsafe.Sizeof(STARTUPINFOW{})),
	}
	// 指定交互桌面 winsta0\default。
	desktop, _ := windows.UTF16PtrFromString(`winsta0\default`)
	si.LpDesktop = desktop

	var pi PROCESS_INFORMATION

	// 不使用 CREATE_NO_WINDOW，NapCat 注入 QQ 需要桌面访问能力。
	creationFlags := uint32(CREATE_UNICODE_ENVIRONMENT | NORMAL_PRIORITY_CLASS)

	r, _, e = procCreateProcessAsUserW.Call(
		uintptr(hDupToken),
		0,
		uintptr(unsafe.Pointer(cmdLinePtr)),
		0,
		0,
		0,
		uintptr(creationFlags),
		finalEnvBlock,
		uintptr(unsafe.Pointer(wd)),
		uintptr(unsafe.Pointer(&si)),
		uintptr(unsafe.Pointer(&pi)),
	)
	runtime.KeepAlive(mergedEnv)
	if r == 0 {
		return fmt.Errorf("CreateProcessAsUserW: %w", e)
	}

	windows.CloseHandle(pi.HProcess)
	windows.CloseHandle(pi.HThread)
	return nil
}

// findExplorerPID 返回 session 1 中 explorer.exe 的进程 ID。
func findExplorerPID() (uint32, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err := windows.Process32First(snapshot, &pe); err != nil {
		return 0, err
	}
	for {
		name := windows.UTF16ToString(pe.ExeFile[:])
		if name == "explorer.exe" {
			// 确认进程位于 session 1。
			h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, pe.ProcessID)
			if err == nil {
				var sessionID uint32
				if err2 := windows.ProcessIdToSessionId(pe.ProcessID, &sessionID); err2 == nil && sessionID == 1 {
					windows.CloseHandle(h)
					return pe.ProcessID, nil
				}
				windows.CloseHandle(h)
			}
		}
		if err := windows.Process32Next(snapshot, &pe); err != nil {
			break
		}
	}
	return 0, fmt.Errorf("explorer.exe not found in session 1")
}

// mergeEnvBlock 解析 rawBlock 指向的 UTF-16 环境变量块，
// 合并 extraEnv 覆盖项后返回 CreateProcessAsUserW 可用的双零结尾环境块。
func mergeEnvBlock(rawBlock *uint16, extraEnv []string) []uint16 {
	// 解析现有环境变量块（UTF-16，单零分隔，双零结尾）。
	env := map[string]string{}
	if rawBlock != nil {
		ptr := unsafe.Pointer(rawBlock)
		for {
			// 读取单条以零结尾的 UTF-16 字符串。
			var chars []uint16
			for {
				w := *(*uint16)(ptr)
				ptr = unsafe.Add(ptr, unsafe.Sizeof(uint16(0)))
				if w == 0 {
					break
				}
				chars = append(chars, w)
			}
			if len(chars) == 0 {
				break // 双零结尾：环境块结束
			}
			kv := windows.UTF16ToString(chars)
			if idx := strings.Index(kv, "="); idx > 0 {
				env[strings.ToUpper(kv[:idx])] = kv[idx+1:]
			}
		}
	}
	// 使用 extraEnv 覆盖变量。
	for _, kv := range extraEnv {
		if idx := strings.Index(kv, "="); idx > 0 {
			env[strings.ToUpper(kv[:idx])] = kv[idx+1:]
		}
	}
	// 重建环境变量块。
	var out []uint16
	for k, v := range env {
		entry := k + "=" + v
		u16, _ := windows.UTF16FromString(entry)
		out = append(out, u16...) // UTF16FromString 已包含末尾零
	}
	out = append(out, 0) // 补齐双零结尾
	return out
}

// mergeEnvPairs 合并基础环境变量与覆盖项，并转换为 Windows 双零结尾环境块。
func mergeEnvPairs(baseEnv, extraEnv []string) []uint16 {
	env := map[string]string{}
	for _, kv := range baseEnv {
		if idx := strings.Index(kv, "="); idx > 0 {
			env[strings.ToUpper(kv[:idx])] = kv[idx+1:]
		}
	}
	for _, kv := range extraEnv {
		if idx := strings.Index(kv, "="); idx > 0 {
			env[strings.ToUpper(kv[:idx])] = kv[idx+1:]
		}
	}

	var out []uint16
	for k, v := range env {
		entry := k + "=" + v
		u16, _ := windows.UTF16FromString(entry)
		out = append(out, u16...)
	}
	out = append(out, 0)
	return out
}

// launchNapCatInUserSession 在交互会话中拉起 NapCat。
// 策略：先写 PowerShell 脚本并通过 schtasks 执行，失败后回退 CreateProcessAsUser。
func launchNapCatInUserSession(exePath, workDir string) error {
	batPath := findNapCatLauncherBat(workDir)
	if batPath == "" {
		return fmt.Errorf("launcher-user.bat not found in %s", workDir)
	}
	batDir := filepath.Dir(batPath)

	// 写入 PS1 包装脚本，避免 bat 阻塞。
	psContent := fmt.Sprintf(
		"$p = Start-Process -FilePath 'cmd.exe' -ArgumentList '/c \"%s\"' -WorkingDirectory '%s' -WindowStyle Hidden -PassThru; exit 0\r\n",
		batPath, batDir)
	psFile := filepath.Join(os.TempDir(), "napcat_launch.ps1")
	if err := os.WriteFile(psFile, []byte(psContent), 0644); err != nil {
		return fmt.Errorf("write ps1: %w", err)
	}

	taskCmd := fmt.Sprintf(`powershell.exe -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File "%s"`, psFile)

	// 优先使用 WTS 查询到的用户名执行 schtasks。
	username := getSession1Username()
	taskName := "ClawPanelStartNapCat"
	exec.Command("schtasks", "/Delete", "/TN", taskName, "/F").Run()

	var createErr error
	if username != "" {
		createErr = exec.Command("schtasks", "/Create", "/F",
			"/TN", taskName, "/SC", "ONCE", "/ST", "00:00",
			"/RU", username, "/TR", taskCmd, "/RL", "HIGHEST",
		).Run()
	}
	if username == "" || createErr != nil {
		// 回退：不传 /RU，仍尝试由 schtasks 调度执行。
		createErr = exec.Command("schtasks", "/Create", "/F",
			"/TN", taskName, "/SC", "ONCE", "/ST", "00:00",
			"/TR", taskCmd, "/RL", "HIGHEST",
		).Run()
	}
	if createErr != nil {
		// 最后回退到 CreateProcessAsUser。
		cmdLine, extraEnv, err := buildNapCatCommandLine(exePath, workDir)
		if err != nil {
			return fmt.Errorf("buildNapCatCommandLine: %w", err)
		}
		return launchAsInteractiveUser(cmdLine, workDir, extraEnv)
	}

	runErr := exec.Command("schtasks", "/Run", "/TN", taskName).Run()
	go func() {
		time.Sleep(20 * time.Second)
		exec.Command("schtasks", "/Delete", "/TN", taskName, "/F").Run()
		os.Remove(psFile)
	}()
	if runErr != nil {
		return fmt.Errorf("schtasks run: %w", runErr)
	}
	return nil
}

// getSession1Username 返回 session 1 当前登录用户（domain\username）。
// 通过 WTSQuerySessionInformationW 获取 Unicode 用户名。
func getSession1Username() string {
	const WTSUserName = 5
	const WTSDomainName = 7
	const WTS_CURRENT_SERVER_HANDLE = 0

	getDomainUser := func(sessionID uint32, infoClass uintptr) string {
		var pBuf *uint16
		var bytes uint32
		r, _, _ := procWTSQuerySessionInformationW.Call(
			WTS_CURRENT_SERVER_HANDLE,
			uintptr(sessionID),
			infoClass,
			uintptr(unsafe.Pointer(&pBuf)),
			uintptr(unsafe.Pointer(&bytes)),
		)
		if r == 0 || pBuf == nil {
			return ""
		}
		defer procWTSFreeMemory.Call(uintptr(unsafe.Pointer(pBuf)))
		// pBuf 指向以零结尾的 UTF-16 字符串。
		nChars := bytes / 2
		if nChars == 0 {
			return ""
		}
		u16 := unsafe.Slice(pBuf, nChars)
		return windows.UTF16ToString(u16)
	}

	domain := getDomainUser(1, WTSDomainName)
	user := getDomainUser(1, WTSUserName)
	if user == "" {
		return ""
	}
	if domain != "" && !strings.EqualFold(domain, ".") {
		return domain + `\` + user
	}
	return user
}

// findNapCatLauncherBat 在 NapCat Shell 目录树中定位 launcher 脚本。
func findNapCatLauncherBat(shellDir string) string {
	// 优先检查内层目录（versions/.../napcat/）。
	innerDir := findNapCatInnerDir(shellDir)
	if innerDir != "" {
		p := filepath.Join(innerDir, "launcher-user.bat")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		p = filepath.Join(innerDir, "launcher.bat")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// 再检查顶层 shell 目录。
	for _, name := range []string{"launcher-user.bat", "napcat.bat"} {
		p := filepath.Join(shellDir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// buildNapCatCommandLine 组装 NapCatWinBootMain.exe 启动命令。
// 参数与环境变量尽量与 launcher-user.bat 保持一致。
func buildNapCatCommandLine(_, napcatShellDir string) (cmdLine string, extraEnv []string, err error) {
	// 先从注册表定位 QQ.exe。
	qqPath := findQQExePath()
	if qqPath == "" {
		err = fmt.Errorf("QQ.exe not found in registry or common paths")
		return
	}

	// 定位内层资源目录（包含 napcat.mjs）。
	innerDir := findNapCatInnerDir(napcatShellDir)
	if innerDir == "" {
		innerDir = napcatShellDir // 回退到顶层目录
	}

	// 优先使用内层 NapCatWinBootMain.exe（实际注入器）。
	innerBootMain := filepath.Join(innerDir, "NapCatWinBootMain.exe")
	if _, e := os.Stat(innerBootMain); e != nil {
		// 回退到顶层 shell 目录。
		innerBootMain = filepath.Join(napcatShellDir, "NapCatWinBootMain.exe")
	}

	// NAPCAT_MAIN_PATH 使用正斜杠，与 bat 行为一致。
	napcatMjs := filepath.Join(innerDir, "napcat.mjs")
	napcatMjsFwd := strings.ReplaceAll(napcatMjs, `\`, `/`)

	// 在内层目录写入 loadNapCat.js（与 bat 逻辑一致）。
	loadPath := filepath.Join(innerDir, "loadNapCat.js")
	loadContent := fmt.Sprintf(`(async () => {await import("file:///%s")})()`, napcatMjsFwd)
	_ = os.WriteFile(loadPath, []byte(loadContent), 0644)

	// 优先使用内层 NapCatWinBootHook.dll。
	injectDll := filepath.Join(innerDir, "NapCatWinBootHook.dll")
	if _, e := os.Stat(injectDll); e != nil {
		injectDll = filepath.Join(napcatShellDir, "NapCatWinBootHook.dll")
	}

	extraEnv = []string{
		"NAPCAT_PATCH_PACKAGE=" + filepath.Join(innerDir, "qqnt.json"),
		"NAPCAT_LOAD_PATH=" + loadPath,
		"NAPCAT_INJECT_PATH=" + injectDll,
		"NAPCAT_LAUNCHER_PATH=" + innerBootMain,
		"NAPCAT_MAIN_PATH=" + napcatMjs,
	}

	// 命令格式："NapCatWinBootMain.exe" "D:\QQ\QQ.exe" "NapCatWinBootHook.dll"
	cmdLine = fmt.Sprintf(`"%s" "%s" "%s"`, innerBootMain, qqPath, injectDll)
	return
}

// findNapCatInnerDir 通过遍历 NapCat Shell 目录树定位内层资源目录。
func findNapCatInnerDir(shellDir string) string {
	var found string
	filepath.WalkDir(shellDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil
		}
		if !d.IsDir() && strings.EqualFold(d.Name(), "napcat.mjs") {
			found = filepath.Dir(path)
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// findQQExePath 先从注册表查找 QQ.exe，失败后检查常见安装路径。
func findQQExePath() string {
	// 使用与 launcher.bat 相同的注册表项。
	out, err := exec.Command("reg", "query",
		`HKEY_LOCAL_MACHINE\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\QQ`,
		"/v", "UninstallString").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(strings.ToLower(line), "uninstallstring") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					uninstStr := strings.Trim(parts[len(parts)-1], `"`)
					qqDir := filepath.Dir(uninstStr)
					qqExe := filepath.Join(qqDir, "QQ.exe")
					if _, e := os.Stat(qqExe); e == nil {
						return qqExe
					}
				}
			}
		}
	}
	// 回退到常见安装路径。
	for _, p := range []string{`D:\QQ\QQ.exe`, `C:\Program Files\Tencent\QQ\QQ.exe`, `C:\QQ\QQ.exe`} {
		if _, e := os.Stat(p); e == nil {
			return p
		}
	}
	return ""
}

// launchNapCatViaBat 通过 launcher bat 启动 NapCat（回退路径）。
func launchNapCatViaBat(napcatDir string) error {
	batPath := filepath.Join(napcatDir, "launcher-user.bat")
	if _, err := os.Stat(batPath); err != nil {
		// 回退到顶层 napcat.bat。
		batPath = filepath.Join(filepath.Dir(napcatDir), "napcat.bat")
		if _, err2 := os.Stat(batPath); err2 != nil {
			return fmt.Errorf("no launcher bat found")
		}
	}
	cmdLine := fmt.Sprintf(`cmd.exe /c "%s"`, batPath)
	return launchAsInteractiveUser(cmdLine, napcatDir, nil)
}

// launchNapCatViaSchtasks 通过 schtasks 执行命令（用于回退场景）。
func launchNapCatViaSchtasks(cmdLine, workDir string) error {
	psContent := fmt.Sprintf("Set-Location '%s'\n%s\n", workDir, cmdLine)
	psFile := filepath.Join(os.TempDir(), "napcat_launch.ps1")
	_ = os.WriteFile(psFile, []byte(psContent), 0644)

	username := getInteractiveUsername()
	taskName := "ClawPanelStartNapCat"
	tr := fmt.Sprintf(`powershell.exe -NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File "%s"`, psFile)
	exec.Command("schtasks", "/Delete", "/TN", taskName, "/F").Run()
	var createArgs []string
	createArgs = []string{"/Create", "/F", "/TN", taskName, "/SC", "ONCE", "/ST", "00:00", "/TR", tr, "/RL", "HIGHEST"}
	if username != "" {
		createArgs = append(createArgs, "/RU", username)
	}
	err := exec.Command("schtasks", createArgs...).Run()
	if err == nil {
		if err = exec.Command("schtasks", "/Run", "/TN", taskName).Run(); err == nil {
			go func() {
				time.Sleep(15 * time.Second)
				exec.Command("schtasks", "/Delete", "/TN", taskName, "/F").Run()
				os.Remove(psFile)
			}()
			return nil
		}
	}
	return fmt.Errorf("schtasks failed: %w", err)
}
