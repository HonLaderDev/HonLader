package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/HonLaderDev/HonLader/define"
)

// ParseWorldRangeFromZipLevelName 从 zip 文件中的 levelname.txt 解析世界范围坐标。
func ParseWorldRangeFromZipLevelName(path string) (start define.BlockPos, end define.BlockPos, found bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return define.BlockPos{}, define.BlockPos{}, false, err
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return define.BlockPos{}, define.BlockPos{}, false, err
	}

	for _, file := range reader.File {
		if !IsLevelNameFile(file.Name) {
			continue
		}
		content, err := ReadZipTextFile(file)
		if err != nil {
			return define.BlockPos{}, define.BlockPos{}, false, err
		}
		start, end, found := ParseWorldRange(content)
		return start, end, found, nil
	}
	return define.BlockPos{}, define.BlockPos{}, false, nil
}

// IsLevelNameFile 判断 zip 条目是否是 levelname.txt。
func IsLevelNameFile(name string) bool {
	return strings.EqualFold(filepath.Base(name), "levelname.txt")
}

// ReadZipTextFile 读取 zip 文本文件内容。
func ReadZipTextFile(file *zip.File) (string, error) {
	if file == nil {
		return "", fmt.Errorf("zip 文件条目不能为空")
	}
	rc, err := file.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
