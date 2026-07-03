package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) body(content string, height int) string {
	if m.isLandscape() {
		return m.landscapeBody(content, height)
	}
	return m.portraitBody(content, height)
}

func (m model) isLandscape() bool {
	return m.width >= logoWidth()+4+62 && m.width > m.height*2
}

func (m model) portraitBody(content string, height int) string {
	logoBlock := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logoStyle.Render(logo))
	logoY := height / 6
	contentY := max(logoY+lipgloss.Height(logoBlock)+1, (height-lipgloss.Height(content))/2)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Repeat("\n", logoY),
		logoBlock,
		strings.Repeat("\n", max(0, contentY-logoY-lipgloss.Height(logoBlock))),
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content),
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center, m.detailPanel()),
	)
	return fitHeight(body, height)
}

func (m model) landscapeBody(content string, height int) string {
	leftWidth := logoWidth()
	rightWidth := m.width - leftWidth - 4

	left := lipgloss.NewStyle().Width(leftWidth).Render(lipgloss.JoinVertical(
		lipgloss.Left,
		logoStyle.Render(logo),
		"",
		m.detailPanel(),
	))
	right := lipgloss.NewStyle().Width(rightWidth).Render(lipgloss.PlaceHorizontal(rightWidth, lipgloss.Center, content))

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		fitHeight(left, height),
		strings.Repeat(" ", 4),
		fitHeight(right, height),
	)
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, row)
}

func logoWidth() int {
	width := 0
	for _, line := range strings.Split(logo, "\n") {
		width = max(width, lipgloss.Width(line))
	}
	return width
}

func fitHeight(content string, height int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}
