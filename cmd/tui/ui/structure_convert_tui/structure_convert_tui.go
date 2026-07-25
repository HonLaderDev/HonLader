package structure_convert_tui

import (
	"fmt"
	"sync"
	"time"

	tui_define "github.com/HonLaderDev/HonLader/cmd/tui/define"
	structure_utils "github.com/HonLaderDev/HonLader/utils/structure"
	ws_define "github.com/HonLaderDev/WaterStructure/define"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// StructureConvertTUI 负责显示结构转换进度。
type StructureConvertTUI struct {
	tui      tui_define.TextUI
	mu       sync.Mutex
	progress *mpb.Progress
	bar      *mpb.Bar
	total    int
	lastTick time.Time
}

// NewStructureConvertTUI 创建结构转换进度界面。
func NewStructureConvertTUI(tui tui_define.TextUI) *StructureConvertTUI {
	return &StructureConvertTUI{tui: tui}
}

// Register 注册结构转换事件。
func (t *StructureConvertTUI) Register(converter *structure_utils.StructureConverter) {
	if converter == nil {
		return
	}
	_ = converter.EventBus().Subscribe(structure_utils.EventNameConvertStart, t.OnConvertStart)
	_ = converter.EventBus().Subscribe(structure_utils.EventNameConvertParsed, t.OnConvertParsed)
	_ = converter.EventBus().Subscribe(structure_utils.EventNameConvertProgressInit, t.OnConvertProgressInit)
	_ = converter.EventBus().Subscribe(structure_utils.EventNameConvertProgress, t.OnConvertProgress)
	_ = converter.EventBus().Subscribe(structure_utils.EventNameConvertSuccess, t.OnConvertSuccess)
}

// Reset 重置转换进度。
func (t *StructureConvertTUI) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.progress != nil {
		t.progress.Shutdown()
	}
	t.progress = nil
	t.bar = nil
	t.total = 0
	t.lastTick = time.Time{}
}

// OnConvertStart 显示转换开始。
func (t *StructureConvertTUI) OnConvertStart(srcPath, destPath string) {
	t.tui.Control().Print(fmt.Sprintf("检测到非 mcworld/zip 文件，开始转换结构：%s\n", srcPath))
}

// OnConvertParsed 显示结构解析结果。
func (t *StructureConvertTUI) OnConvertParsed(name string, size ws_define.Size) {
	t.tui.Control().Print(fmt.Sprintf("结构类型：%s，尺寸：%dx%dx%d\n", name, size.Width, size.Height, size.Length))
}

// OnConvertProgressInit 初始化转换进度条。
func (t *StructureConvertTUI) OnConvertProgressInit(total int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.total = total
	t.lastTick = time.Now()
	t.progress = mpb.New(mpb.WithWidth(32))
	t.bar = t.progress.AddBar(
		int64(total),
		mpb.PrependDecorators(
			decor.Name("转换进度 ", decor.WCSyncSpaceR),
			decor.CountersNoUnit("%d/%d", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncSpace),
			decor.EwmaSpeed(nil, " %.2f 区块/秒", 30),
			decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncSpace),
		),
	)
}

// OnConvertProgress 推进转换进度。
func (t *StructureConvertTUI) OnConvertProgress() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.bar != nil {
		now := time.Now()
		iterDur := now.Sub(t.lastTick)
		if iterDur <= 0 {
			iterDur = time.Millisecond
		}
		t.lastTick = now
		t.bar.EwmaIncrement(iterDur)
	}
}

// OnConvertSuccess 显示转换完成。
func (t *StructureConvertTUI) OnConvertSuccess(path string) {
	t.mu.Lock()
	if t.bar != nil {
		t.bar.SetTotal(int64(t.total), true)
	}
	if t.progress != nil {
		t.progress.Wait()
		t.progress = nil
		t.bar = nil
	}
	t.mu.Unlock()
	t.tui.Control().Print(fmt.Sprintf("结构转换完成：%s\n", path))
}
