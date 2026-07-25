package utils

import (
	"testing"

	"github.com/HonLaderDev/HonLader/define"
)

func TestParseWorldRange(t *testing.T) {
	start, end, found := ParseWorldRange("VOH3-0主城@[0,0,0]~[170,320,220].mcworld")
	if !found {
		t.Fatal("expected world range to be found")
	}
	if start != (define.BlockPos{0, 0, 0}) {
		t.Fatalf("unexpected start: %v", start)
	}
	if end != (define.BlockPos{170, 320, 220}) {
		t.Fatalf("unexpected end: %v", end)
	}
}

func TestParseWorldRangeWithSpacesAndNegativeValues(t *testing.T) {
	start, end, found := ParseWorldRange("name@[-1, 2, -3]~[ 4, -5, 6 ].mcworld")
	if !found {
		t.Fatal("expected world range to be found")
	}
	if start != (define.BlockPos{-1, 2, -3}) {
		t.Fatalf("unexpected start: %v", start)
	}
	if end != (define.BlockPos{4, -5, 6}) {
		t.Fatalf("unexpected end: %v", end)
	}
}

func TestParseWorldRangeNotFound(t *testing.T) {
	_, _, found := ParseWorldRange("VOH3-0主城.mcworld")
	if found {
		t.Fatal("expected world range not to be found")
	}
}
