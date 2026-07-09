package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/HonLaderDev/HonLader/define"
	"github.com/HonLaderDev/HonLader/utils/data"
	tea "github.com/charmbracelet/bubbletea"
)

type page int

const (
	pageMain page = iota
	pageMore
	pageServers
	pageServerActions
	pageServerForm
	pageBuildServers
	pageBuildBasicForm
	pageBuildSections
	pageBuildForm
)

type action int

const (
	actionBuild action = iota
	actionExport
	actionMore
	actionExit
	actionTaskGroups
	actionServers
	actionCreateServer
	actionViewServer
	actionEditServer
	actionDeleteServer
	actionBackMain
	actionBackMore
	actionBackServers
	actionSelectBuildServer
	actionOpenBuildSection
	actionSaveBuild
	actionBackBuildServers
)

type menuItem struct {
	title        string
	detail       string
	action       action
	buildSection buildSection
}

type model struct {
	page             page
	cursor           int
	dialog           string
	buildMoreConfirm bool
	buildMoreCursor  int
	width            int
	height           int
	dataManager      define.DataManager
	servers          []define.ServerConfig
	serverForm       serverForm
	serverIndex      int
	buildForm        buildForm
	buildServer      *define.ServerConfig
}

type serverForm struct {
	cursor  int
	values  [3]string
	edit    bool
	oldName string
}

type buildForm struct {
	cursor  int
	section buildSection
	values  map[string]string
}

func newModel() model {
	m := model{
		page:        pageMain,
		dataManager: data.NewDataManager(nil),
	}
	m.reloadServers()
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		return m.updateKey(msg)
	}
	return m, nil
}

func (m model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.buildMoreConfirm {
		return m.updateBuildMoreConfirm(msg)
	}
	if m.dialog != "" {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "backspace", "enter", " ", "q":
			m.dialog = ""
		}
		return m, nil
	}
	if m.page == pageServerForm {
		return m.updateServerForm(msg)
	}
	if m.page == pageBuildBasicForm || m.page == pageBuildForm {
		return m.updateBuildForm(msg)
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "backspace":
		m.back()
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "enter", " ":
		return m.selectItem()
	}
	return m, nil
}

func (m *model) back() {
	switch m.page {
	case pageBuildForm:
		m.page = pageBuildSections
	case pageBuildBasicForm:
		m.page = pageBuildServers
	case pageBuildSections:
		m.page = pageBuildServers
	case pageBuildServers:
		m.page = pageMain
	case pageServerForm:
		if m.serverForm.edit {
			m.page = pageServerActions
		} else {
			m.page = pageServers
		}
	case pageServerActions:
		m.page = pageServers
	case pageServers:
		m.page = pageMore
	case pageMore:
		m.page = pageMain
	default:
		return
	}
	m.cursor = 0
	m.dialog = ""
}

func (m *model) moveCursor(step int) {
	items := m.items()
	if len(items) == 0 {
		return
	}
	m.cursor = (m.cursor + step + len(items)) % len(items)
}

func (m model) selectItem() (tea.Model, tea.Cmd) {
	items := m.items()
	if len(items) == 0 {
		return m, nil
	}

	switch items[m.cursor].action {
	case actionBuild:
		m.page = pageBuildServers
		m.cursor = 0
		m.dialog = ""
		m.reloadServers()
	case actionExport:
		m.dialog = "导出建筑暂时不支持。"
	case actionMore:
		m.page = pageMore
		m.cursor = 0
		m.dialog = ""
	case actionExit:
		return m, tea.Quit
	case actionTaskGroups:
		m.dialog = "任务组配置功能入口已预留。"
	case actionServers:
		m.page = pageServers
		m.cursor = 0
		m.dialog = ""
		m.reloadServers()
	case actionCreateServer:
		m.startServerForm()
	case actionViewServer:
		if m.page == pageServerActions {
			m.dialog = m.selectedServerDetail()
		} else {
			m.openServerActions()
		}
	case actionEditServer:
		m.startEditServerForm()
	case actionDeleteServer:
		m.deleteSelectedServer()
	case actionBackMain:
		m.page = pageMain
		m.cursor = 0
		m.dialog = ""
		m.buildServer = nil
		m.buildForm = buildForm{}
	case actionBackMore:
		m.page = pageMore
		m.cursor = 0
		m.dialog = ""
	case actionBackServers:
		m.page = pageServers
		m.cursor = 0
		m.dialog = ""
	case actionSelectBuildServer:
		m.selectBuildServer()
	case actionOpenBuildSection:
		m.openBuildSection(items[m.cursor].buildSection)
	case actionSaveBuild:
		m.saveBuildForm()
	case actionBackBuildServers:
		m.page = pageBuildServers
		m.cursor = 0
	default:
		return m, nil
	}
	return m, nil
}

func (m *model) selectBuildServer() {
	if m.cursor < 0 || m.cursor >= len(m.servers) {
		return
	}
	server := m.servers[m.cursor]
	m.buildServer = &server
	m.page = pageBuildBasicForm
	m.cursor = 0
	m.dialog = ""
	m.buildMoreConfirm = false
	m.buildMoreCursor = 0
	m.buildForm = newBuildForm()
}

func (m *model) openBuildSection(section buildSection) {
	m.page = pageBuildForm
	m.cursor = 0
	m.dialog = ""
	m.buildForm.section = section
	m.buildForm.cursor = 0
}

func (m *model) reloadServers() {
	if m.dataManager == nil {
		return
	}
	configs, err := m.dataManager.ListServerConfigs()
	if err != nil {
		m.dialog = fmt.Sprintf("读取服务器配置失败：%v", err)
		return
	}
	m.servers = m.servers[:0]
	for _, config := range configs {
		m.servers = append(m.servers, config)
	}
	sort.Slice(m.servers, func(i int, j int) bool {
		return m.servers[i].CreatedAt.Before(m.servers[j].CreatedAt)
	})
}

func (m *model) openServerActions() {
	index := m.cursor - 1
	if index < 0 || index >= len(m.servers) {
		return
	}
	m.serverIndex = index
	m.page = pageServerActions
	m.cursor = 0
	m.dialog = ""
}

func (m *model) startServerForm() {
	m.page = pageServerForm
	m.cursor = 0
	m.dialog = ""
	m.serverForm = serverForm{}
}

func (m *model) startEditServerForm() {
	if m.serverIndex < 0 || m.serverIndex >= len(m.servers) {
		m.dialog = "服务器配置不存在。"
		return
	}
	server := m.servers[m.serverIndex]
	m.page = pageServerForm
	m.cursor = 0
	m.dialog = ""
	m.serverForm = serverForm{edit: true, oldName: server.Name}
	m.serverForm.values[0] = server.Name
	m.serverForm.values[1] = server.ServerCode
	m.serverForm.values[2] = server.ServerPassword
}

func (m model) updateServerForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.serverForm.edit {
			m.page = pageServerActions
		} else {
			m.page = pageServers
		}
		m.cursor = 0
	case "up":
		m.serverForm.cursor = max(0, m.serverForm.cursor-1)
	case "down":
		m.serverForm.cursor = min(len(m.serverForm.values)-1, m.serverForm.cursor+1)
	case "enter":
		if m.serverForm.cursor == len(m.serverForm.values)-1 {
			m.saveServerForm()
		} else {
			m.serverForm.cursor++
		}
	case "backspace":
		value := m.serverForm.values[m.serverForm.cursor]
		if value != "" {
			runes := []rune(value)
			m.serverForm.values[m.serverForm.cursor] = string(runes[:len(runes)-1])
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.serverForm.values[m.serverForm.cursor] += msg.String()
		}
	}
	return m, nil
}

func (m model) updateBuildForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	fields := m.currentBuildFields()
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.page == pageBuildBasicForm {
			m.page = pageBuildServers
		} else {
			m.page = pageBuildSections
		}
		m.cursor = 0
	case "up":
		m.buildForm.cursor = max(0, m.buildForm.cursor-1)
	case "down":
		m.buildForm.cursor = min(len(fields)-1, m.buildForm.cursor+1)
	case "enter":
		field := m.currentBuildField()
		if field.kind == buildFieldBool && m.buildForm.cursor != len(fields)-1 {
			m.toggleBuildBool(field.key)
		}
		if m.buildForm.cursor == len(fields)-1 {
			m.finishBuildFormPage()
		} else {
			m.buildForm.cursor = min(len(fields)-1, m.buildForm.cursor+1)
		}
	case " ":
		field := m.currentBuildField()
		if field.kind == buildFieldBool {
			m.toggleBuildBool(field.key)
		}
	case "backspace":
		field := m.currentBuildField()
		if field.kind != buildFieldText {
			return m, nil
		}
		value := m.buildForm.values[field.key]
		if value != "" {
			runes := []rune(value)
			m.buildForm.values[field.key] = string(runes[:len(runes)-1])
		}
	default:
		field := m.currentBuildField()
		if msg.Type == tea.KeyRunes && field.kind == buildFieldText {
			m.buildForm.values[field.key] += msg.String()
		}
	}
	return m, nil
}

func (m *model) finishBuildFormPage() {
	if m.page == pageBuildBasicForm {
		m.finishBuildBasicForm()
		return
	}
	m.page = pageBuildSections
	m.cursor = 0
	m.buildForm.cursor = 0
}

func (m model) currentBuildFields() []buildFormField {
	if m.page == pageBuildBasicForm {
		return buildBasicFields()
	}
	return buildFormFields(m.buildForm.section)
}

func (m model) currentBuildField() buildFormField {
	fields := m.currentBuildFields()
	if m.buildForm.cursor < 0 || m.buildForm.cursor >= len(fields) {
		return fields[0]
	}
	return fields[m.buildForm.cursor]
}

func (m model) updateBuildMoreConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "backspace":
		m.buildMoreConfirm = false
		m.page = pageBuildBasicForm
		m.buildForm.cursor = len(buildBasicFields()) - 1
	case "left", "up", "h", "k":
		m.buildMoreCursor = max(0, m.buildMoreCursor-1)
	case "right", "down", "l", "j":
		m.buildMoreCursor = min(1, m.buildMoreCursor+1)
	case "enter", " ":
		m.buildMoreConfirm = false
		if m.buildMoreCursor == 0 {
			m.page = pageBuildSections
			m.cursor = 0
			m.buildForm.cursor = 0
			return m, nil
		}
		m.saveBuildForm()
	}
	return m, nil
}

func (m *model) saveServerForm() {
	name := strings.TrimSpace(m.serverForm.values[0])
	if name == "" {
		m.dialog = "服务器配置名称不能为空。"
		return
	}
	config := define.ServerConfig{
		Name:           name,
		ServerCode:     strings.TrimSpace(m.serverForm.values[1]),
		ServerPassword: m.serverForm.values[2],
	}
	if err := m.dataManager.SaveServerConfig(config); err != nil {
		m.dialog = fmt.Sprintf("保存服务器配置失败：%v", err)
		return
	}
	if m.serverForm.edit && name != m.serverForm.oldName {
		if _, err := m.dataManager.DeleteServerConfig(m.serverForm.oldName); err != nil {
			m.dialog = fmt.Sprintf("删除旧服务器配置失败：%v", err)
			return
		}
	}
	m.reloadServers()
	m.cursor = m.serverCursorByName(name)
	m.page = pageServers
	m.serverForm = serverForm{}
	m.dialog = "已保存服务器配置：" + name
}

func (m *model) deleteSelectedServer() {
	if m.serverIndex < 0 || m.serverIndex >= len(m.servers) {
		m.dialog = "服务器配置不存在。"
		return
	}
	name := m.servers[m.serverIndex].Name
	deleted, err := m.dataManager.DeleteServerConfig(name)
	if err != nil {
		m.dialog = fmt.Sprintf("删除服务器配置失败：%v", err)
		return
	}
	m.reloadServers()
	m.page = pageServers
	m.cursor = 0
	if deleted {
		m.dialog = "已删除服务器配置：" + name
	} else {
		m.dialog = "服务器配置不存在：" + name
	}
}

func (m model) serverCursorByName(name string) int {
	for i, server := range m.servers {
		if server.Name == name {
			return i + 1
		}
	}
	return 0
}
