package build_progress_tui

import (
	"fmt"
	"math"
	"sync"
	"time"

	tui_define "github.com/HonLaderDev/HonLader/cmd/tui/define"
	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/frame"
	"github.com/HonLaderDev/HonLader/frame/task/build"
	"github.com/HonLaderDev/bedrock-world-operator/chunk"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const totalProgressScale = 100

// BuildProgressTUI 负责展示构建任务中用户可感知的进度。
type BuildProgressTUI struct {
	tui         tui_define.TextUI
	mu          sync.Mutex
	registered  bool
	progress    *mpb.Progress
	totalBar    *mpb.Bar
	localBar    *mpb.Bar
	total       int
	current     int
	groupSide   int
	groupIndex  int
	groupTotal  int
	groupChunks int
	localTotal  int
	local       int
	totalAt     time.Time
	localAt     time.Time
	localTickAt time.Time
	startAt     time.Time
}

// NewBuildProgressTUI 创建构建进度展示实例。
func NewBuildProgressTUI(tui tui_define.TextUI) *BuildProgressTUI {
	return &BuildProgressTUI{tui: tui}
}

// Register 注册构建进度事件。
func (t *BuildProgressTUI) Register(taskFrame define.TaskFrame) {
	t.mu.Lock()
	if t.registered {
		t.mu.Unlock()
		return
	}
	t.registered = true
	t.mu.Unlock()

	_ = taskFrame.EventBus().Subscribe(build.EventNameInitStart, t.OnInitStart)
	_ = taskFrame.EventBus().Subscribe(build.EventNameInitOpenWorld, t.OnInitOpenWorld)
	_ = taskFrame.EventBus().Subscribe(build.EventNameInitFinish, t.OnInitFinish)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunStart, t.OnRunStart)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunChunkGroupStart, t.OnChunkGroupStart)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunChunkGroupLoaded, t.OnChunkGroupLoaded)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunCommandsGenerated, t.OnCommandsGenerated)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunCommandSent, t.OnCommandSent)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunChunkGroupWaitLoadRetry, t.OnChunkGroupWaitLoadRetry)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunTickingAreaDisabled, t.OnTickingAreaDisabled)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunItemCleanStart, t.OnItemCleanStart)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunItemCleanFinish, t.OnItemCleanFinish)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunCommandBlocksDisableFinish, t.OnCommandBlocksDisableFinish)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunCommandBlocksReenabled, t.OnCommandBlocksReenabled)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunCommandBlocksGuardFailed, t.OnCommandBlocksGuardFailed)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunChunkGroupFinish, t.OnChunkGroupFinish)
	_ = taskFrame.EventBus().Subscribe(build.EventNameRunFinish, t.OnRunFinish)
	_ = taskFrame.EventBus().Subscribe(frame.EventNameTaskFrameTaskFailed, t.OnTaskFailed)
}

// Reset 重置本次构建进度状态。
func (t *BuildProgressTUI) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.progress != nil {
		t.progress.Shutdown()
	}
	t.progress = nil
	t.totalBar = nil
	t.localBar = nil
	t.total = 0
	t.current = 0
	t.groupSide = 0
	t.groupIndex = 0
	t.groupTotal = 0
	t.groupChunks = 0
	t.localTotal = 0
	t.local = 0
	t.totalAt = time.Time{}
	t.localAt = time.Time{}
	t.localTickAt = time.Time{}
	t.startAt = time.Time{}
}

// OnInitStart 显示初始化开始。
func (t *BuildProgressTUI) OnInitStart() {
	t.tui.Control().Print("正在初始化构建任务...\n")
}

// OnInitOpenWorld 显示正在打开源世界。
func (t *BuildProgressTUI) OnInitOpenWorld(worldPath string) {
	t.tui.Control().Print(fmt.Sprintf("正在打开建筑文件：%s\n", worldPath))
}

// OnInitFinish 显示初始化完成。
func (t *BuildProgressTUI) OnInitFinish() {
	t.tui.Control().Print("构建任务初始化完成\n")
}

// OnRunStart 初始化构建进度条。
func (t *BuildProgressTUI) OnRunStart(size define.Size, total int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.total = size.ChunkCount()
	t.current = 0
	t.groupSide = t.GuessChunkGroupSide(size, total)
	t.groupIndex = 0
	t.groupTotal = total
	t.groupChunks = 0
	now := time.Now()
	t.totalAt = now
	t.startAt = now
	t.tui.Control().Print(fmt.Sprintf(
		"开始构建：%dx%dx%d，区块 %d 个，区块组 %d 个\n",
		size.Width,
		size.Height,
		size.Length,
		t.total,
		total,
	))
	t.progress = mpb.New(mpb.WithWidth(32))
	t.totalBar = t.progress.AddBar(
		int64(t.total*totalProgressScale),
		mpb.PrependDecorators(
			decor.Name("构建总进度 ", decor.WCSyncSpaceR),
			t.TotalProgressCounter(),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncSpace),
			decor.Name(" "),
			t.TotalProgressSpeed(),
			decor.Name(" "),
			t.TotalProgressTime(),
		),
	)
	t.localBar = t.progress.AddBar(
		0,
		mpb.PrependDecorators(
			decor.Name("区块组进度 ", decor.WCSyncSpaceR),
			decor.CountersNoUnit("%d/%d", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncSpace),
			decor.Name(" "),
			decor.EwmaSpeed(nil, "%.2f 命令/秒", 60, decor.WCSyncSpace),
			decor.Name(" "),
			t.LocalProgressETA(),
		),
	)
	t.localBar.SetCurrent(0)
}

// OnChunkGroupStart 记录当前区块组进度。
func (t *BuildProgressTUI) OnChunkGroupStart(progress int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.groupIndex = progress
	t.current = t.CompletedChunksBeforeGroup(progress)
	t.groupChunks = 0
	t.resetLocalProgressLocked(0)
	t.updateTotalProgressLocked()
}

// OnChunkGroupLoaded 记录当前区块组实际区块数量。
func (t *BuildProgressTUI) OnChunkGroupLoaded(chunks map[define.ChunkPos]*chunk.Chunk, nbts map[define.ChunkPos][]map[string]any) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.groupChunks = t.EstimateCurrentGroupChunks()
	t.updateTotalProgressLocked()
}

// OnCommandsGenerated 初始化当前阶段命令发送进度。
func (t *BuildProgressTUI) OnCommandsGenerated(commandCount int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.resetLocalProgressLocked(commandCount)
}

// OnCommandSent 推进当前阶段命令发送进度。
func (t *BuildProgressTUI) OnCommandSent(string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.localTotal <= 0 {
		return
	}
	if t.local < t.localTotal {
		t.local++
		t.incrementLocalProgressLocked()
		t.updateTotalProgressLocked()
		return
	}
	t.setLocalProgressLocked(t.local, t.localTotal)
}

// OnChunkGroupWaitLoadRetry 显示区块加载等待提示。
func (t *BuildProgressTUI) OnChunkGroupWaitLoadRetry(groupPos define.ChunkPos, attempt int, timeout bool, message string) {
	if attempt == 1 || timeout {
		t.tui.Control().Print(fmt.Sprintf("等待区块加载：%v，%s\n", groupPos, message))
	}
}

// OnTickingAreaDisabled 显示常加载区域降级提示。
func (t *BuildProgressTUI) OnTickingAreaDisabled(available int, required int, reason string) {
	t.tui.Control().Print(fmt.Sprintf("常加载区域已关闭：可用 %d，需要 %d，原因：%s\n", available, required, reason))
}

// OnItemCleanStart 显示掉落物清理阶段。
func (t *BuildProgressTUI) OnItemCleanStart(define.ChunkPos, string) {
}

// OnItemCleanFinish 标记掉落物清理完成。
func (t *BuildProgressTUI) OnItemCleanFinish(define.ChunkPos, string) {
}

// OnCommandBlocksDisableFinish 显示命令方块禁用完成。
func (t *BuildProgressTUI) OnCommandBlocksDisableFinish() {
	t.tui.Control().Print("命令方块运行已关闭\n")
}

// OnCommandBlocksReenabled 显示命令方块被重新启用。
func (t *BuildProgressTUI) OnCommandBlocksReenabled() {
	t.tui.Control().Print("检测到命令方块运行被重新启用，正在再次关闭\n")
}

// OnCommandBlocksGuardFailed 显示命令方块守护失败。
func (t *BuildProgressTUI) OnCommandBlocksGuardFailed(err error) {
	t.tui.Control().Print(fmt.Sprintf("命令方块守护失败：%v\n", err))
}

// OnChunkGroupFinish 推进构建进度条。
func (t *BuildProgressTUI) OnChunkGroupFinish() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.totalBar == nil {
		return
	}
	t.completeLocalProgressLocked()
	if t.groupChunks <= 0 {
		t.groupChunks = t.EstimateCurrentGroupChunks()
	}
	t.current += t.groupChunks
	t.updateTotalProgressLocked()
}

// OnRunFinish 结束构建进度条。
func (t *BuildProgressTUI) OnRunFinish() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.completeLocalProgressLocked()
	if t.totalBar != nil {
		t.totalBar.SetTotal(int64(t.total*totalProgressScale), true)
	}
	if t.localBar != nil {
		t.localBar.SetTotal(int64(max(t.localTotal, 1)), true)
	}
	if t.progress != nil {
		t.progress.Wait()
		t.progress = nil
	}
	t.tui.Control().Print("构建完成\n")
}

// OnTaskFailed 显示构建失败提示。
func (t *BuildProgressTUI) OnTaskFailed(index int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.totalBar != nil {
		t.totalBar.Abort(true)
	}
	if t.localBar != nil {
		t.localBar.Abort(true)
	}
	if t.progress != nil {
		t.progress.Shutdown()
		t.progress = nil
	}
	t.tui.Control().Print(fmt.Sprintf("构建失败：%v\n", err))
}

// resetLocalProgressLocked 重置区块组进度条，调用方必须持有 t.mu。
func (t *BuildProgressTUI) resetLocalProgressLocked(total int) {
	if total < 0 {
		total = 0
	}
	t.local = 0
	t.localTotal = total
	t.localAt = time.Now()
	t.localTickAt = t.localAt
	t.setLocalProgressLocked(0, total)
}

// completeLocalProgressLocked 将区块组进度条补到完成，调用方必须持有 t.mu。
func (t *BuildProgressTUI) completeLocalProgressLocked() {
	total := t.localTotal
	if total < 0 {
		total = 0
	}
	t.local = total
	t.setLocalProgressLocked(total, total)
	t.updateTotalProgressLocked()
}

// setLocalProgressLocked 更新区块组进度条，调用方必须持有 t.mu。
func (t *BuildProgressTUI) setLocalProgressLocked(current int, total int) {
	if t.localBar == nil {
		return
	}
	if total < 0 {
		total = 0
	}
	if current < 0 {
		current = 0
	}
	if current > total {
		current = total
	}
	t.localBar.SetTotal(int64(total), false)
	t.localBar.SetCurrent(int64(current))
}

// incrementLocalProgressLocked 推进区块组进度并更新速度估算，调用方必须持有 t.mu。
func (t *BuildProgressTUI) incrementLocalProgressLocked() {
	if t.localBar == nil {
		return
	}
	now := time.Now()
	iterDur := now.Sub(t.localTickAt)
	if iterDur <= 0 {
		iterDur = time.Nanosecond
	}
	t.localTickAt = now
	t.localBar.EwmaIncrement(iterDur)
}

// updateTotalProgressLocked 按当前区块组命令进度折算小数构建进度，调用方必须持有 t.mu。
func (t *BuildProgressTUI) updateTotalProgressLocked() {
	if t.totalBar == nil {
		return
	}
	now := time.Now()
	iterDur := now.Sub(t.totalAt)
	if iterDur <= 0 {
		iterDur = time.Nanosecond
	}
	t.totalAt = now
	t.totalBar.EwmaSetCurrent(t.scaledTotalProgressLocked(), iterDur)
}

// scaledTotalProgressLocked 返回放大后的总构建进度，调用方必须持有 t.mu。
func (t *BuildProgressTUI) scaledTotalProgressLocked() int64 {
	if t.total <= 0 {
		return 0
	}
	progress := float64(t.current)
	groupChunks := t.groupChunks
	if groupChunks <= 0 {
		groupChunks = t.EstimateCurrentGroupChunks()
	}
	if t.localTotal > 0 && groupChunks > 0 {
		progress += float64(groupChunks) * float64(t.local) / float64(t.localTotal)
	}
	if progress < 0 {
		progress = 0
	}
	if progress > float64(t.total) {
		progress = float64(t.total)
	}
	return int64(math.Round(progress * totalProgressScale))
}

// TotalProgressCounter 返回支持小数构建进度的计数装饰器。
func (t *BuildProgressTUI) TotalProgressCounter() decor.Decorator {
	return decor.Any(func(stat decor.Statistics) string {
		current := float64(stat.Current) / totalProgressScale
		total := float64(stat.Total) / totalProgressScale
		return fmt.Sprintf("%.2f/%.0f", current, total)
	}, decor.WCSyncWidth)
}

// TotalProgressSpeed 返回按小数构建进度计算的平均速度装饰器。
func (t *BuildProgressTUI) TotalProgressSpeed() decor.Decorator {
	return decor.Any(func(stat decor.Statistics) string {
		if t.startAt.IsZero() {
			return "0.00 区块/秒"
		}
		elapsed := time.Since(t.startAt).Seconds()
		if elapsed <= 0 {
			return "0.00 区块/秒"
		}
		current := float64(stat.Current) / totalProgressScale
		return fmt.Sprintf("%.2f 区块/秒", current/elapsed)
	}, decor.WCSyncSpace)
}

// TotalProgressTime 返回区块构建预计剩余时间装饰器。
func (t *BuildProgressTUI) TotalProgressTime() decor.Decorator {
	return decor.Any(func(stat decor.Statistics) string {
		if t.startAt.IsZero() {
			return "ETA --:--:--"
		}
		elapsed := time.Since(t.startAt)
		current := float64(stat.Current) / totalProgressScale
		total := float64(stat.Total) / totalProgressScale
		eta := "--:--:--"
		if current >= total && total > 0 {
			eta = "00:00:00"
		} else if elapsed > 0 && current > 0 {
			speed := current / elapsed.Seconds()
			if speed > 0 {
				remaining := (total - current) / speed
				if remaining >= 0 {
					eta = FormatDuration(time.Duration(remaining * float64(time.Second)))
				}
			}
		}
		return fmt.Sprintf("ETA %s", eta)
	}, decor.WCSyncSpace)
}

// LocalProgressETA 返回当前区块组命令预计剩余时间装饰器。
func (t *BuildProgressTUI) LocalProgressETA() decor.Decorator {
	return decor.Any(func(stat decor.Statistics) string {
		current := float64(stat.Current)
		total := float64(stat.Total)
		if total <= 0 {
			return "ETA --:--:--"
		}
		if current >= total {
			return "ETA 00:00:00"
		}
		startAt := t.localAt
		if startAt.IsZero() || current <= 0 {
			return "ETA --:--:--"
		}
		elapsed := time.Since(startAt).Seconds()
		if elapsed <= 0 {
			return "ETA --:--:--"
		}
		speed := current / elapsed
		if speed <= 0 {
			return "ETA --:--:--"
		}
		remaining := (total - current) / speed
		if remaining < 0 {
			remaining = 0
		}
		return fmt.Sprintf("ETA %s", FormatDuration(time.Duration(remaining*float64(time.Second))))
	}, decor.WCSyncSpace)
}

// GuessChunkGroupSide 根据总区块数和区块组总数估算区块组边长。
func (t *BuildProgressTUI) GuessChunkGroupSide(size define.Size, groupTotal int) int {
	if groupTotal <= 0 {
		return 1
	}
	chunkCount := size.ChunkCount()
	if chunkCount <= 0 {
		return 1
	}
	chunksPerGroup := int(math.Ceil(float64(chunkCount) / float64(groupTotal)))
	side := int(math.Round(math.Sqrt(float64(chunksPerGroup))))
	if side <= 0 {
		return 1
	}
	return side
}

// CompletedChunksBeforeGroup 估算指定区块组之前已经完成的区块数。
func (t *BuildProgressTUI) CompletedChunksBeforeGroup(groupIndex int) int {
	if groupIndex <= 0 {
		return 0
	}
	if t.groupSide <= 0 {
		return groupIndex
	}
	chunks := groupIndex * t.groupSide * t.groupSide
	if chunks > t.total {
		return t.total
	}
	return chunks
}

// EstimateCurrentGroupChunks 估算当前区块组实际包含的区块数。
func (t *BuildProgressTUI) EstimateCurrentGroupChunks() int {
	if t.groupSide <= 0 {
		return 1
	}
	doneBefore := t.CompletedChunksBeforeGroup(t.groupIndex)
	remaining := t.total - doneBefore
	if remaining <= 0 {
		return 0
	}
	groupCapacity := t.groupSide * t.groupSide
	if remaining < groupCapacity {
		return remaining
	}
	return groupCapacity
}

// FormatDuration 将时间格式化为 HH:MM:SS。
func FormatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	totalSeconds := int64(duration.Round(time.Second).Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}
