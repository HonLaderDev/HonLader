package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HonLaderDev/bedrock-world-operator/block"
	bwo_world "github.com/HonLaderDev/bedrock-world-operator/world"
)

// Init 初始化导出任务运行时依赖。
func (e *ExportTask) Init() error {
	e.publish(EventNameInitStart)
	if e.world != nil {
		_ = e.world.CloseWorld()
	}
	e.world = nil
	e.tempDir = ""
	e.worldDir = ""
	e.outputPath = ""

	e.FillDefault()
	if e.FilePath == "" {
		e.FilePath = "export.mcworld"
	}

	tempDir, err := os.MkdirTemp("", "honlader-export-*")
	if err != nil {
		return fmt.Errorf("ExportTask.Init: create temp dir: %w", err)
	}
	e.tempDir = tempDir
	e.worldDir = filepath.Join(tempDir, "world")
	if err := os.MkdirAll(e.worldDir, 0755); err != nil {
		return fmt.Errorf("ExportTask.Init: create world dir: %w", err)
	}

	e.publish(EventNameInitOpenWorld, e.worldDir)
	w, err := bwo_world.Open(e.worldDir, nil, block.NewBlockRuntimeIDTable(false))
	if err != nil {
		return fmt.Errorf("ExportTask.Init: open world: %w", err)
	}
	e.world = w
	e.publish(EventNameInitFinish)
	return nil
}

func (e *ExportTask) openWorld() *bwo_world.BedrockWorld {
	return e.world
}
