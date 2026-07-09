package ui

import (
	"github.com/HonLaderDev/HonLader/cmd/tui/control"
	"github.com/HonLaderDev/HonLader/define"
)

type TextUI struct {
	l      define.Launcher
	c      control.TextUIControl
	Format TextUIFormat
}

func NewTextUI(launcher define.Launcher, control control.TextUIControl) *TextUI {
	return &TextUI{
		l: launcher,
		c: control,
	}
}
