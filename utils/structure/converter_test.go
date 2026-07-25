package structure

import "testing"

func TestWorldNameWithRange(t *testing.T) {
	got := WorldNameWithRange("建筑", 1457, 98, 1482)
	want := "建筑@[0,-64,0]~[1456,33,1481]"
	if got != want {
		t.Fatalf("WorldNameWithRange() = %q, want %q", got, want)
	}
}

func TestWorldNameWithRangeTrimsExistingRange(t *testing.T) {
	got := WorldNameWithRange("建筑@[1,2,3]~[4,5,6]", 2, 3, 4)
	want := "建筑@[0,-64,0]~[1,-62,3]"
	if got != want {
		t.Fatalf("WorldNameWithRange() = %q, want %q", got, want)
	}
}
