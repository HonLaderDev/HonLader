package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBuildFormEnterOnLastFieldReturnsToSections(t *testing.T) {
	m := newModel()
	m.page = pageBuildForm
	m.buildForm = newBuildForm()
	m.buildForm.section = buildSectionAdvanced
	m.buildForm.cursor = len(buildFormFields(buildSectionAdvanced)) - 1

	updated, _ := m.updateBuildForm(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(model)

	if got.page != pageBuildSections {
		t.Fatalf("page = %v, want %v", got.page, pageBuildSections)
	}
	if got.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", got.cursor)
	}
	if got.buildForm.cursor != 0 {
		t.Fatalf("buildForm.cursor = %d, want 0", got.buildForm.cursor)
	}
}
