package main

import (
	"fmt"
	"strings"

	"github.com/HonLaderDev/HonLader/consts"
	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	content := appStyle.Render(m.mainPanel())
	if m.width <= 0 || m.height <= 0 {
		return content
	}

	footer := m.footer()
	bodyHeight := max(1, m.height-lipgloss.Height(footer))
	body := m.body(content, bodyHeight)
	if m.dialog != "" {
		body = lipgloss.Place(m.width, bodyHeight, lipgloss.Center, lipgloss.Center, m.dialogPanel())
	}
	if m.buildMoreConfirm {
		body = lipgloss.Place(m.width, bodyHeight, lipgloss.Center, lipgloss.Center, m.buildMoreConfirmPanel())
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, footer)
}

func (m model) mainPanel() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.topBar(), m.menuPanel())
}

func (m model) topBar() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		headerStyle.Render(consts.Name),
		headerMetaStyle.Render("v"+consts.Version),
	)
}

func (m model) menuPanel() string {
	if m.page == pageServerForm {
		return m.serverFormPanel()
	}
	if m.page == pageBuildBasicForm || m.page == pageBuildForm {
		return m.buildFormPanel()
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("当前页面 - " + m.title()))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(m.subtitle()))
	b.WriteString("\n\n")

	for i, item := range m.items() {
		cursor := "  "
		if m.cursor == i {
			cursor = "▸ "
		}
		if m.cursor == i {
			line := fmt.Sprintf("%s%02d  %s", cursor, i+1, item.title)
			line = selectedItemStyle.Render(line)
			b.WriteString(line)
		} else {
			line := fmt.Sprintf("%s%s  %s", cursor, indexStyle.Render(fmt.Sprintf("%02d", i+1)), item.title)
			line = itemStyle.Width(panelContentWidth).Render(line)
			b.WriteString(line)
		}
		b.WriteString("\n\n")
	}
	return panelStyle.Render(b.String())
}

func (m model) buildFormPanel() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("当前页面 - " + m.title()))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(m.subtitle()))
	b.WriteString("\n\n")

	fields := m.currentBuildFields()
	for i, field := range fields {
		cursor := "  "
		if m.buildForm.cursor == i {
			cursor = "▸ "
		}
		value := buildFormValue(m.buildForm, field)
		if value == "" && field.kind == buildFieldText {
			value = "未填写"
		}
		line := fmt.Sprintf("%s%s：%s", cursor, field.name, value)
		if m.buildForm.cursor == i {
			line = selectedItemStyle.Render(line)
		} else {
			line = itemStyle.Width(panelContentWidth).Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n\n")
	}
	return panelStyle.Render(b.String())
}

func buildFormValue(form buildForm, field buildFormField) string {
	switch field.kind {
	case buildFieldBool:
		if form.values[field.key] == "true" {
			return "开"
		}
		return "关"
	default:
		return form.values[field.key]
	}
}

func (m model) serverFormPanel() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("当前页面 - " + m.title()))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(m.subtitle()))
	b.WriteString("\n\n")

	for i, label := range serverFormLabels() {
		cursor := "  "
		if m.serverForm.cursor == i {
			cursor = "▸ "
		}
		value := m.serverForm.values[i]
		if label.secret && value != "" {
			value = strings.Repeat("*", len([]rune(value)))
		}
		if value == "" {
			value = "未填写"
		}
		line := fmt.Sprintf("%s%s：%s", cursor, label.name, value)
		if m.serverForm.cursor == i {
			line = selectedItemStyle.Render(line)
		} else {
			line = itemStyle.Width(panelContentWidth).Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n\n")
	}
	return panelStyle.Render(b.String())
}

func (m model) detailPanel() string {
	item, ok := m.selectedItem()
	if !ok || item.detail == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("当前选项 - " + item.title))
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render(item.detail))
	return detailPanelStyle.Render(b.String())
}

func (m model) dialogPanel() string {
	var b strings.Builder
	b.WriteString(dialogTitleStyle.Render("提示"))
	b.WriteString("\n\n")
	b.WriteString(m.dialog)
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Enter/Esc 关闭"))
	return dialogStyle.Render(b.String())
}

func (m model) buildMoreConfirmPanel() string {
	options := []string{"配置更多", "直接保存"}
	var b strings.Builder
	b.WriteString(dialogTitleStyle.Render("是否配置更多"))
	b.WriteString("\n\n")
	b.WriteString("基础配置已填写完成，是否继续调整其他板块？")
	b.WriteString("\n\n")
	for i, option := range options {
		text := "  " + option
		if m.buildMoreCursor == i {
			text = selectedItemStyle.Render("▸ " + option)
		} else {
			text = itemStyle.Render(text)
		}
		b.WriteString(text)
		if i != len(options)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("←/→ 切换  Enter 确认  Esc 返回"))
	return dialogStyle.Render(b.String())
}

func (m model) footer() string {
	help := helpStyle.Width(m.width).Align(lipgloss.Center).Render("↑/↓ 或 j/k 选择  Enter 确认  Esc 返回  q 退出")
	status := statusStyle.Width(m.width).Render(m.statusText())
	return lipgloss.JoinVertical(lipgloss.Left, help, status)
}
