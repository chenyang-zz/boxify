package window

const (
	startupMainPageID  = "index"
	startupLoginPageID = "login"
)

// SelectStartupPageID 根据登录状态选择启动页面。
func SelectStartupPageID(loggedIn bool) string {
	if loggedIn {
		return startupMainPageID
	}
	return startupLoginPageID
}
