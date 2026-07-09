# TUI 构建配置设计草案

## 参考结论

本设计参考了旧项目 `~/Fatalder/control/task` 中 `BuildTask` 的字段与实际执行路径。

关键结论：

- `fill` 不是服务器兼容能力。它是普通方块阶段的构建策略：开启时把连续同类方块合并为更少的 `fill` 命令，关闭时退回逐方块 `setblock`。
- `speed` 是命令发送速率限制，不是构建质量开关。速度过高主要影响目标服承压能力。
- `chunk_group_side` 是一次处理的区块组边长，会同时影响读取粒度、清理范围、加载检查范围和断点推进粒度。
- `wait_chunk_load` 与 `use_ticking_area` 都是区块可访问性保障；前者探测区块是否已加载，后者用常加载区域辅助加载。
- `pre_handle_next_chunk_group` 是提前读取和生成下一组数据，主要消耗本机 CPU/IO；`pre_wait_next_chunk_load` 是提前让服务器加载下一组区域，主要消耗服务器加载能力和常加载区域槽位。
- `console_world_pos` 只服务 NBT 方块写入的控制台区域，不应放在普通目标坐标配置里让用户误以为它会影响普通方块。
- `deny` 与 `border` 会改变生成后的建筑边界/底层结构，不只是“保护开关”，应该在摘要里显式提示。
- `commandBlocksEnabled` 是运行保护类游戏规则，不是方块处理策略。它的意义是避免构建期间已有命令方块干扰机器人，例如传送、击杀、踢出或修改游戏状态。

## 设计目标

TUI 不应该把所有字段平铺给用户。推荐采用“两段式”流程：

- 第一段是快速向导，只收集能开始构建的最小信息。
- 第二段是高级配置，按用户意图分组修改已有字段。

配置分组按“用户想解决什么问题”划分，而不是按 Go 结构体或字段前缀划分。

## 推荐主流程

```text
请选择构建预设：
  [1] 标准：稳定默认，适合大多数构建
  [2] 快速：提高吞吐，启用预处理与预加载
  [3] 保守：降低吞吐，减少并发和服务器压力
  [4] 修补：从指定进度继续，进入修补相关流程
> 1

请输入建筑文件路径：./world.mcworld

请输入建筑区域(例如 0,0,0~170,320,220)：0,0,0~170,320,220

请输入构建起点(例如 1200,0,1200)：1200,0,1200

请选择构建维度：
  [1] 主世界
  [2] 下界
  [3] 末地
  [4] 自定义
> 1

配置摘要：
  预设：标准
  建筑文件：./world.mcworld
  建筑区域：0,0,0 ~ 170,320,220
  建筑尺寸：171 x 321 x 221
  构建起点：1200,0,1200
  构建维度：主世界
  命令速度：3000/s
  区块组：2 x 2
  命令合并：开启
  等待区块加载：开启
  常加载区域：关闭
  起始进度：0

确认开始？[Y/n]
是否继续配置[y/N]：
```

说明：

- 摘要先于“是否继续配置”，因为用户可能确认无误后直接开始。
- “是否继续配置”只在用户不确认开始或主动选择继续时进入高级菜单。
- 区域输入提示只展示一种格式，但解析可以兼容 `0,0,0 170,320,220`。

## 高级配置菜单

推荐菜单：

```text
请选择配置项：
  [1] 建筑来源：建筑文件、建筑区域、源维度
  [2] 构建位置：构建起点、目标维度
  [3] 构建策略：fill 命令合并、命令速度、区块组边长
  [4] 加载策略：等待区块加载、常加载区域、预加载下一组
  [5] 预处理策略：预读取下一组、预生成下一组命令
  [6] 运行保护：禁用命令方块运行、构建前清理方块、构建后清理掉落物
  [7] 边界保护：deny、border
  [8] 特殊方块：命令方块、其他 NBT 方块、旧命令升级、NBT 控制台坐标
  [9] 断点与修补：起始进度、直接进入修补模式、自动进入修补模式、修补超时
  [10] 进度显示：游戏内进度、刷新间隔
  [11] 查看摘要并开始
```

### 为什么这样分

- `构建策略` 是“如何把建筑变成命令”：`fill`、`setblock`、速度和区块组都影响命令数量或发送节奏。
- `加载策略` 是“目标区域是否已加载”：等待区块、常加载区域、预加载都依赖服务器加载状态。
- `预处理策略` 是“本机是否提前干活”：读取下一组、生成下一组命令不等于服务器加载。
- `运行保护` 是“如何避免构建现场影响机器人”：禁用命令方块运行、清理方块和清理掉落物都服务于稳定执行。
- `边界保护` 是“是否额外生成保护结构”：deny 和 border 会改变目标建筑占用范围，需要单独展示。
- `特殊方块` 是“普通方块以外的数据如何写入”：命令方块、其他 NBT 方块和旧命令升级走单独阶段。

## 字段语义与归类

| UI 名称 | 当前字段 | 语义 | 推荐分组 |
| --- | --- | --- | --- |
| 建筑文件 | `world_path` | 源世界目录或压缩包路径。文件按 zip/mcworld 打开，目录按基岩版存档打开。 | 建筑来源 |
| 建筑区域 | `world_start_pos`, `world_end_pos` | 从源世界裁剪的方块范围。 | 建筑来源 |
| 源维度 | `world_dimension` | 从源世界哪个维度读取建筑数据。 | 建筑来源 |
| 构建起点 | `start_pos` | 建筑局部坐标 `(0,0,0)` 对应的目标世界坐标。 | 构建位置 |
| 目标维度 | `dimension` | 在目标服务器哪个维度构建。 | 构建位置 |
| 命令速度 | `speed` | 每秒命令发送速率限制。 | 构建策略 |
| 区块组边长 | `chunk_group_side` | 每次处理 N x N 个区块，影响吞吐、内存、加载范围和断点粒度。 | 构建策略 |
| fill 命令合并 | `!disable_auto_fill_build_mode` | 开启后将连续方块合并为 `fill`，显著减少普通方块命令数量。 | 构建策略 |
| 等待区块加载 | `!disable_auto_wait_chunk_load` | 构建前探测目标区块组是否可访问。 | 加载策略 |
| 使用常加载区域 | `use_ticking_area` | 用 `tickingarea` 辅助目标区块组加载。 | 加载策略 |
| 预加载下一组 | `pre_wait_next_chunk_load` | 当前组构建时后台等待下一组区块加载。 | 加载策略 |
| 预处理下一组 | `pre_handle_next_chunk_group` | 当前组构建时后台读取下一组数据并准备构建输入。 | 预处理策略 |
| 禁用命令方块运行 | `!disable_auto_command_blocks_disabled` | 构建前设置 `commandBlocksEnabled false`，避免已有命令方块干扰机器人；结束后恢复。 | 运行保护 |
| 构建前清理方块 | `enable_auto_clean_block` | 构建当前区块组前用 air 清理目标范围，避免旧方块阻挡构建结果。 | 运行保护 |
| 构建后清理掉落物 | `!disable_auto_clean_item` | 构建当前区块组后清理目标范围内掉落物，避免实体堆积影响执行。 | 运行保护 |
| 放置 deny | `enable_auto_place_deny_block` | 在建筑底层生成 deny 保护层，会影响建筑占用范围。 | 边界保护 |
| 放置 border | `enable_auto_place_border_block` | 在建筑边界生成 border，会影响建筑占用范围。 | 边界保护 |
| 构建命令方块 | `!ignore_command_block` | 是否处理命令方块 NBT。 | 特殊方块 |
| 构建其他 NBT 方块 | `!ignore_other_nbt_block` | 是否处理箱子、告示牌等非命令方块 NBT。 | 特殊方块 |
| 自动升级旧命令 | `!disable_auto_upgrade_command_block` | 写入命令方块时尝试升级旧版命令文本。 | 特殊方块 |
| 起始进度 | `progress` | 从指定区块序号或百分比开始。断点 `current_chunk` 优先级更高。 | 断点与修补 |
| 直接进入修补模式 | `enter_fix_mode_directly` | 连接并进入修补流程，不执行完整构建。 | 断点与修补 |
| 自动进入修补模式 | `!disable_auto_enter_fix_mode` | 构建完成后自动进入修补流程。 | 断点与修补 |
| 修补超时 | `fix_mode_timeout` | 修补模式等待玩家操作或反馈的超时时间。 | 断点与修补 |
| 游戏内进度 | `!disable_game_progress` | 是否在游戏内 actionbar 展示进度。 | 进度显示 |
| 进度刷新间隔 | `game_progress_refresh_delay` | 游戏内进度刷新频率。 | 进度显示 |
| 控制台坐标 | `console_world_pos` | NBT 方块写入使用的临时控制台区域坐标。 | 特殊方块 |

## 推荐预设

预设只写入默认值，不锁死字段。用户进入高级配置后可以覆盖。

### 标准

目标是稳定且不太慢。

```text
speed = 3000
chunk_group_side = 2
disable_auto_fill_build_mode = false
disable_auto_wait_chunk_load = false
use_ticking_area = false
pre_handle_next_chunk_group = false
pre_wait_next_chunk_load = false
enable_auto_clean_block = false
disable_auto_clean_item = false
```

### 快速

目标是更高吞吐，适合本机和目标服务器都比较稳的场景。

```text
speed = 6000
chunk_group_side = 3
disable_auto_fill_build_mode = false
disable_auto_wait_chunk_load = false
use_ticking_area = true
pre_handle_next_chunk_group = true
pre_wait_next_chunk_load = true
```

风险提示：

- 会增加本机预读取/命令生成压力。
- 会更频繁地触发目标服务器区块加载。
- 常加载区域槽位不足时，应自动降级为普通等待区块加载。

### 保守

目标是降低服务器压力，适合弱服务器或不稳定网络。

```text
speed = 1200
chunk_group_side = 1
disable_auto_fill_build_mode = false
disable_auto_wait_chunk_load = false
use_ticking_area = false
pre_handle_next_chunk_group = false
pre_wait_next_chunk_load = false
```

说明：

- 保守模式仍建议开启 `fill` 命令合并，因为它减少命令数量，通常比逐方块 `setblock` 更稳。
- 保守模式真正要关闭的是并发预处理、预加载和较大的区块组。

### 修补

目标是恢复或检查已有构建。

```text
progress = 用户输入
enter_fix_mode_directly = 视用户选择
disable_auto_enter_fix_mode = false
fix_mode_timeout = 10
```

修补预设应追加询问：

```text
请选择修补方式：
  [1] 从指定进度继续构建
  [2] 直接进入修补模式
```

## 字段展示建议

界面上尽量使用正向文案，内部再映射到现有反向字段。

| UI 文案 | 内部映射 |
| --- | --- |
| fill 命令合并 | `!disable_auto_fill_build_mode` |
| 等待区块加载 | `!disable_auto_wait_chunk_load` |
| 禁用命令方块运行 | `!disable_auto_command_blocks_disabled` |
| 清理掉落物 | `!disable_auto_clean_item` |
| 自动升级旧命令 | `!disable_auto_upgrade_command_block` |
| 自动进入修补模式 | `!disable_auto_enter_fix_mode` |
| 游戏内进度 | `!disable_game_progress` |
| 构建命令方块 | `!ignore_command_block` |
| 构建其他 NBT 方块 | `!ignore_other_nbt_block` |

## 摘要建议

摘要应该展示用户最容易配错、且会显著影响结果的字段。

建议展示：

- 预设
- 建筑文件
- 建筑区域
- 建筑尺寸
- 源维度
- 构建起点
- 目标维度
- 命令速度
- 区块组边长
- fill 命令合并
- 等待区块加载
- 常加载区域
- 预处理下一组
- 预加载下一组
- 禁用命令方块运行
- 清理方块
- 清理掉落物
- deny / border
- 命令方块 / 其他 NBT
- 起始进度
- 修补模式

摘要后必须确认：

```text
确认开始？[Y/n]
```

## 暂不建议暴露的旧字段

旧 `Fatalder` 中还有一些字段当前 HonLader 构建任务没有完整对应，TUI 暂不应该先暴露：

- `convert_to_mcworld`：当前构建入口已经以基岩版世界目录/压缩包为输入，不应在 TUI 中先引入结构文件转换。
- `verify_after_chunk`、`verify_chunk_level`：旧版有构建后校验和修补闭环，当前 HonLader 构建主流程未完整迁移这套字段。
- `billing_mode`、`billing_claim_key`：计费相关，不属于本地 TUI 构建配置。
- `show_nbt_progress`：当前进度展示只保留游戏内进度刷新相关字段。

如果后续实现这些能力，应新增独立分组，不要塞进现有“性能策略”或“加载策略”里。
