package data

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/RedLaderDev/Fatalder/utils"
	"github.com/spf13/afero"
)

const jsonFileSuffix = ".json"

const fileTimeLayout = "20060102T150405.000000000"

func jsonFilePath(dir, fileName string) (string, error) {
	if fileName == "" {
		return "", fmt.Errorf("empty file name")
	}
	if filepath.Base(fileName) != fileName {
		return "", fmt.Errorf("invalid file name %q", fileName)
	}
	return filepath.Join(dir, fileName+jsonFileSuffix), nil
}

func configFileName(createdAt time.Time) string {
	return createdAt.UTC().Format(fileTimeLayout)
}

func saveJSON(fs afero.Fs, dir string, createdAt time.Time, value any) error {
	fileName := configFileName(createdAt)
	filePath, err := jsonFilePath(dir, fileName)
	if err != nil {
		return fmt.Errorf("saveJSON: %w", err)
	}
	if err := fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("saveJSON: mkdir %q: %w", dir, err)
	}

	encoded, err := utils.MarshalMap(value)
	if err != nil {
		return fmt.Errorf("saveJSON: encode %q: %w", fileName, err)
	}
	data, err := json.MarshalIndent(encoded, "", "  ")
	if err != nil {
		return fmt.Errorf("saveJSON: marshal %q: %w", fileName, err)
	}

	tmpPath := filePath + ".tmp"
	if err := afero.WriteFile(fs, tmpPath, data, 0644); err != nil {
		return fmt.Errorf("saveJSON: write temp %q: %w", tmpPath, err)
	}
	if err := fs.Rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("saveJSON: rename %q to %q: %w", tmpPath, filePath, err)
	}
	return nil
}

func loadJSONFile(fs afero.Fs, dir, fileName string, value any) (bool, error) {
	filePath, err := jsonFilePath(dir, fileName)
	if err != nil {
		return false, fmt.Errorf("loadJSON: %w", err)
	}
	exists, err := afero.Exists(fs, filePath)
	if err != nil {
		return false, fmt.Errorf("loadJSON: stat %q: %w", filePath, err)
	}
	if !exists {
		return false, nil
	}

	data, err := afero.ReadFile(fs, filePath)
	if err != nil {
		return false, fmt.Errorf("loadJSON: read %q: %w", filePath, err)
	}
	var encoded map[string]any
	if err := json.Unmarshal(data, &encoded); err != nil {
		return false, fmt.Errorf("loadJSON: unmarshal %q: %w", filePath, err)
	}
	if err := utils.UnmarshalMap(encoded, value); err != nil {
		return false, fmt.Errorf("loadJSON: decode %q: %w", filePath, err)
	}
	return true, nil
}

func loadJSONByName[T any](fs afero.Fs, dir, name string, itemName func(T) string) (T, bool, error) {
	var zero T
	names, err := listJSONNames(fs, dir)
	if err != nil {
		return zero, false, fmt.Errorf("loadJSONByName: %w", err)
	}
	for _, fileName := range names {
		var item T
		exists, err := loadJSONFile(fs, dir, fileName, &item)
		if err != nil {
			return zero, false, fmt.Errorf("loadJSONByName: load %q: %w", fileName, err)
		}
		if exists && itemName(item) == name {
			return item, true, nil
		}
	}
	return zero, false, nil
}

func deleteJSONByName[T any](fs afero.Fs, dir, name string, itemName func(T) string) (bool, error) {
	names, err := listJSONNames(fs, dir)
	if err != nil {
		return false, fmt.Errorf("deleteJSONByName: %w", err)
	}
	for _, fileName := range names {
		var item T
		exists, err := loadJSONFile(fs, dir, fileName, &item)
		if err != nil {
			return false, fmt.Errorf("deleteJSONByName: load %q: %w", fileName, err)
		}
		if !exists || itemName(item) != name {
			continue
		}
		filePath, err := jsonFilePath(dir, fileName)
		if err != nil {
			return false, fmt.Errorf("deleteJSONByName: path %q: %w", fileName, err)
		}
		if err := fs.Remove(filePath); err != nil {
			return false, fmt.Errorf("deleteJSONByName: remove %q: %w", filePath, err)
		}
		return true, nil
	}
	return false, nil
}

func listJSONNames(fs afero.Fs, dir string) ([]string, error) {
	entries, err := afero.ReadDir(fs, dir)
	if err != nil {
		exists, existsErr := afero.DirExists(fs, dir)
		if existsErr != nil {
			return nil, fmt.Errorf("listJSONNames: stat %q: %w", dir, existsErr)
		}
		if !exists {
			return []string{}, nil
		}
		return nil, fmt.Errorf("listJSONNames: read dir %q: %w", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name, ok := strings.CutSuffix(entry.Name(), jsonFileSuffix)
		if ok {
			names = append(names, name)
		}
	}
	return names, nil
}
