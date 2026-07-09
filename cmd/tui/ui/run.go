package ui

import (
	"context"
	"fmt"
)

func (t *TextUI) Run(ctx context.Context) error {
	sel, err := t.c.Select(
		ctx,
		"请选择操作：\n",
		[]string{
			"开始构建",
			"开始导出",
		},
	)
	if err != nil {
		return err
	}
	switch sel {
	case 0:
		return t.StartBuild(ctx)
	case 1:
		return fmt.Errorf("导出功能暂未实现")
	default:
		return fmt.Errorf("未知操作：%d", sel)
	}
}
