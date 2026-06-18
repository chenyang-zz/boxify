package window

import "testing"

func TestSelectStartupPageID(t *testing.T) {
	tests := []struct {
		name     string
		loggedIn bool
		want     string
	}{
		{name: "未登录时启动登录页", loggedIn: false, want: "login"},
		{name: "已登录时启动主页面", loggedIn: true, want: "index"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectStartupPageID(tt.loggedIn); got != tt.want {
				t.Fatalf("SelectStartupPageID(%v) = %q, want %q", tt.loggedIn, got, tt.want)
			}
		})
	}
}
