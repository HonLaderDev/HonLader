package utils

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/HonLaderDev/HonLader/define"
)

// 正则：匹配三组正负整数，分隔符 空格/逗号/点
var posRegex = regexp.MustCompile(`(-?\d+)[ ,.]+(-?\d+)[ ,.]+(-?\d+)`)

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
