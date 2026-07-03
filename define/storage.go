package define

import "github.com/spf13/afero"

// Storage 提供统一的文件系统入口和常用用户目录。
type Storage interface {
	// FileSystem 返回底层文件系统实现。
	FileSystem() afero.Fs
	// ConfigDir 返回配置目录。
	ConfigDir() string
	// DataDir 返回数据目录。
	DataDir() string
	// TmpDir 返回临时目录。
	TmpDir() string
	// DownloadDir 返回下载目录。
	DownloadDir() string
}
