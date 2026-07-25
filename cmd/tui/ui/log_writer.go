package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// TextUILogWriter 把 Core 日志输出到文本界面。
type TextUILogWriter struct {
	tui *TextUI
}

// NewTextUILogWriter 创建文本界面日志写入器。
func NewTextUILogWriter(tui *TextUI) TextUILogWriter {
	return TextUILogWriter{tui: tui}
}

// Write 实现 io.Writer，把 slog JSON 日志格式化成人类可读文本。
func (w TextUILogWriter) Write(data []byte) (int, error) {
	text := FormatCoreLog(data)
	w.tui.Control().Print(text)
	return len(data), nil
}

// FormatCoreLog 把 slog JSON 日志格式化成人类可读文本。
func FormatCoreLog(data []byte) string {
	rawText := string(data)
	rawText = strings.TrimRight(rawText, "\r\n")
	if rawText == "" {
		return string(data)
	}

	entry := map[string]any{}
	if err := json.Unmarshal([]byte(rawText), &entry); err != nil {
		return string(data)
	}

	level := formatLogValue(entry["level"])
	message := formatLogValue(entry["msg"])
	if level == "" {
		level = "INFO"
	}
	if message == "" {
		message = rawText
	}

	delete(entry, "time")
	delete(entry, "level")
	delete(entry, "msg")
	delete(entry, "raddr")
	delete(entry, "src")

	attrs := formatLogAttrs(entry)
	if attrs != "" {
		return fmt.Sprintf("Core 日志 [%s] %s %s\n", level, message, attrs)
	}
	return fmt.Sprintf("Core 日志 [%s] %s\n", level, message)
}

func formatLogAttrs(attrs map[string]any) string {
	if len(attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		value := formatLogValue(attrs[key])
		if value == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(parts, " ")
}

func formatLogValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}
