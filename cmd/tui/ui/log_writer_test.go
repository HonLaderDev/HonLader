package ui

import "testing"

func TestFormatCoreLog(t *testing.T) {
	data := []byte(`{"time":"2026-07-14T17:19:58.764613595+08:00","level":"DEBUG","msg":"Dialing to the Minecraft Bedrock Edition server","src":"dialer","raddr":"127.0.0.1:19132"}`)

	got := FormatCoreLog(data)
	want := "Core 日志 [DEBUG] Dialing to the Minecraft Bedrock Edition server\n"
	if got != want {
		t.Fatalf("FormatCoreLog() = %q, want %q", got, want)
	}
}

func TestFormatCoreLogFallback(t *testing.T) {
	data := []byte("plain log\n")

	got := FormatCoreLog(data)
	if got != string(data) {
		t.Fatalf("FormatCoreLog() = %q, want %q", got, string(data))
	}
}
