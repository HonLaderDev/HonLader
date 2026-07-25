package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/HonLaderDev/HonLader/define"
)

type buildFieldKind int
type buildSection int

const (
	buildFieldText buildFieldKind = iota
	buildFieldBool
)

const (
	buildSectionGeneral buildSection = iota
	buildSectionWorld
	buildSectionTarget
	buildSectionBuild
	buildSectionAuto
	buildSectionAdvanced
)

type buildFormField struct {
	key  string
	name string
	kind buildFieldKind
}

const (
	buildKeyName                         = "name"
	buildKeyWorldPath                    = "world_path"
	buildKeyWorldStartPos                = "world_start_pos"
	buildKeyWorldEndPos                  = "world_end_pos"
	buildKeyWorldDimension               = "world_dimension"
	buildKeyStartPos                     = "start_pos"
	buildKeyDimension                    = "dimension"
	buildKeySpeed                        = "speed"
	buildKeyProgress                     = "progress"
	buildKeyChunkGroupSide               = "chunk_group_side"
	buildKeyGameProgressRefreshDelay     = "game_progress_refresh_delay"
	buildKeyConsoleWorldPos              = "console_world_pos"
	buildKeyFixModeTimeout               = "fix_mode_timeout"
	buildKeyEnableAutoCleanBlock         = "enable_auto_clean_block"
	buildKeyDisableAutoWaitChunkLoad     = "disable_auto_wait_chunk_load"
	buildKeyDisableAutoFillBuildMode     = "disable_auto_fill_build_mode"
	buildKeyDisableAutoCleanItem         = "disable_auto_clean_item"
	buildKeyDisableCommandBlocksDisabled = "disable_auto_command_blocks_disabled"
	buildKeyDisableAutoUpgradeCommand    = "disable_auto_upgrade_command_block"
	buildKeyEnableAutoPlaceDenyBlock     = "enable_auto_place_deny_block"
	buildKeyEnableAutoPlaceBorderBlock   = "enable_auto_place_border_block"
	buildKeyDisableAutoEnterFixMode      = "disable_auto_enter_fix_mode"
	buildKeyDisableGameProgress          = "disable_game_progress"
	buildKeyIgnoreCommandBlock           = "ignore_command_block"
	buildKeyIgnoreOtherNBTBlock          = "ignore_other_nbt_block"
	buildKeyEnterFixModeDirectly         = "enter_fix_mode_directly"
	buildKeyUseTickingArea               = "use_ticking_area"
	buildKeyPreWaitNextChunkLoad         = "pre_wait_next_chunk_load"
	buildKeyPreHandleNextChunkGroup      = "pre_handle_next_chunk_group"
)

type buildSectionInfo struct {
	title  string
	detail string
	value  buildSection
}

func buildSections() []buildSectionInfo {
	return []buildSectionInfo{
		{title: "基础配置", detail: "任务组名称等基础信息。", value: buildSectionGeneral},
		{title: "源世界配置", detail: "源世界路径、区域和维度。", value: buildSectionWorld},
		{title: "目标配置", detail: "目标世界起点和维度。", value: buildSectionTarget},
		{title: "构建配置", detail: "速度、进度、区块组和过滤选项。", value: buildSectionBuild},
		{title: "自动行为", detail: "自动清理、等待、deny/border 等开关。", value: buildSectionAuto},
		{title: "高级配置", detail: "常加载区域、预处理、修补模式等配置。", value: buildSectionAdvanced},
	}
}

func buildSectionTitle(section buildSection) string {
	for _, item := range buildSections() {
		if item.value == section {
			return item.title
		}
	}
	return "未知板块"
}

func buildBasicFields() []buildFormField {
	return []buildFormField{
		{key: buildKeyName, name: "任务组名称"},
		{key: buildKeyWorldPath, name: "世界路径"},
		{key: buildKeyWorldStartPos, name: "源起点 x,y,z"},
		{key: buildKeyWorldEndPos, name: "源终点 x,y,z"},
		{key: buildKeyWorldDimension, name: "源维度"},
		{key: buildKeyStartPos, name: "目标起点 x,y,z"},
		{key: buildKeyDimension, name: "目标维度"},
		{key: buildKeySpeed, name: "速度"},
	}
}

func buildFormFields(section buildSection) []buildFormField {
	switch section {
	case buildSectionGeneral:
		return []buildFormField{
			{key: buildKeyName, name: "任务组名称"},
		}
	case buildSectionWorld:
		return []buildFormField{
			{key: buildKeyWorldPath, name: "世界路径"},
			{key: buildKeyWorldStartPos, name: "源起点 x,y,z"},
			{key: buildKeyWorldEndPos, name: "源终点 x,y,z"},
			{key: buildKeyWorldDimension, name: "源维度"},
		}
	case buildSectionTarget:
		return []buildFormField{
			{key: buildKeyStartPos, name: "目标起点 x,y,z"},
			{key: buildKeyDimension, name: "目标维度"},
		}
	case buildSectionBuild:
		return []buildFormField{
			{key: buildKeySpeed, name: "速度"},
			{key: buildKeyProgress, name: "起始进度"},
			{key: buildKeyChunkGroupSide, name: "区块组边长"},
			{key: buildKeyDisableGameProgress, name: "关闭游戏内进度", kind: buildFieldBool},
			{key: buildKeyIgnoreCommandBlock, name: "忽略命令方块", kind: buildFieldBool},
			{key: buildKeyIgnoreOtherNBTBlock, name: "忽略其他 NBT 方块", kind: buildFieldBool},
		}
	case buildSectionAuto:
		return []buildFormField{
			{key: buildKeyEnableAutoCleanBlock, name: "自动清理方块", kind: buildFieldBool},
			{key: buildKeyDisableAutoWaitChunkLoad, name: "关闭等待区块加载", kind: buildFieldBool},
			{key: buildKeyDisableAutoFillBuildMode, name: "关闭 fill 构建模式", kind: buildFieldBool},
			{key: buildKeyDisableAutoCleanItem, name: "关闭清理掉落物", kind: buildFieldBool},
			{key: buildKeyDisableCommandBlocksDisabled, name: "关闭禁用命令方块", kind: buildFieldBool},
			{key: buildKeyDisableAutoUpgradeCommand, name: "关闭升级旧命令", kind: buildFieldBool},
			{key: buildKeyEnableAutoPlaceDenyBlock, name: "放置 deny 方块", kind: buildFieldBool},
			{key: buildKeyEnableAutoPlaceBorderBlock, name: "放置 border 方块", kind: buildFieldBool},
			{key: buildKeyDisableAutoEnterFixMode, name: "关闭进入修补模式", kind: buildFieldBool},
		}
	case buildSectionAdvanced:
		return []buildFormField{
			{key: buildKeyGameProgressRefreshDelay, name: "进度刷新延迟"},
			{key: buildKeyConsoleWorldPos, name: "控制台坐标 x,y,z"},
			{key: buildKeyFixModeTimeout, name: "修补模式超时"},
			{key: buildKeyEnterFixModeDirectly, name: "直接进入修补模式", kind: buildFieldBool},
			{key: buildKeyUseTickingArea, name: "使用常加载区域", kind: buildFieldBool},
			{key: buildKeyPreWaitNextChunkLoad, name: "预等待下个区块", kind: buildFieldBool},
			{key: buildKeyPreHandleNextChunkGroup, name: "预处理下个区块组", kind: buildFieldBool},
		}
	default:
		return []buildFormField{
			{key: buildKeyName, name: "任务组名称"},
		}
	}
}

func newBuildForm() buildForm {
	form := buildForm{values: make(map[string]string)}
	form.values[buildKeyName] = ""
	form.values[buildKeyWorldStartPos] = "0,0,0"
	form.values[buildKeyWorldEndPos] = "0,0,0"
	form.values[buildKeyWorldDimension] = "0"
	form.values[buildKeyStartPos] = "0,0,0"
	form.values[buildKeyDimension] = "0"
	form.values[buildKeySpeed] = "3000"
	form.values[buildKeyProgress] = "0"
	form.values[buildKeyChunkGroupSide] = "2"
	form.values[buildKeyGameProgressRefreshDelay] = "0.5"
	form.values[buildKeyConsoleWorldPos] = "-50,0,-50"
	form.values[buildKeyFixModeTimeout] = "10"
	return form
}

func (m *model) finishBuildBasicForm() {
	m.fillDefaultBuildTaskName()
	m.buildMoreConfirm = true
	m.buildMoreCursor = 0
}

func (m *model) fillDefaultBuildTaskName() {
	if strings.TrimSpace(m.buildForm.values[buildKeyName]) != "" || m.buildServer == nil {
		return
	}
	m.buildForm.values[buildKeyName] = fmt.Sprintf("%s-%s", m.buildServer.Metadata.Name, time.Now().Format("2006-01-02 15:04:05"))
}

func (m *model) toggleBuildBool(key string) {
	if m.buildForm.values[key] == "true" {
		m.buildForm.values[key] = "false"
		return
	}
	m.buildForm.values[key] = "true"
}

func (m *model) saveBuildForm() {
	m.fillDefaultBuildTaskName()
	name := strings.TrimSpace(m.buildForm.values[buildKeyName])
	if name == "" {
		m.dialog = "任务组名称不能为空。"
		return
	}
	if m.buildServer == nil {
		m.dialog = "请先选择服务器配置。"
		return
	}

	config, err := m.buildTaskConfig()
	if err != nil {
		m.dialog = err.Error()
		return
	}
	taskGroup := define.TaskGroupConfig{Tasks: []define.TaskConfig{config}}
	if err := m.dataManager.SaveTaskConfig(name, taskGroup); err != nil {
		m.dialog = "保存任务配置失败：" + err.Error()
		return
	}
	m.page = pageMain
	m.cursor = 0
	m.dialog = fmt.Sprintf("已保存构建任务组：%s\n服务器配置：%s", name, m.buildServer.Metadata.Name)
	m.buildForm = buildForm{}
	m.buildServer = nil
}

func (m model) buildTaskConfig() (define.TaskConfig, error) {
	config := define.TaskConfig{
		buildKeyWorldPath:                    strings.TrimSpace(m.buildForm.values[buildKeyWorldPath]),
		buildKeyProgress:                     strings.TrimSpace(m.buildForm.values[buildKeyProgress]),
		buildKeyEnableAutoCleanBlock:         parseBool(m.buildForm.values[buildKeyEnableAutoCleanBlock]),
		buildKeyDisableAutoWaitChunkLoad:     parseBool(m.buildForm.values[buildKeyDisableAutoWaitChunkLoad]),
		buildKeyDisableAutoFillBuildMode:     parseBool(m.buildForm.values[buildKeyDisableAutoFillBuildMode]),
		buildKeyDisableAutoCleanItem:         parseBool(m.buildForm.values[buildKeyDisableAutoCleanItem]),
		buildKeyDisableCommandBlocksDisabled: parseBool(m.buildForm.values[buildKeyDisableCommandBlocksDisabled]),
		buildKeyDisableAutoUpgradeCommand:    parseBool(m.buildForm.values[buildKeyDisableAutoUpgradeCommand]),
		buildKeyEnableAutoPlaceDenyBlock:     parseBool(m.buildForm.values[buildKeyEnableAutoPlaceDenyBlock]),
		buildKeyEnableAutoPlaceBorderBlock:   parseBool(m.buildForm.values[buildKeyEnableAutoPlaceBorderBlock]),
		buildKeyDisableAutoEnterFixMode:      parseBool(m.buildForm.values[buildKeyDisableAutoEnterFixMode]),
		buildKeyDisableGameProgress:          parseBool(m.buildForm.values[buildKeyDisableGameProgress]),
		buildKeyIgnoreCommandBlock:           parseBool(m.buildForm.values[buildKeyIgnoreCommandBlock]),
		buildKeyIgnoreOtherNBTBlock:          parseBool(m.buildForm.values[buildKeyIgnoreOtherNBTBlock]),
		buildKeyEnterFixModeDirectly:         parseBool(m.buildForm.values[buildKeyEnterFixModeDirectly]),
		buildKeyUseTickingArea:               parseBool(m.buildForm.values[buildKeyUseTickingArea]),
		buildKeyPreWaitNextChunkLoad:         parseBool(m.buildForm.values[buildKeyPreWaitNextChunkLoad]),
		buildKeyPreHandleNextChunkGroup:      parseBool(m.buildForm.values[buildKeyPreHandleNextChunkGroup]),
	}

	if err := setBlockPos(config, buildKeyWorldStartPos, m.buildForm.values[buildKeyWorldStartPos]); err != nil {
		return nil, fmt.Errorf("源起点格式错误：%w", err)
	}
	if err := setBlockPos(config, buildKeyWorldEndPos, m.buildForm.values[buildKeyWorldEndPos]); err != nil {
		return nil, fmt.Errorf("源终点格式错误：%w", err)
	}
	if err := setDimension(config, buildKeyWorldDimension, m.buildForm.values[buildKeyWorldDimension]); err != nil {
		return nil, fmt.Errorf("源维度格式错误：%w", err)
	}
	if err := setBlockPos(config, buildKeyStartPos, m.buildForm.values[buildKeyStartPos]); err != nil {
		return nil, fmt.Errorf("目标起点格式错误：%w", err)
	}
	if err := setDimension(config, buildKeyDimension, m.buildForm.values[buildKeyDimension]); err != nil {
		return nil, fmt.Errorf("目标维度格式错误：%w", err)
	}
	for _, item := range []struct {
		key   string
		label string
	}{
		{buildKeySpeed, "速度"},
		{buildKeyChunkGroupSide, "区块组边长"},
	} {
		if err := setOptionalInt(config, item.key, m.buildForm.values[item.key]); err != nil {
			return nil, fmt.Errorf("%s格式错误：%w", item.label, err)
		}
	}
	for _, item := range []struct {
		key   string
		label string
	}{
		{buildKeyGameProgressRefreshDelay, "进度刷新延迟"},
		{buildKeyFixModeTimeout, "修补模式超时"},
	} {
		if err := setOptionalFloat(config, item.key, m.buildForm.values[item.key]); err != nil {
			return nil, fmt.Errorf("%s格式错误：%w", item.label, err)
		}
	}
	if err := setOptionalBlockPos(config, buildKeyConsoleWorldPos, m.buildForm.values[buildKeyConsoleWorldPos]); err != nil {
		return nil, fmt.Errorf("控制台坐标格式错误：%w", err)
	}
	return config, nil
}

func setBlockPos(config define.TaskConfig, key string, value string) error {
	pos, err := parseBlockPos(value)
	if err != nil {
		return err
	}
	config[key] = pos
	return nil
}

func setOptionalBlockPos(config define.TaskConfig, key string, value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return setBlockPos(config, key, value)
}

func setDimension(config define.TaskConfig, key string, value string) error {
	dimension, err := parseDimension(value)
	if err != nil {
		return err
	}
	config[key] = dimension
	return nil
}

func setOptionalInt(config define.TaskConfig, key string, value string) error {
	num, err := parseOptionalInt(value)
	if err != nil {
		return err
	}
	if num != nil {
		config[key] = *num
	}
	return nil
}

func setOptionalFloat(config define.TaskConfig, key string, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	config[key] = num
	return nil
}

func parseBlockPos(value string) (define.BlockPos, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 3 {
		return define.BlockPos{}, fmt.Errorf("需要 x,y,z")
	}
	var nums [3]int
	for i, part := range parts {
		num, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return define.BlockPos{}, err
		}
		nums[i] = num
	}
	return define.BlockPos{nums[0], nums[1], nums[2]}, nil
}

func parseDimension(value string) (define.Dimension, error) {
	num, err := strconv.Atoi(strings.TrimSpace(value))
	return define.Dimension(num), err
}

func parseOptionalInt(value string) (*int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &num, nil
}

func parseBool(value string) bool {
	return strings.TrimSpace(value) == "true"
}
