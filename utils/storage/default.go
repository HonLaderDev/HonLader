package storage

import (
	"os"

	"github.com/adrg/xdg"
	"github.com/spf13/afero"
)

// DefaultStorage 使用本机操作系统文件系统和 XDG 目录规范。
type DefaultStorage struct {
	fs afero.Fs
}

// NewDefaultStorage 创建默认存储实现。
func NewDefaultStorage() *DefaultStorage {
	return &DefaultStorage{fs: afero.NewOsFs()}
}

// FileSystem 返回本机操作系统文件系统。
func (s *DefaultStorage) FileSystem() afero.Fs {
	return s.fs
}

// ConfigDir 返回配置目录。
func (s *DefaultStorage) ConfigDir() string {
	return xdg.ConfigHome
}

// DataDir 返回数据目录。
func (s *DefaultStorage) DataDir() string {
	return xdg.DataHome
}

// TmpDir 返回临时目录。
func (s *DefaultStorage) TmpDir() string {
	return os.TempDir()
}

// DownloadDir 返回下载目录。
func (s *DefaultStorage) DownloadDir() string {
	return xdg.UserDirs.Download
}
