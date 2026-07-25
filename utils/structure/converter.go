package structure

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	water_structure "github.com/HonLaderDev/WaterStructure/structure"
	"github.com/HonLaderDev/bedrock-world-operator/block"
	bwo_define "github.com/HonLaderDev/bedrock-world-operator/define"
	bwo_world "github.com/HonLaderDev/bedrock-world-operator/world"
	"github.com/asaskevich/EventBus"
)

const (
	EventNameConvertStart        = "structure.convert.start"
	EventNameConvertParsed       = "structure.convert.parsed"
	EventNameConvertProgressInit = "structure.convert.progress.init"
	EventNameConvertProgress     = "structure.convert.progress"
	EventNameConvertSuccess      = "structure.convert.success"
)

// StructureConverter 负责把结构文件转换成 mcworld。
type StructureConverter struct {
	eventBus EventBus.Bus
}

// ConvertResult 描述结构转换结果。
type ConvertResult struct {
	Path      string
	Width     int
	Height    int
	Length    int
	WorldName string
}

// NewStructureConverter 创建结构转换器。
func NewStructureConverter() *StructureConverter {
	return &StructureConverter{eventBus: EventBus.New()}
}

// EventBus 返回转换器事件总线。
func (c *StructureConverter) EventBus() EventBus.Bus {
	return c.eventBus
}

// ConvertToMCWorld 把结构文件转换成 mcworld 文件。
func (c *StructureConverter) ConvertToMCWorld(srcPath, destPath string) (ConvertResult, error) {
	return c.ConvertToMCWorldWithName(srcPath, destPath, DefaultWorldName(srcPath))
}

// ConvertToMCWorldWithName 把结构文件转换成指定世界名的 mcworld 文件。
func (c *StructureConverter) ConvertToMCWorldWithName(srcPath, destPath, worldName string) (ConvertResult, error) {
	if c == nil {
		return ConvertResult{}, fmt.Errorf("结构转换器不能为空")
	}
	c.publish(EventNameConvertStart, srcPath, destPath)

	tempDir, err := os.MkdirTemp("", "honlader-mcworld-*")
	if err != nil {
		return ConvertResult{}, err
	}
	defer os.RemoveAll(tempDir)

	worldDir := filepath.Join(tempDir, "world")
	if err := os.MkdirAll(worldDir, 0755); err != nil {
		return ConvertResult{}, err
	}

	bedrockWorld, err := bwo_world.Open(worldDir, nil, block.NewBlockRuntimeIDTable(false))
	if err != nil {
		return ConvertResult{}, err
	}

	file, err := os.Open(srcPath)
	if err != nil {
		_ = bedrockWorld.CloseWorld()
		return ConvertResult{}, err
	}
	defer file.Close()

	structure, err := water_structure.StructureFromFile(file)
	if err != nil {
		_ = bedrockWorld.CloseWorld()
		return ConvertResult{}, err
	}
	defer structure.Close()

	size := structure.GetSize()
	worldName = WorldNameWithRange(worldName, size.Width, size.Height, size.Length)
	result := ConvertResult{
		Path:      destPath,
		Width:     size.Width,
		Height:    size.Height,
		Length:    size.Length,
		WorldName: worldName,
	}
	c.publish(EventNameConvertParsed, structure.Name(), size)
	if err := structure.ToMCWorld(
		bedrockWorld,
		bwo_define.SubChunkPos{0, -4, 0},
		func(total int) { c.publish(EventNameConvertProgressInit, total) },
		func() { c.publish(EventNameConvertProgress) },
	); err != nil {
		_ = bedrockWorld.CloseWorld()
		return ConvertResult{}, err
	}

	bedrockWorld.LevelDat().LevelName = worldName
	if err := bedrockWorld.CloseWorld(); err != nil {
		return ConvertResult{}, err
	}
	if err := ZipDirectoryContents(worldDir, destPath); err != nil {
		return ConvertResult{}, err
	}

	c.publish(EventNameConvertSuccess, destPath)
	return result, nil
}

// WorldNameWithRange 生成带建筑范围后缀的世界名。
func WorldNameWithRange(name string, width, height, length int) string {
	if index := strings.Index(name, "@"); index >= 0 {
		name = name[:index]
	}
	if name == "" {
		name = "world"
	}
	return fmt.Sprintf("%s@[0,-64,0]~[%d,%d,%d]", name, width-1, height-64-1, length-1)
}

// DefaultWorldName 根据结构文件名生成默认世界名。
func DefaultWorldName(srcPath string) string {
	name := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	if index := strings.Index(name, "@"); index >= 0 {
		name = name[:index]
	}
	if name == "" {
		return "world"
	}
	return name
}

// IsZipFile 判断文件是否是 zip 格式。
func IsZipFile(path string) bool {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	_ = reader.Close()
	return true
}

// ZipDirectoryContents 把目录内容打包到 zip 根目录。
func ZipDirectoryContents(srcDir, destPath string) error {
	tmpPath := destPath + ".tmp"
	if err := os.RemoveAll(tmpPath); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	zipWriter := zip.NewWriter(out)
	walkErr := filepath.WalkDir(srcDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	closeErr := zipWriter.Close()
	fileCloseErr := out.Close()
	if walkErr != nil {
		return walkErr
	}
	if closeErr != nil {
		return closeErr
	}
	if fileCloseErr != nil {
		return fileCloseErr
	}

	if err := os.RemoveAll(destPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(tmpPath, destPath)
}

func (c *StructureConverter) publish(topic string, args ...any) {
	if c.eventBus != nil {
		c.eventBus.Publish(topic, args...)
	}
}
