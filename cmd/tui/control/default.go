package control

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// DefaultTextUIControl 极简实现，绑定标准输入输出流
type DefaultTextUIControl struct {
	in      *bufio.Reader
	out     io.Writer
	ctx     context.Context
	cancel  context.CancelFunc
	inputCh chan string
	errCh   chan error
	wg      sync.WaitGroup
}

// NewDefaultTextUIControl 创建 DefaultTextUIControl 实例
func NewDefaultTextUIControl(in io.Reader, out io.Writer) *DefaultTextUIControl {
	ctx, cancel := context.WithCancel(context.Background())
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	return &DefaultTextUIControl{
		in:      bufio.NewReader(in),
		out:     out,
		ctx:     ctx,
		cancel:  cancel,
		inputCh: make(chan string, 1),
		errCh:   make(chan error, 1),
	}
}

// Run 阻塞运行主IO读取循环，后台持续读取输入，统一处理取消
func (d *DefaultTextUIControl) Run() error {
	defer d.cancel()
	d.wg.Add(1)
	go d.ReadLoop()

	// 阻塞等待读取协程退出
	d.wg.Wait()

	select {
	case err := <-d.errCh:
		return err
	default:
		return nil
	}
}

// ReadLoop 后台持续读取标准输入，仅由Run启动
func (d *DefaultTextUIControl) ReadLoop() {
	defer d.wg.Done()
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			line, err := d.in.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				d.errCh <- err
				return
			}
			trimmed := strings.TrimSpace(line)
			d.inputCh <- trimmed
		}
	}
}

// Print 输出文本
func (d *DefaultTextUIControl) Print(msg string) {
	_, _ = fmt.Fprint(d.out, msg)
}

// Prompt 交互式问答输入，不再新开goroutine，复用Run后台读取流
func (d *DefaultTextUIControl) Prompt(ctx context.Context, hint string) (string, error) {
	d.Print(hint)

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-d.ctx.Done():
		return "", d.ctx.Err()
	case err := <-d.errCh:
		return "", err
	case val := <-d.inputCh:
		return val, nil
	}
}

// Select 数字单选菜单，返回选中索引（从0开始）
func (d *DefaultTextUIControl) Select(ctx context.Context, title string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, errors.New("options cannot be empty")
	}

	d.Print(title)
	for i, opt := range options {
		d.Print(fmt.Sprintf("  [%d] %s\n", i+1, opt))
	}

	for {
		input, err := d.Prompt(ctx, "> ")
		if err != nil {
			return -1, err
		}
		num, err := strconv.Atoi(input)
		if err != nil {
			d.Print("无效输入，请输入数字\n")
			continue
		}
		idx := num - 1
		if idx >= 0 && idx < len(options) {
			return idx, nil
		}
		d.Print(fmt.Sprintf("无效输入，请输入 1~%d 之间的数字\n", len(options)))
	}
}

// Confirm 确认弹窗，def为空输入时的默认返回值
func (d *DefaultTextUIControl) Confirm(ctx context.Context, hint string, def bool) (bool, error) {
	for {
		input, err := d.Prompt(ctx, hint)
		if err != nil {
			return false, err
		}
		input = strings.ToLower(strings.TrimSpace(input))

		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		case "":
			return def, nil
		default:
			d.Print("无效输入，请输入 y 或 n\n")
		}
	}
}
