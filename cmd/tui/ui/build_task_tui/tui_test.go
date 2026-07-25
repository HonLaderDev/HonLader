package build_task_tui

import "testing"

func TestDefaultConvertedMCWorldPath(t *testing.T) {
	got := DefaultConvertedMCWorldPath("/tmp/建筑@[0,0,0]~[1,2,3].bdx")
	want := "/tmp/建筑.mcworld"
	if got != want {
		t.Fatalf("DefaultConvertedMCWorldPath() = %q, want %q", got, want)
	}
}

func TestNormalizeConvertedMCWorldPathWithName(t *testing.T) {
	got := NormalizeConvertedMCWorldPath("/tmp/source.bdx", "target")
	want := "/tmp/target.mcworld"
	if got != want {
		t.Fatalf("NormalizeConvertedMCWorldPath() = %q, want %q", got, want)
	}
}

func TestNormalizeConvertedMCWorldPathWithPath(t *testing.T) {
	got := NormalizeConvertedMCWorldPath("/tmp/source.bdx", "/data/target")
	want := "/data/target.mcworld"
	if got != want {
		t.Fatalf("NormalizeConvertedMCWorldPath() = %q, want %q", got, want)
	}
}
