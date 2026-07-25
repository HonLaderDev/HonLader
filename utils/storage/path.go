package storage

import "github.com/spf13/afero"

// PathStorage 使用调用方传入的绝对路径作为底层目录。
type PathStorage struct {
	fs          afero.Fs
	configDir   string
	dataDir     string
	tmpDir      string
	downloadDir string
}

// NewPathStorage 创建基于固定目录的 Storage。
func NewPathStorage(configDir, dataDir, tmpDir, downloadDir string) *PathStorage {
	return &PathStorage{
		fs:          afero.NewOsFs(),
		configDir:   configDir,
		dataDir:     dataDir,
		tmpDir:      tmpDir,
		downloadDir: downloadDir,
	}
}

func (s *PathStorage) FileSystem() afero.Fs { return s.fs }

func (s *PathStorage) ConfigDir() string { return s.configDir }

func (s *PathStorage) DataDir() string { return s.dataDir }

func (s *PathStorage) TmpDir() string { return s.tmpDir }

func (s *PathStorage) DownloadDir() string { return s.downloadDir }
