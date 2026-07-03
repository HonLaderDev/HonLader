package build

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	packet_pb "github.com/RedLaderDev/RedLader-core-api/pb/minecraft/protocol/packet"
	"github.com/RedLaderDev/RedLader/define"
)

const (
	tickingAreaServerLimit   = 10
	tickingAreaRequiredSlots = 2
	tickingAreaPreloadSlots  = 3
	tickingAreaCommandWait   = 3 * time.Second
	tickingAreaAddPrefix     = "RedLader"
)

type tickingAreaRuntime struct {
	enabled bool
	mu      sync.Mutex
	records []tickingAreaRecord
}

type tickingAreaRecord struct {
	name    string
	group   define.ChunkPos
	release *tickingAreaRelease
}

type tickingAreaRelease struct {
	done chan struct{}
	err  error
}

// prepareTickingAreaRuntime 在构建开始前检查服务端是否至少还有两个常加载区域空位。
func (b *BuildTask) prepareTickingAreaRuntime(ctx context.Context) error {
	b.preHandleNextChunkGroup = b.PreHandleNextChunkGroup
	b.preWaitNextChunkLoad = b.PreWaitNextChunkLoad
	b.preWaitNextChunkTickingArea = b.UseTickingArea && b.PreWaitNextChunkLoad
	b.tickingAreaRequiredSlots = tickingAreaRequiredSlots
	if b.preWaitNextChunkTickingArea {
		b.tickingAreaRequiredSlots = tickingAreaPreloadSlots
		b.tickingAreaPreloadSlotsNeeded = tickingAreaPreloadSlots
	}

	b.tickingAreas = &tickingAreaRuntime{enabled: b.UseTickingArea}
	if !b.UseTickingArea {
		return nil
	}
	b.publish(EventNameRunTickingAreaCheckStart)
	available, err := b.tickingAreaAvailableSlots(ctx)
	if err != nil {
		return fmt.Errorf("BuildTask.prepareTickingAreaRuntime: %w", err)
	}
	b.publish(EventNameRunTickingAreaCheckFinish, available, b.tickingAreaRequiredSlots)
	if b.preWaitNextChunkTickingArea && available < tickingAreaPreloadSlots {
		b.preWaitNextChunkTickingArea = false
		b.tickingAreaRequiredSlots = tickingAreaRequiredSlots
		b.publish(EventNameRunTickingAreaDisabled, available, tickingAreaPreloadSlots, "available tickingarea slots are less than preload required")
	}
	if available < b.tickingAreaRequiredSlots {
		b.tickingAreas.enabled = false
		b.preWaitNextChunkTickingArea = false
		b.publish(EventNameRunTickingAreaDisabled, available, b.tickingAreaRequiredSlots, "available tickingarea slots are less than required")
	}
	return nil
}

func (b *BuildTask) useTickingArea() bool {
	return b.tickingAreas != nil && b.tickingAreas.enabled
}

func (b *BuildTask) tickingAreaAvailableSlots(ctx context.Context) (int, error) {
	resp, timeout, err := b.sendWSCommandWithTimeout(ctx, "tickingarea list", tickingAreaCommandWait)
	if err != nil {
		return 0, fmt.Errorf("BuildTask.tickingAreaAvailableSlots: %w", err)
	}
	if timeout {
		return 0, nil
	}
	used := countTickingAreas(resp)
	available := tickingAreaServerLimit - used
	if available < 0 {
		return 0, nil
	}
	return available, nil
}

func countTickingAreas(resp *packet_pb.CommandOutput) int {
	if resp == nil {
		return 0
	}
	names := make(map[string]struct{})
	re := regexp.MustCompile(`(?m)^-\s*([^:\r\n]+):`)
	for _, message := range resp.GetOutputMessages() {
		for _, match := range re.FindAllStringSubmatch(message.GetMessage(), -1) {
			if len(match) < 2 {
				continue
			}
			name := strings.TrimSpace(match[1])
			if name != "" {
				names[name] = struct{}{}
			}
		}
	}
	return len(names)
}

// prepareChunkLoadTickingArea 创建当前区块组对应的常加载区域。
func (b *BuildTask) prepareChunkLoadTickingArea(ctx context.Context, groupPos define.ChunkPos) error {
	if !b.useTickingArea() {
		return nil
	}

	b.tickingAreas.mu.Lock()
	for _, record := range b.tickingAreas.records {
		if record.group == groupPos {
			b.tickingAreas.mu.Unlock()
			return nil
		}
	}
	b.tickingAreas.mu.Unlock()

	if err := b.waitTickingAreaSlot(ctx); err != nil {
		return fmt.Errorf("BuildTask.prepareChunkLoadTickingArea: %w", err)
	}

	name := fmt.Sprintf("%s%06d", tickingAreaAddPrefix, rand.Intn(1000000))
	startX, y, startZ, endX, endZ := b.chunkLoadBounds(groupPos)
	command := fmt.Sprintf(
		"tickingarea add %d %d %d %d %d %d \"%s\" true",
		startX,
		y,
		startZ,
		endX,
		y,
		endZ,
		name,
	)
	b.publish(EventNameRunTickingAreaAddStart, groupPos, name)
	if err := b.sendSettingsCommand(ctx, command, false); err != nil {
		return fmt.Errorf("BuildTask.prepareChunkLoadTickingArea: add tickingarea: %w", err)
	}
	b.publish(EventNameRunTickingAreaAddFinish, groupPos, name)

	b.tickingAreas.mu.Lock()
	b.tickingAreas.records = append(b.tickingAreas.records, tickingAreaRecord{name: name, group: groupPos})
	b.tickingAreas.mu.Unlock()
	return nil
}

func (b *BuildTask) waitTickingAreaSlot(ctx context.Context) error {
	for {
		b.tickingAreas.mu.Lock()
		if len(b.tickingAreas.records) < b.tickingAreaRequiredSlots {
			b.tickingAreas.mu.Unlock()
			return nil
		}
		record := b.tickingAreas.records[0]
		b.tickingAreas.mu.Unlock()

		if record.release == nil {
			return fmt.Errorf("BuildTask.waitTickingAreaSlot: tickingarea %q has no release signal", record.name)
		}
		b.publish(EventNameRunTickingAreaWaitStart, record.name)
		select {
		case <-record.release.done:
			if record.release.err != nil {
				return fmt.Errorf("BuildTask.waitTickingAreaSlot: release tickingarea %q: %w", record.name, record.release.err)
			}
			b.publish(EventNameRunTickingAreaWaitFinish, record.name)
			b.tickingAreas.mu.Lock()
			if len(b.tickingAreas.records) > 0 && b.tickingAreas.records[0].name == record.name {
				b.tickingAreas.records = b.tickingAreas.records[1:]
			}
			b.tickingAreas.mu.Unlock()
		case <-ctx.Done():
			return fmt.Errorf("BuildTask.waitTickingAreaSlot: %w", ctx.Err())
		}
	}
}

// releasePreviousTickingArea 异步删除上一个区块组使用的常加载区域。
func (b *BuildTask) releasePreviousTickingArea(ctx context.Context) error {
	if !b.useTickingArea() {
		return nil
	}

	b.tickingAreas.mu.Lock()
	if len(b.tickingAreas.records) == 0 {
		b.tickingAreas.mu.Unlock()
		return nil
	}
	index := len(b.tickingAreas.records) - 1
	if index == 0 {
		b.tickingAreas.mu.Unlock()
		return nil
	}
	record := &b.tickingAreas.records[index-1]
	if record.release != nil {
		b.tickingAreas.mu.Unlock()
		return nil
	}
	name := record.name
	release := &tickingAreaRelease{done: make(chan struct{})}
	record.release = release
	b.tickingAreas.mu.Unlock()

	go func() {
		release.err = b.removeTickingArea(context.Background(), name)
		close(release.done)
	}()
	return nil
}

func (b *BuildTask) releaseAllTickingAreas(ctx context.Context) error {
	if !b.useTickingArea() {
		return nil
	}
	for {
		b.tickingAreas.mu.Lock()
		if len(b.tickingAreas.records) == 0 {
			b.tickingAreas.mu.Unlock()
			return nil
		}
		record := &b.tickingAreas.records[0]
		if record.release == nil {
			release := &tickingAreaRelease{done: make(chan struct{})}
			record.release = release
			name := record.name
			go func() {
				release.err = b.removeTickingArea(context.Background(), name)
				close(release.done)
			}()
		}
		release := record.release
		name := record.name
		b.tickingAreas.mu.Unlock()

		select {
		case <-release.done:
			if release.err != nil {
				return fmt.Errorf("BuildTask.releaseAllTickingAreas: release tickingarea %q: %w", name, release.err)
			}
			b.tickingAreas.mu.Lock()
			if len(b.tickingAreas.records) > 0 && b.tickingAreas.records[0].name == name {
				b.tickingAreas.records = b.tickingAreas.records[1:]
			}
			b.tickingAreas.mu.Unlock()
		case <-ctx.Done():
			return fmt.Errorf("BuildTask.releaseAllTickingAreas: %w", ctx.Err())
		}
	}
}

func (b *BuildTask) removeTickingArea(ctx context.Context, name string) error {
	command := fmt.Sprintf("tickingarea remove %s", name)
	b.publish(EventNameRunTickingAreaRemoveStart, name)
	_, timeout, err := b.sendWSCommandWithTimeout(ctx, command, tickingAreaCommandWait)
	if err != nil {
		b.publish(EventNameRunTickingAreaRemoveFailed, name, err)
		return fmt.Errorf("BuildTask.removeTickingArea: %w", err)
	}
	if timeout {
		err := fmt.Errorf("remove tickingarea %q timeout", name)
		b.publish(EventNameRunTickingAreaRemoveFailed, name, err)
		return fmt.Errorf("BuildTask.removeTickingArea: %w", err)
	}
	b.publish(EventNameRunTickingAreaRemoveFinish, name)
	return nil
}
