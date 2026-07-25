package building_world

import (
	"testing"

	"github.com/HonLaderDev/HonLader/define"
)

func TestSourceBlockPosOffsetsYByWorldStart(t *testing.T) {
	w := &BuildingWorld{
		BuildingWorldConfig: BuildingWorldConfig{
			StartPos: define.BlockPos{10, -64, 20},
			EndPos:   define.BlockPos{20, 33, 30},
		},
	}

	got := w.sourceBlockPos(define.BlockPos{0, 0, 0})
	want := define.BlockPos{10, -64, 20}
	if got != want {
		t.Fatalf("sourceBlockPos() = %v, want %v", got, want)
	}
}

func TestShiftNBTIntoTargetChunkOffsetsYByWorldStart(t *testing.T) {
	w := &BuildingWorld{
		BuildingWorldConfig: BuildingWorldConfig{
			StartPos: define.BlockPos{10, -64, 20},
			EndPos:   define.BlockPos{20, 33, 30},
		},
	}

	shifted, ok := w.shiftNBTIntoTargetChunk(map[string]any{
		"x": int32(10),
		"y": int32(-64),
		"z": int32(20),
	}, define.ChunkPos{0, 0})
	if !ok {
		t.Fatal("shiftNBTIntoTargetChunk() did not keep NBT in target chunk")
	}
	if got, want := shifted["y"], int32(0); got != want {
		t.Fatalf("shifted y = %v, want %v", got, want)
	}
}
