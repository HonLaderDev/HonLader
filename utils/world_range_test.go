package utils

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/HonLaderDev/HonLader/define"
)

func TestParseWorldRangeFromZipLevelName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "world.mcworld")
	writeTestZip(t, path, map[string]string{
		"db/CURRENT":    "ignored",
		"levelname.txt": "VOH3-0主城@[0,0,0]~[170,320,220]",
	})

	start, end, found, err := ParseWorldRangeFromZipLevelName(path)
	if err != nil {
		t.Fatalf("ParseWorldRangeFromZipLevelName() error = %v", err)
	}
	if !found {
		t.Fatal("ParseWorldRangeFromZipLevelName() found = false")
	}
	if start != (define.BlockPos{0, 0, 0}) {
		t.Fatalf("start = %#v", start)
	}
	if end != (define.BlockPos{170, 320, 220}) {
		t.Fatalf("end = %#v", end)
	}
}

func TestParseWorldRangeFromZipLevelNameInDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "world.mcworld")
	writeTestZip(t, path, map[string]string{
		"world/levelname.txt": "name@[-1,-2,-3]~[4,5,6]",
	})

	start, end, found, err := ParseWorldRangeFromZipLevelName(path)
	if err != nil {
		t.Fatalf("ParseWorldRangeFromZipLevelName() error = %v", err)
	}
	if !found {
		t.Fatal("ParseWorldRangeFromZipLevelName() found = false")
	}
	if start != (define.BlockPos{-1, -2, -3}) {
		t.Fatalf("start = %#v", start)
	}
	if end != (define.BlockPos{4, 5, 6}) {
		t.Fatalf("end = %#v", end)
	}
}

func writeTestZip(t *testing.T, path string, files map[string]string) {
	t.Helper()

	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()

	writer := zip.NewWriter(out)
	defer writer.Close()

	for name, content := range files {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = file.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
}
