package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils"
)

// PromptRequired 读取必填文本，空输入时使用可选默认值。
func (t *TextUI) PromptRequired(ctx context.Context, hint string, defaultValues ...string) (string, error) {
	for {
		value, err := t.c.Prompt(ctx, hint)
		if err != nil {
			return "", err
		}
		value = strings.TrimSpace(value)
		if value != "" {
			return value, nil
		}
		if len(defaultValues) > 0 {
			return defaultValues[0], nil
		}
		t.c.Print("输入不能为空，请重新输入\n")
	}
}

// PromptBlockPos 读取并解析 x,y,z 格式的方块坐标。
func (t *TextUI) PromptBlockPos(ctx context.Context, hint string, defaultValues ...define.BlockPos) (define.BlockPos, error) {
	for {
		value, err := t.c.Prompt(ctx, hint)
		if err != nil {
			return define.BlockPos{}, err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			if len(defaultValues) == 0 {
				t.c.Print("输入不能为空，请重新输入\n")
				continue
			}
			return defaultValues[0], nil
		}
		pos, err := utils.ParseBlockPos(value)
		if err == nil {
			return pos, nil
		}
		t.c.Print(fmt.Sprintf("坐标格式错误：%v，请按 x,y,z 重新输入\n", err))
	}
}

// PromptDimension 读取并解析整数维度 ID。
func (t *TextUI) PromptDimension(ctx context.Context, hint string, defaultValues ...define.Dimension) (define.Dimension, error) {
	for {
		value, err := t.c.Prompt(ctx, hint)
		if err != nil {
			return 0, err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			if len(defaultValues) == 0 {
				t.c.Print("输入不能为空，请重新输入\n")
				continue
			}
			return defaultValues[0], nil
		}
		num, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil {
			return define.Dimension(num), nil
		}
		t.c.Print("维度格式错误，请输入整数\n")
	}
}

// PromptInt 读取并解析整数输入。
func (t *TextUI) PromptInt(ctx context.Context, hint string, defaultValues ...int) (int, error) {
	for {
		value, err := t.c.Prompt(ctx, hint)
		if err != nil {
			return 0, err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			if len(defaultValues) == 0 {
				t.c.Print("输入不能为空，请重新输入\n")
				continue
			}
			return defaultValues[0], nil
		}
		num, err := strconv.Atoi(value)
		if err == nil {
			return num, nil
		}
		t.c.Print("整数格式错误，请重新输入\n")
	}
}

// PromptFloat 读取并解析浮点数输入。
func (t *TextUI) PromptFloat(ctx context.Context, hint string, defaultValues ...float64) (float64, error) {
	for {
		value, err := t.c.Prompt(ctx, hint)
		if err != nil {
			return 0, err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			if len(defaultValues) == 0 {
				t.c.Print("输入不能为空，请重新输入\n")
				continue
			}
			return defaultValues[0], nil
		}
		num, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return num, nil
		}
		t.c.Print("数字格式错误，请重新输入\n")
	}
}
