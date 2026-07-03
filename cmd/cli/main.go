package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "启动 CLI 失败: %v\n", err)
		os.Exit(1)
	}
}
