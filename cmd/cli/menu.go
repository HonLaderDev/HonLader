package main

import (
	"fmt"

	"github.com/RedLaderDev/Fatalder/define"
)

func (m model) title() string {
	switch m.page {
	case pageMore:
		return "更多功能"
	case pageServers:
		return "服务器配置"
	case pageServerActions:
		return "服务器配置 - " + m.selectedServerName()
	case pageServerForm:
		if m.serverForm.edit {
			return "编辑服务器配置"
		}
		return "新建服务器配置"
	case pageBuildServers:
		return "构建建筑 - 选择服务器"
	case pageBuildBasicForm:
		return "构建建筑 - 基础配置"
	case pageBuildSections:
		return "构建建筑 - 配置板块"
	case pageBuildForm:
		return "构建建筑 - " + buildSectionTitle(m.buildForm.section)
	default:
		return "主菜单"
	}
}

func (m model) subtitle() string {
	switch m.page {
	case pageMore:
		return "配置、维护和扩展入口。"
	case pageServers:
		return "查看服务器配置，或创建一份新的默认配置。"
	case pageServerActions:
		return "查看、编辑或删除当前服务器配置。"
	case pageServerForm:
		return "填写服务器连接参数，Enter 切换下一项并在最后一项保存。"
	case pageBuildServers:
		return "选择用于构建的服务器配置。"
	case pageBuildBasicForm:
		return "填写任务名、源世界区域、目标起点、目标维度和构建速度。"
	case pageBuildSections:
		return "按板块配置构建任务，完成后选择保存。"
	case pageBuildForm:
		return "文本项直接输入，Enter 下一项或完成，开关项按空格切换。"
	default:
		return "选择一个操作开始处理建筑任务。"
	}
}

func (m model) statusText() string {
	if m.page == pageServerForm {
		return fmt.Sprintf("状态: 编辑服务器配置  字段: %d/%d", m.serverForm.cursor+1, len(m.serverForm.values))
	}
	if m.page == pageBuildBasicForm || m.page == pageBuildForm {
		return fmt.Sprintf("状态: 编辑构建任务  字段: %d/%d", m.buildForm.cursor+1, len(m.currentBuildFields()))
	}
	return fmt.Sprintf("状态: 就绪  页面: %s  选项: %d/%d", m.title(), m.cursor+1, len(m.items()))
}

func (m model) selectedItem() (menuItem, bool) {
	items := m.items()
	if m.cursor < 0 || m.cursor >= len(items) {
		return menuItem{}, false
	}
	return items[m.cursor], true
}

func (m model) items() []menuItem {
	switch m.page {
	case pageBuildSections:
		items := make([]menuItem, 0, len(buildSections())+2)
		for _, section := range buildSections() {
			items = append(items, menuItem{
				title:        section.title,
				detail:       section.detail,
				action:       actionOpenBuildSection,
				buildSection: section.value,
			})
		}
		items = append(items,
			menuItem{title: "保存任务配置", detail: "保存当前构建任务组配置。", action: actionSaveBuild},
			menuItem{title: "返回选择服务器", detail: "重新选择用于构建的服务器配置。", action: actionBackBuildServers},
		)
		return items
	case pageBuildServers:
		items := make([]menuItem, 0, len(m.servers)+1)
		for _, server := range m.servers {
			items = append(items, menuItem{
				title:  server.Name,
				detail: serverDetail(server),
				action: actionSelectBuildServer,
			})
		}
		items = append(items, menuItem{title: "返回主菜单", detail: "回到 Fatalder CLI 主菜单。", action: actionBackMain})
		return items
	case pageMore:
		return []menuItem{
			{title: "任务组配置", detail: "查看、选择和维护保存下来的构建任务组。", action: actionTaskGroups},
			{title: "服务器配置", detail: "维护连接目标、嵌入式运行模式和后续构建需要的服务器参数。", action: actionServers},
			{title: "返回主菜单", detail: "回到 Fatalder CLI 主菜单。", action: actionBackMain},
		}
	case pageServers:
		items := []menuItem{
			{title: "新建服务器配置", detail: "填写服务器连接参数并保存。", action: actionCreateServer},
		}
		for _, server := range m.servers {
			items = append(items, menuItem{
				title:  server.Name,
				detail: serverDetail(server),
				action: actionViewServer,
			})
		}
		items = append(items, menuItem{title: "返回更多功能", detail: "回到更多功能页面。", action: actionBackMore})
		return items
	case pageServerActions:
		return []menuItem{
			{title: "查看配置", detail: m.selectedServerDetail(), action: actionViewServer},
			{title: "编辑配置", detail: "修改当前服务器配置并保存。", action: actionEditServer},
			{title: "删除配置", detail: "删除当前服务器配置。", action: actionDeleteServer},
			{title: "返回服务器配置", detail: "回到服务器配置列表。", action: actionBackServers},
		}
	default:
		return []menuItem{
			{title: "构建建筑", detail: "从任务组配置中选择构建任务，并驱动机器人把建筑写入目标世界。", action: actionBuild},
			{title: "导出建筑", detail: "导出流程还没有接入，当前只展示入口和提示。", action: actionExport},
			{title: "更多功能", detail: "进入配置维护页面，管理任务组和服务器配置。", action: actionMore},
			{title: "退出程序", detail: "结束当前 CLI 会话，不会启动或修改任何任务。", action: actionExit},
		}
	}
}

func (m model) selectedServerDetail() string {
	if m.serverIndex < 0 || m.serverIndex >= len(m.servers) {
		return "服务器配置不存在。"
	}
	return serverDetail(m.servers[m.serverIndex])
}

func (m model) selectedServerName() string {
	if m.serverIndex < 0 || m.serverIndex >= len(m.servers) {
		return "未知"
	}
	return m.servers[m.serverIndex].Name
}

func serverDetail(server define.ServerConfig) string {
	return fmt.Sprintf("服务器码：%s\n创建时间：%s", emptyText(server.ServerCode), server.CreatedAt.Format("2006-01-02 15:04:05"))
}

func emptyText(value string) string {
	if value == "" {
		return "未设置"
	}
	return value
}
