package model

import "testing"

func TestNextPrevFileWrapsVisible(t *testing.T) {
	m := &model{visible: []int{2, 4, 7}, activeFile: 2}
	m.nextFile()
	if m.activeFile != 4 {
		t.Fatalf("next active file = %d, want 4", m.activeFile)
	}
	m.prevFile()
	if m.activeFile != 2 {
		t.Fatalf("prev active file = %d, want 2", m.activeFile)
	}
	m.prevFile()
	if m.activeFile != 7 {
		t.Fatalf("prev wrap active file = %d, want 7", m.activeFile)
	}
}
