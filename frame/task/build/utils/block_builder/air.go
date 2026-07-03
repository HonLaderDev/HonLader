package block_builder

import (
	"fmt"

	"github.com/RedLaderDev/RedLader/define"
	build_utils "github.com/RedLaderDev/RedLader/frame/task/build/utils"
	"github.com/RedLaderDev/RedLader/frame/task/build/utils/chunk_fill"
)

// BuildAirCommands 生成把指定区域完整填充为空气的最少 fill 命令。
func (b *BlockBuilder) BuildAirCommands(startPos, endPos define.BlockPos) []string {
	start, end := startPos.SortStartAndEndPos(endPos)
	width := end.X() - start.X() + 1
	height := end.Y() - start.Y() + 1
	depth := end.Z() - start.Z() + 1
	if width <= 0 || height <= 0 || depth <= 0 {
		return nil
	}

	plan := bestAirFillPlan(width, height, depth)
	commands := make([]string, 0, plan.xParts*plan.yParts*plan.zParts)
	for x := 0; x < width; x += plan.xSize {
		xSize := build_utils.MinInt(plan.xSize, width-x)
		for y := 0; y < height; y += plan.ySize {
			ySize := build_utils.MinInt(plan.ySize, height-y)
			for z := 0; z < depth; z += plan.zSize {
				zSize := build_utils.MinInt(plan.zSize, depth-z)
				commands = append(commands, airFillCommand(
					start.X()+x,
					start.Y()+y,
					start.Z()+z,
					start.X()+x+xSize-1,
					start.Y()+y+ySize-1,
					start.Z()+z+zSize-1,
				))
			}
		}
	}
	return commands
}

type airFillPlan struct {
	xParts int
	yParts int
	zParts int
	xSize  int
	ySize  int
	zSize  int
}

func bestAirFillPlan(width, height, depth int) airFillPlan {
	best := airFillPlan{
		xParts: width,
		yParts: height,
		zParts: depth,
		xSize:  1,
		ySize:  1,
		zSize:  1,
	}
	bestCount := width * height * depth

	for xParts := 1; xParts <= width; xParts++ {
		xSize := build_utils.CeilQuotient(width, xParts)
		for yParts := 1; yParts <= height; yParts++ {
			ySize := build_utils.CeilQuotient(height, yParts)
			zSizeLimit := chunk_fill.MaxFillVolume / (xSize * ySize)
			if zSizeLimit <= 0 {
				continue
			}
			zParts := build_utils.CeilQuotient(depth, zSizeLimit)
			if zParts <= 0 {
				continue
			}
			count := xParts * yParts * zParts
			if count > bestCount {
				continue
			}
			zSize := build_utils.CeilQuotient(depth, zParts)
			if xSize*ySize*zSize > chunk_fill.MaxFillVolume {
				continue
			}
			if count < bestCount || betterAirFillPlan(xParts, yParts, zParts, best) {
				best = airFillPlan{
					xParts: xParts,
					yParts: yParts,
					zParts: zParts,
					xSize:  xSize,
					ySize:  ySize,
					zSize:  zSize,
				}
				bestCount = count
			}
		}
	}
	return best
}

func betterAirFillPlan(xParts, yParts, zParts int, current airFillPlan) bool {
	if yParts != current.yParts {
		return yParts < current.yParts
	}
	if xParts != current.xParts {
		return xParts < current.xParts
	}
	return zParts < current.zParts
}

func airFillCommand(startX, startY, startZ, endX, endY, endZ int) string {
	return fmt.Sprintf(
		"fill %d %d %d %d %d %d minecraft:air []\n",
		startX,
		startY,
		startZ,
		endX,
		endY,
		endZ,
	)
}
