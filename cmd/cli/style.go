package main

import "github.com/charmbracelet/lipgloss"

const logo = `██████╗ ███████╗██████╗ ██╗      █████╗ ██████╗ ███████╗██████╗
██╔══██╗██╔════╝██╔══██╗██║     ██╔══██╗██╔══██╗██╔════╝██╔══██╗
██████╔╝█████╗  ██║  ██║██║     ███████║██║  ██║█████╗  ██████╔╝
██╔══██╗██╔══╝  ██║  ██║██║     ██╔══██║██║  ██║██╔══╝  ██╔══██╗
██║  ██║███████╗██████╔╝███████╗██║  ██║██████╔╝███████╗██║  ██║
╚═╝  ╚═╝╚══════╝╚═════╝ ╚══════╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═╝`

const (
	panelWidth        = 58
	panelContentWidth = panelWidth - 6
)

var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Padding(0, 2)
	headerMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("236")).
			Padding(0, 2)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))
	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("43")).
			Padding(1, 2).
			Width(panelWidth)
	detailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("36")).
				Padding(1, 2).
				Width(58)
	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 1)
	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("30")).
				Padding(0, 1)
	indexStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("50")).
			Bold(true)
	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("43")).
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(1, 3).
			Width(46)
	dialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("30")).
			Padding(0, 1)
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))
)
