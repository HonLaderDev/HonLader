package utils

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/HonLaderDev/HonLader/define"
)

// 正则：匹配三组正负整数，分隔符 空格/逗号/点
var posRegex = regexp.MustCompile(`(-?\d+)[ ,.]+(-?\d+)[ ,.]+(-?\d+)`)

// 正则：匹配文件名中的世界范围，例如 name@[0,0,0]~[170,320,220].mcworld
var worldRangeRegex = regexp.MustCompile(`@\[\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*\]\s*~\s*\[\s*(-?\d+)\s*,\s*(-?\d+)\s*,\s*(-?\d+)\s*\]`)

// ParseBlockPos 从字符串解析出 BlockPos([]int)
// str: 待解析字符串，如 "-10,10,10"、"-1 . 6 8"
func ParseBlockPos(str string) (define.BlockPos, error) {
	// 匹配正则
	match := posRegex.FindStringSubmatch(str)
	if len(match) != 4 {
		return define.BlockPos{}, errors.New("未匹配到合法的三组坐标数字")
	}

	// 截取三个数字字符串
	s1, s2, s3 := match[1], match[2], match[3]

	// 转 int
	x, err1 := strconv.Atoi(s1)
	y, err2 := strconv.Atoi(s2)
	z, err3 := strconv.Atoi(s3)
	if err1 != nil || err2 != nil || err3 != nil {
		return define.BlockPos{}, errors.New("数字转换失败，存在非法整数")
	}

	// 构造 BlockPos
	return define.BlockPos{x, y, z}, nil
}

// ParseWorldRange 从字符串解析世界范围坐标。
// 支持格式：VOH3-0主城@[0,0,0]~[170,320,220].mcworld。
func ParseWorldRange(str string) (start define.BlockPos, end define.BlockPos, found bool) {
	match := worldRangeRegex.FindStringSubmatch(str)
	if len(match) != 7 {
		return define.BlockPos{}, define.BlockPos{}, false
	}

	values := make([]int, 6)
	for i := range values {
		value, err := strconv.Atoi(match[i+1])
		if err != nil {
			return define.BlockPos{}, define.BlockPos{}, false
		}
		values[i] = value
	}

	return define.BlockPos{values[0], values[1], values[2]},
		define.BlockPos{values[3], values[4], values[5]},
		true
}
