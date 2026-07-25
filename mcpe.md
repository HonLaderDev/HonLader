# MCPE 静态追踪笔记

目标：找到 `SelfSignedId` / `IdentityData.Identity` 的真实来源。

已确认：

1. `0xa650b58` / `0xa7732e0` 是 `SelfSignedId` 的 JSON 输出路径。
2. `0x78ab804 -> 0x678272c` 会把 `x20+0x68` 传给后续结构拼装。
3. `0x78b55cc` 是 auth-state 结构拷贝构造，`x0+0x68 <- x1+0x68`。
4. `0x78ab3b4` 的两个直接调用点：
   - `0x635ad70`
   - `0x783a480`
5. `0x783a480` 先调用 `0xa77bd10` 生成临时 auth 结构，再调用 `0x78ab3b4`。
6. `0xa77bd10` 是一个 auth/identity 结构构造函数，字段大致为：
   - `x2 -> +0x8`
   - `x3 -> +0x20`
   - `x5 -> +0x38`
   - `x6 -> +0x50/+0x60`
   - `sp[8] -> +0x68/+0x78`
   - `w7 -> +0x80`
7. `0xa77c24c(object, string, flag)` 是 `SelfSignedId` 的 setter。
8. `0xa84c988` 是 UUIDv3 生成器，格式为：
   - `MD5("pocket-auth-1-xuid:" + input)`
9. `0xa84c988` 的直接调用点：
   - `0x6e75eec`
   - `0x7d6cfe0`
   - `0x9bbd44c`

当前判断：

- `SelfSignedId` 的结构字段在 `auth_obj + 0x10 + 0x60`
- 真实来源大概率来自反序列化输入中的 `"SelfSignedId"` 字段
- 另一条独立链在生成 `IdentityData.Identity`，和 `pocket-auth-1-xuid:` / UUIDv3 有关

待追：

1. `0x76ca568` 中 `x6` 的具体来源。
2. `0xa77c24c` 的表驱动注册路径。
3. `0x9bbd44c` 的返回 UUID 是否最终进入 `IdentityData.Identity`。

新增观察：

1. `0x76ca400` 只调用 `0x76ca4a0`。
2. `0x76ca4a0` 是一个轻量包装：
   - 若 `x0+0x18` 置位，则释放 `[x19+0x28]`
   - 最后返回 `x19+0x10`
3. 所以 `0x76ca400` 的后续链里，`x19+0x10` 很可能就是 auth 子对象入口，后面再接 `0xa77bd10`。
4. `0x76ca4d0` 再往下会把一组临时 string 组装进 `0xa77bd10`，其中 `x6` 来自 `sp+0x40`。

修正：

1. `0x76ca4d0 -> 0xa77bd10` 这条路径里 `x6 = sp+0x40`，而 `sp+0x40` 在当前函数内是临时/空字段，不像最终 JSON 输出的真实 `SelfSignedId` 来源。
2. JSON 输出路径更关键：
   - `0x76c9f7c` 取 `x0 = [x19+0x58]`
   - 虚表调用 `[vtable+0x4c0]`
   - 返回对象后 `add x0, x0, #0x10`
   - `bl 0xa77f950`
   - `x24 = 返回值`
   - `0x76ca0c8 -> 0xa650b58` 时 `x2 = x24`
3. 所以真实 `SelfSignedId` 来源现在应追 `[x19+0x58]` 指向对象的虚表槽 `0x4c0` 对应实现。

进一步确认：

1. `0x76c9f8c: add x0, x0, #0x10`
2. `0x76c9f90: bl 0xa77f950`
3. `0xa77f950: add x0, x0, #0x60`
4. 因此 `x24 = auth_obj + 0x10 + 0x60`
5. `0x76ca0c8 -> 0xa650b58` 时，`x2 = x24`，所以 JSON 输出里的 `SelfSignedId` 就是这个字段
6. `0x76c9eb8` / `0x76c9ec4` 说明 `0x4c0` 虚表槽返回的是 auth 包/认证对象，再经 `+0x10` 进入其内部 auth data

2026-07-20 15:14 继续追踪：

1. `0x76c9e80..0x76ca0c8` 里 `[x19+0x58]` 对象被连续调用多个虚表槽：
   - `vtable+0x5d0`：结果作为某个字段传给 `0xa860b58`
   - `vtable+0x4c0`：返回 auth 包/认证对象
   - `vtable+0x4d8`：生成 `sp+0x98` 这一路
   - `vtable+0x8c8`：返回 bool/flag 到 `w27`
   - `vtable+0x5f0`：返回另一个对象，再取其内部虚表 `+0x158/+0x290/+0x288`
   - `vtable+0xc10`：返回 bool/flag 到 `sp+0x30`
2. `0x76c9f7c -> vtable+0x4c0 -> add +0x10 -> 0xa77f950 -> +0x60` 是当前最硬证据链。
3. `0x783a338..0x783a480` 的 `0xa77bd10` 调用是构造临时 auth-state 后传给 `0x78ab3b4`：
   - `x6 = sp+0x98`
   - 该临时结构后续被塞进 `x19 = [x19+0xe0]` 对应对象
   - 这条链看起来是更新/插入 auth-state，不是 `0x76ca0c8` JSON 输出时直接读取的最终来源。
4. `0x635ad2c..0x635ad70` 是另一个 `0x78ab3b4` 包装调用：
   - 先调用传入对象的 `vtable+0xe30`
   - 从 `sp+0x10` 取目标对象作为 `x0`
   - 再把外部传入的 auth-state `x1` 和 flag `w2` 传给 `0x78ab3b4`
5. 当前下一步：追 `[x19+0x58]` 这个成员是谁写入的，或识别它的虚表/类名；不要再把 `0x783a480` 的临时 `x6` 当最终 `SelfSignedId` 来源。

2026-07-20 15:14 补充：

1. `strings` 里还能看到：
   - `get_auth_netease_sid`
   - `netease_sid`
   - `pocket-auth-1-xuid:`
   - `SelfSignedId`
2. `0x9bbd44c` 是另一条独立 UUID 链：
   - 直接 `bl 0xa84c988`
   - `0xa84c988` 负责 `MD5("pocket-auth-1-xuid:" + input)` 的 UUIDv3 生成
   - 结果被 `stp x0, x1, [x19, #0x1a0]`
3. 这条链更像 `IdentityData.Identity` / `netease_sid` 相关生成逻辑，不是 `SelfSignedId` 主输出链。
4. `0xa773400` 附近的字符串拼装已经命中 `SelfSignedId` 字面名，说明这里是反射/序列化注册点之一；但它本身还不能直接说明来源，只能证明字段被纳入同一类数据模型。

2026-07-20 继续追 `[x19+0x58]`：

1. 新增临时工具：
   - `tmp/revtools/scan_vslot.c`：扫描 `ldr vtable; ldr slot; blr` 形态的虚表调用。
   - `tmp/revtools/scan_vtable_slot.c`：尝试从数据段反推虚表槽实现；当前受 PLT/trampoline 和 relocation 影响，噪声很大，只作为辅助。
2. `vtable+0x4c0` 全库直接调用只有 17 个，集中在当前登录/连接处理代码：
   - `0x76c8808`
   - `0x76c883c`
   - `0x76c9eb4`
   - `0x76c9f80`
   - `0x76c9f9c`
   - `0x76cb650`
   - `0x76cb670`
   - `0x76cb68c`
   - 以及少量其它上下文调用
3. `0x76c7a8c` 是一个 move/copy 构造，整体搬迁包含 `+0x58` 的对象：
   - `x20 = ([x0+0x90] == 0 ? x1 : x0)`，即按 flag 选择源对象
   - `ldp x8, x9, [x20, #0x58]`
   - `stp x8, x9, [x19, #0x58]`
   - 随后清空源对象 `x20+0x58`
4. `0x76c7a8c` 只有一个直接调用点：
   - `0x76c6304`
5. `0x76c6304` 的调用形态：
   - `x0 = sp+0x4c0`
   - `x1 = sp+0x920`
   - `bl 0x76c7a8c`
   - 所以这里把临时对象 `sp+0x920` 搬到 `sp+0x4c0`
6. `0x76c7c78` 是另一个构造列表/元素的函数，直接调用点：
   - `0x76c6240`
   - 它内部 `0x76c7ef4: str x9, [x8,#0x58]` 的 `x9` 来自局部 `[x29-0x28]`，该局部此前被清零；这条更像列表元素临时字段，不像最终 `SelfSignedId` 的来源。
7. `0x76c5f88 -> 0x7486168` 会按 `sp+0x9b0` 查找已有对象：
   - 若命中，取 `[x0+0x40]` 放到 `sp+0x1c8`
   - 若未命中，走 `0x76c6278` 创建空/临时对象，再通过 `0x76c6304` 搬到 `sp+0x4c0`
8. 当前更可靠的来源表述：
   - `SelfSignedId` 最终来自连接/登录请求对象提供的 auth data：`[sp+0x280] -> vtable+0x4c0 -> +0x10 -> +0x60`
   - 该 auth data 进入 JSON writer 前未在 `0x76ca0c8` 现场生成。
   - 还需要继续追 `sp+0x280` 这个连接/登录请求对象的构造与其 `vtable+0x4c0` 返回对象的真实填充点。

2026-07-20 继续追 `0x95464a4`：

1. `0x95464a4` 不是直接生成 `SelfSignedId` 的函数，而是服务/组件定位器：
   - `0x95464a8: ldr x0, [x0,#0x38]` 取调用方上下文里的 key/handle。
   - 查全局哈希表：`0x1336d830`、`0x1336d838`、`0x1336d840`。
   - 命中后 `0x9546590: ldr w8, [x11,#0x14]`，`0x9546594: eor x0, x8, x0` 返回解混淆后的对象指针。
2. 因此 `0x76c5198: bl 0x95464a4` 之后的 `x0/x23` 是某个服务对象，不是 auth data 本体。
3. 下一步继续追：`0x76c5198` 调用前 `x0` 的来源，以及返回服务对象 `vtable+0x1e8` 在 `0x76c51a8` 输出到 `sp+0x270/sp+0x280` 的实际实现。

2026-07-20 继续追 `sp+0x280` 到最终请求对象：

1. `0x76c51xx` 所在函数由字符串确认是 `ChangeSkinWithSkinInfo`：
   - `0x29b6aec`: `ChangeSkinWithSkinInfo item_id=%s  uuid=%s`
   - `0x2becdf3`: `ChangeSkinWithSkinInfo item_id=%s`
   - `0x2a407ec`: `ChangeSkinWithSkinInfo minecraftGame not found`
2. `0x9545d98 -> 0x95464a4` 返回的对象更像 `MinecraftGame` 服务；`0x76c51a8` 的 `vtable+0x1e8` 把一个 shared/optional 风格对象写到 `sp+0x270`，实际对象指针在 `sp+0x280`。
3. `sp+0x280` 后续多次作为游戏/客户端对象使用：
   - `vtable+0x370` -> 输出到 `sp+0x260`
   - `vtable+0x378` -> 返回对象给后续协议/请求构造
   - `vtable+0x3a8` -> 输出字符串/数据到 `sp+0xdf0` 或 `sp+0x4c0`
   - `vtable+0xa10`、`vtable+0xb90` 等也在同一流程使用
4. 当前最终 JSON/auth 候选对象不是 `sp+0x280` 本身，而是 `sp+0x4c0`：
   - `0x76c5f58`: `[sp+0x280]->vtable+0x3a8` 输出到 `sp+0x4c0`
   - `0x76c5f88`: `0x7486168(x24, sp+0x9b0)` 查缓存/表
   - 命中则 `0x76c5f90: ldp x25,x8,[x0,#0x40]` 到 `sp+0x1c8`
   - 未命中则 `0x76c6278..0x76c6304` 新建 `sp+0x920`，再 `0x76c7a8c(sp+0x4c0, sp+0x920)` move 到 `sp+0x4c0`
5. 这修正当前表述：`SelfSignedId` 来源不是 `MinecraftGame` 直接生成，而是 `MinecraftGame`/客户端对象提供的数据参与构造 `sp+0x4c0` 请求对象；后续仍需证明 `sp+0x4c0 + 0x58` 何时被填成最终 `vtable+0x4c0` 可读的 auth provider。

2026-07-20 拆 `0xdc6a944 -> 0xdc5ea3c` 后的修正：

1. `0xdc6a944` 只是分配 `0x330` 字节对象并转发参数到 `0xdc5ea3c`，最后把新对象挂到输出 smart pointer。
2. `0xdc5ea3c` 构造的是换肤请求对象，字段布局里 `x19+0x48/+0x58` 是 `std::string` 内部布局（`+0x58` 是长字符串堆指针/容量相关），不是前面 JSON 链中 `[state+0x58]` 的 provider 指针。
3. 因此不能把 `dc5*` 区域里的 `+0x58` 当作 `SelfSignedId` provider 字段。
4. `0x76c66f4: x7 = sp+0x4c0` 传入 `0xdc6a944` 的是请求构造附加结构；`0x76c6638` 清零 `sp+0x518` 证明这一路不直接持有最终 `SelfSignedId`。
5. 现在应回到 `0x76c77f8`、`0x76c7a8c`、`0x76c4c54` 这些 `0x76c` 内部状态/缓存操作函数，追谁把请求对象放到后续 `0x76c9e80` 会读取的状态对象成员 `+0x58`。

2026-07-20 排除 `0x76c7ef4`：

1. `0x76c7c78` 是创建空状态项/列表元素的路径。
2. `0x76c7ef4: str x9, [x8,#0x58]` 看起来像写 provider，但 `x9` 来源是 `[x29-0x28]`。
3. `[x29-0x28]` 属于以 `x26 = x29-0x80` 为基址的临时结构 `+0x58`，在 `0x76c7d84: stur q0, [x26,#0x58]` 被清零；中间未看到真实 provider 填入。
4. 所以 `0x76c7ef4` 是空状态项初始化，不是 `SelfSignedId` 来源。
5. 当前只剩有值路径：
   - `0x76c5f88/0x76c635c -> 0x7486168` 查缓存/表
   - 命中后 `ldp [entry+0x40]` 取已有状态对象到 `sp+0x1c8` 或 `sp+0x4c0`
   - 该已有状态对象的 `+0x58` 才是后续 JSON 输出读取的 auth provider。
6. 下一步应追 `0x7486168` 返回 entry 的插入/构造点，找谁把 `[entry+0x40]` 写入缓存。

2026-07-20 继续追 `0x7486168` 缓存 entry 的 value 来源：

1. `0x12706410` 这张 vtable/函数表的直接 `adrp/add` 引用只有 5 个：
   - `0x742e444`
   - `0x74866c4`
   - `0x7486744`
   - `0x74867c4`
   - `0x74868d0`
2. `0x74867c0..0x7486854` 和 `0x74868b0..0x7486990` 是 `0x70` 大小 value 的 copy/copy-to-dest 构造：
   - 会复制/克隆 `[src+0x40]` 的多态对象。
   - 会把 `[src+0x50..0x68]` 搬到 `[dst+0x50..0x68]`。
   - 因此它们只是传播 auth-provider 状态，不是原始生成点。
3. `0x742e360` 更像 value 原始构造/插入入口：
   - `0x742e430: mov w0,#0x70` 分配 value。
   - `0x742e444` 设置同一张 vtable `0x12706410`。
   - `0x742e4a0..0x742e4b8` 从调用入参 `x1` 指向的数据复制到新 value 的 `+0x50..+0x68`。
   - `0x742e4c0..0x742e4dc` 处理长字符串/堆数据形式，同样写入新 value 的 `+0x50..+0x68`。
4. `scan_bl` 未找到 `0x742e360` 的直接 `bl` 调用点，说明它很可能通过模板/函数指针/表驱动调用。
5. 下一步：反查 `0x742e360` 的函数指针表引用，或追 `0x742e360` 入参 `x1` 的来源。这个 `x1` 当前是最接近 `SelfSignedId` provider 状态真实来源的入口参数。

2026-07-20 进一步确认 `0x742e360` 的位置：

1. `llvm-readelf -r` 看到 `0x12704638 -> R_AARCH64_RELATIVE 0x742e360`，说明 `0x742e360` 是一张数据表/vtable 里的第 7 个槽位，不是普通直接调用目标。
2. 同一张表里还有 `0x12704600 -> 0x742a2c0`、`0x12704608 -> 0x742bc18`、`0x12704610 -> 0x742bf7c` 等连续槽位，整体形态很像类的完整 vtable。
3. `0x742e360` 自身的行为仍然稳定：新建 `0x70` 大小对象后，把调用入参 `x1` 所指向的值写进新对象的 `+0x50..+0x68`。
4. 所以当前更准确的判断是：`SelfSignedId` 不是在 `0x742e360` 里现场生成，而是由外部调用者把某个已有 auth/session 状态作为 `x1` 传入，再由这个 vtable 方法包装成缓存 value。

2026-07-20 继续拆 `0x742e360` 下层调用：

1. `0x742e360` 内部会调用 `0x742db54` 和 `0x742c904`：

2026-07-21 proto 任务记录：

1. 按当前接口设计检查 `/root/HonLaderDev/HonLader-api/proto/honlader/api/v1/checkpoint.proto`。
2. `TaskGroupCheckpoint` 已包含：
   - `repeated TaskConfig task_configs = 1`
   - `repeated TaskCheckpoint task_checkpoints = 2`
   - `ServerConfig server = 3`
   - `int32 current_task_index = 4`
3. `ServerConfig` 当前定义在 `checkpoint.proto`，`data_manager.proto` 通过 import `checkpoint.proto` 使用它，避免和 DataManager 请求响应产生重复定义。

2026-07-21 define 同步记录：

1. 只同步源码 `define` 层，暂时未写 DataManagerServer/服务端实现。
2. `/root/HonLaderDev/HonLader/define/config.go` 的 `TaskGroupCheckpoint` 已补齐 proto 对应字段：
   - `Server ServerConfig`
   - `CurrentTaskIndex int`
3. 已运行 `go test ./define`，结果通过。

2026-07-21 TaskFrame 连接配置重构记录：

1. `define.TaskFrame` 已删除：
   - `ApplyClientConfig`
   - `ClientConfig`
   - `WithTaskGroupName`
2. `define.TaskFrame.Connect` 改为 `Connect(ctx context.Context, server ServerConfig) error`。
3. `frame.TaskFrameConfig` 不再嵌入 Core `ClientConfig`，只保留认证相关 `AuthServer`、`UserToken` 和 `Embedded`。
4. `frame.TaskFrame.Connect` 用 `TaskFrameConfig` 的认证信息加上传入的 `define.ServerConfig` 组装 Core `client.FrameConfig`。
5. `frame.Launcher.Connect` 会从 DataManager 读取已保存认证配置，再调用底层 TaskFrame 连接指定服务器。
6. 断点保存不再读取 `ClientConfig()`，改从当前连接使用的 `ServerConfig` 获取服务器信息。
7. 已更新 TUI 和 `cmd/test` 调用点；服务端 gRPC 旧接口引用暂未改。
8. 已运行：
   - `go test ./define ./frame ./cmd/tui/ui`
   - `go test ./cmd/test/...`

2026-07-21 内置验证服务迁移记录：

1. 已将 TUI 内置验证服务从 `cmd/tui/ui/auth_tui/builtin_server.go` 移动到 `frame/auth/builtin_server.go`。
2. 新包名为 `auth`，导出函数保持：
   - `StartBuiltinAuthServer`
   - `StopBuiltinAuthServer`
3. `cmd/tui/ui/auth_tui` 改为引用 `github.com/HonLaderDev/HonLader/frame/auth`。
4. `cmd/test/enter` 改为引用 `frame/auth`，不再依赖 TUI 包启动内置验证服务。
5. 已运行：
   - `go test ./frame/auth ./cmd/tui/ui/auth_tui ./cmd/tui/ui ./cmd/test/enter`

2026-07-21 内置验证服务结构体化记录：

1. `frame/auth` 删除包级 `StartBuiltinAuthServer` / `StopBuiltinAuthServer`。
2. 新增 `AuthServer` 结构体，提供：
   - `Start(ctx context.Context) (string, error)`
   - `Stop()`
3. `cmd/tui/ui/auth_tui` 使用包内单例 `builtinAuthServer frame_auth.AuthServer`。
4. `cmd/test/enter` 改为显式创建 `frame_auth.AuthServer{}` 并调用 `Start/Stop`。
5. 已运行：
   - `go test ./frame/auth ./cmd/tui/ui/auth_tui ./cmd/tui/ui ./cmd/test/enter`

2026-07-21 TaskFrame 配置加载修正：

1. 已删除 `frame.TaskFrame.ApplyAuthConfig`。
2. `define.TaskFrame` 新增 `LoadConfig(config Config) TaskFrame`。
3. `frame.TaskFrameConfig` 不再保存认证字段，只保留创建参数 `Embedded`。
4. `frame.TaskFrame` 内部通过 `LoadConfig(define.Config)` 保存认证配置，`Connect(ctx, define.ServerConfig)` 时再组装 Core `client.FrameConfig`。
5. `frame.Launcher.Connect` 改为 `l.TaskFrame.LoadConfig(config).Connect(ctx, server)`。
6. `cmd/test` 直接创建 TaskFrame 的入口已改为创建后调用 `LoadConfig(define.Config)`。
7. 已运行：
   - `go test ./define ./frame ./cmd/tui/ui`
   - `go test ./cmd/test/...`

2026-07-21 TaskFrame ConnectConfig 修正：

1. 已删除 `define.TaskFrame.LoadConfig` 和 `frame.TaskFrame.LoadConfig`。
2. 新增 `define.ConnectConfig`，包含：
   - `AuthServer`
   - `AuthToken`
   - `ServerCode`
   - `ServerPassword`
3. `define.TaskFrame.Connect` 与 `frame.TaskFrame.Connect` 改为接收 `define.ConnectConfig`。
4. `TaskFrameConfig` 继续只保留创建参数 `Embedded`。
5. `frame.Launcher.Connect` 同步改为接收 `define.ConnectConfig`。
6. TUI 层新增 `cmd/tui/ui/connect.go`，负责把 `define.Config` 和 `define.ServerConfig` 合并为 `define.ConnectConfig`。
7. `cmd/test` 直接创建 TaskFrame 的入口已改为直接传 `define.ConnectConfig`。
8. 已运行：
   - `go test ./define ./frame ./cmd/tui/ui`
   - `go test ./cmd/test/...`
   - `0x742e550: bl 0x742db54`
   - `0x742e5dc: bl 0x742c904`
2. `0x742c904(x0, x1, x2)` 会先复制 `x1` 的多态 wrapper，再分配 `0x40` 对象，并调用 `0x7425368(x0, x1=x19, x2=sp+0x30)`。
3. `0x7425368` 开头保存参数：
   - `x20 = x0`
   - `x21 = x1`
   - `x19 = x2`
4. `0x7425368` 里有关键来源候选：
   - `0x7425424: ldr x0, [x20,#0x40]`
   - 引用计数后 `x23 = [x20,#0x38]`
   - `0x7425478: x0 = x23`
   - `0x7425480: ldr x8, [x0]`
   - `0x7425480/0x7425484: call vtable+0x378`
   - 返回值保存到 `x22`
   - `0x74254bc..0x74254e8` 从 `x22+0x310` 读取一个 string/状态到 `sp+0x30`
5. 这说明 `0x7425368` 会从 `x20` 内部持有的游戏/会话对象经 `vtable+0x378` 取对象，再读该对象 `+0x310` 的数据。这个 `+0x310` 是当前最接近真实来源的字段候选之一。
6. 但还未证明 `x22+0x310` 就是最终 `SelfSignedId`，需要继续追它如何进入 `sp+0xb0` / 后续缓存 value 的 `+0x50..+0x68`。

2026-07-20 继续追 `0x76c8140` 的调用点：

1. 找到直接调用点：`0x633d064: bl 0x76c8140`。
2. `0x633d064` 之前在同一函数里已经把一批局部/成员值整理到栈上：
   - `sp+0x18..0x50`、`sp+0x58..0x70`、`sp+0x80..0xb0`、`sp+0xc0..0xe0`
   - `x26 = [x29+0x98]`、`x25 = [x29+0xa8]`、`x24 = [x29+0xb0]`、`x23 = [x29+0xc0]`、`x20 = [x29+0x88]`
3. 进入 `0x76c8140` 后的关键映射更清楚了：
   - `x29+0x60` -> `str w8, [x19,#0x310]`
   - `x29+0x68` -> 传给 `0x9bbebb8`，结果写到 `x19,#0x318`
   - `x29+0x70` -> 作为输入对象，最终被搬到 `x19,#0x320..#0x330`
4. 所以 `x19,#0x310/#0x318/#0x320` 是同一个大对象的连续字段，内容来自调用者的 3 个栈槽和 1 个输入对象。
5. 目前能确定的是：`SelfSignedId` 不是在 `0x76c8140` 里凭空算出，而是从 `0x633d064` 这层更上游函数准备好的会话/请求数据里搬运出来。

2026-07-20 追到 `0x76c8140` 的调用者 `0x633d064`：

1. `0x633d064` 所在函数是一个很大的组装函数，不是 `MinecraftGame` 直接成员函数；它前面已经构造了多组临时对象：`sp+0x18/0x30/0x40/0x58/0x68/0x80/0x90/0xa0/b0/c0/d0/f0`。
2. 进入 `0x76c8140` 之前，调用者把很多参数放到栈上：
   - `sp+0x18..0x50` 里是一组多态/字符串包装对象指针
   - `sp+0x58..0x70` 里是另一组包装对象
   - `sp+0x80..0xb0` 里还有第三组包装对象
   - `sp+0x320` 旁边的最终大对象在后续会被赋值
3. `0x76c8140` 的输出对象 `x19` 最终收到：

2026-07-20 继续追 `netease_sid` 读取链的修正：

1. `0x76e7c88..0x76e7cc4` 这一段不是单个查询函数，而是在拼一组 key/value：
   - `0x76e7c88: bl 0x9545d98` 取得全局/服务对象。
   - `0x76e7c8c: bl 0x954d05c`，`0x954d05c` 只是 `add x0, x0, #0xe8; ret`，所以返回的是服务对象 `+0xe8` 的字符串/字段。
   - `0x76e7c9c` 使用字符串 `"netease_sid"`。
   - 因此 `netease_sid` 的值候选是 `0x9545d98()` 返回对象的 `+0xe8` 字段。
2. `0x76e7cbc..0x76e7cc4` 后面再次 `bl 0x9545d98`，再调用 `0x954d698`：
   - `0x954d698` 不是查询函数，而是往 `x8` 指向的字符串缓冲写固定短字符串。
   - 它从 `0x2bd08d3` 读 `"3.9.0.296551"`，结合下一段 key `"engineVersion"`，这是 engineVersion 值。
3. 修正当前判断：
   - `netease_sid` 真实值不在 `0x954d698`。
   - 当前最具体来源是：全局 Session/Auth 服务对象 `0x9545d98()` 的 `+0xe8` 字段，经 `0x954d05c` 取出后被打包到 `"netease_sid"` key。

2026-07-20 继续追 `netease_sid` 写入点：

1. `0x9544d04` 是 `0x9545d98()` 返回的全局/服务对象构造函数。
   - 它在 `0x9544dd8: stur q1, [x0,#0xe8]` 把 `+0xe8` 初始化为空 `std::string`。
   - 所以 `netease_sid` 不是构造时生成。
2. `0x954cc60` 是 `+0xe8` 的字符串 setter/copy：
   - `0x954cc60: add x0, x0, #0xe8`
   - 后续按 `std::string` SSO/long-string 两种路径把 `x1` 复制进该字段。
3. `0x954cc60` 的直接调用点只有：
   - `0x94e3438`
4. `0x94e33a0..0x94e3438` 是 `netease_sid` setter 包装：
   - `x20 = x0` 保存调用入参。
   - `0x9545d98()` 取得全局 auth/service 对象到 `x19`。
   - 对 `x20` 调 `0xded3444` 取长度，说明 `x0` 是 C 字符串/字符串视图输入。
   - 把 `x20` 拷贝到栈上临时 `std::string`。
   - `0x94e3430: x1 = sp`
   - `0x94e3434: x0 = x19`
   - `0x94e3438: bl 0x954cc60`
5. 因此 `netease_sid` 当前真实来源已经回溯到：
   - `0x94e33a0` 的入参 `x0`
   - 它不是在 `0x954cc60` 或全局服务对象内部计算出来，而是由外部调用者传入后缓存到 `service+0xe8`。

2026-07-20 继续追 `netease_sid` 的输入/打包路径：

1. `0x94e33a0` 没有普通 `bl` 调用点，`scan_u64`/`scan_adrp_add` 也没找到裸地址引用。
   - 说明它很可能不是普通 C++ 直接调用，而是导出/注册回调/JNI/脚本桥/函数表间接调用。
   - 当前最稳的运行态 hook 点是 `libminecraftpe.so + 0x94e33a0`，打印 `x0` 可直接得到原始传入的 `netease_sid`。
2. `0x9bbfe20` 是另一个 `"netease_sid"` 字段处理函数：
   - `0x9bbfe3c` 引用字符串 `"netease_sid"`。
   - `0x9bbfe80..0x9bbfe90` 把 key/value 交给 `0x9bbf358`。
   - 这更像 JSON/字典字段抽取或打包，不是生成算法。
3. `0x9bbfe20` 的直接调用点：
   - `0x9bc0964`
4. `0x9bc0888..0x9bc0a3c` 是一组连续字段反序列化/抽取：
   - 多次调用 `0x9bbf838`、`0x9bbf6d8`、`0x9bbfb74`、`0x9bbfcd4`、`0x9bbfe20` 等，把结果依次写到 `x19+0x0`、`+0x60`、`+0x78`、`+0x90`、`+0xa8`、`+0x1b0` 等字段。
   - 其中 `0x9bbfe20` 对应 `netease_sid`。
5. 当前结论更新：
   - `netease_sid` 在 native 里有两个角色：一是全局 auth/service 对象 `+0xe8` 的缓存字段；二是某个 JSON/字典模型里的字段。
   - 已确认的真实写入入口是 `0x94e33a0(x0=原始 sid)`。
   - 尚未找到 `0x94e33a0` 的静态直接调用者，下一步应从注册表/桥接层继续找，或运行态 hook `0x94e33a0` 捕捉调用栈。
6. 对 `SelfSignedId` 的影响：
   - 到目前为止，没有看到 `netease_sid` 在 `0x94e33a0`/`0x954cc60` 附近被用于 MD5/UUID/`SelfSignedId` 计算。
   - 更像是：网易登录层先把外部 sid 写进全局服务，SessionAuth/认证包构造时读取这个 sid 打包；`SelfSignedId` 仍来自认证数据/SessionAuth 输入结构，而不是在 `SelfSignedId` JSON 输出点现场计算。
   - `#0x310` 来自调用者栈上的标志位
   - `#0x318` 来自一个额外转换函数输出
   - `#0x320..#0x330` 来自一个调用者传入的字符串/对象包装
4. 这条链再次确认：`SelfSignedId` 对应的字段值不是 `0x76c8140` 里生成，而是从 `0x633d064` 这一层已经准备好的上游会话/请求材料中带下来的。

2026-07-20 继续追 `SessionAuth` / `netease_sid` 线索：

1. `strings` 命中：
   - `netease_sid`
   - `get_auth_netease_sid`
   - `get_auth_netease_sid is not in SERVER thread`
   - `SessionAuthService`
   - `CachedAsyncSource<std::shared_ptr<const SDL::SessionAuth>>`
2. `0x76c8140` 内部确认：
   - 它没有直接在函数本体里写 `x19+0x58`。
   - 但它会在 `0x76c8804..0x76c884c` 读取 `[x19,#0x58]`，调用 `vtable+0x4c0`，再对返回值 `+0x10` 读 `SelfSignedId`。
   - 所以 `[x19,#0x58]` 是真正的 provider / auth 对象入口。
3. `0x76a5a70` 看起来是一个更底层的认证/会话包装构造函数：
   - 它从输入对象 `x28`、`x20` 搬运多组字符串/包装对象。
   - 它会写出一组连续字段到 `x19`，并设置 `+0x150/+0x158/+0x168/+0x170/+0x190/+0x198/+0x1a0/+0x1a8/+0x1b0/+0x1b8` 等。
   - 这条链更像 `SDL::SessionAuth` 或其相关请求封装，而不是 JSON 现场生成。
4. 当前更强判断：`SelfSignedId` 的真实来源应位于 `SessionAuthService::getSessionAuthInternal` / `get_auth_netease_sid` 这条上游异步链，最终把 `netease_sid` 或相关 auth token 填入 `SessionAuth`，再被 `0x76c8140 -> 0x76c8804` 读出。

2026-07-20 进一步收束到 `SessionAuthService`：

1. `dump_va` 直接显示同一字符串块里相邻出现：
   - `CachedAsyncExecuteAsyncResult<std::shared_ptr<const SDL::SessionAuth>>`
   - `(anonymous namespace)::unity_980c4ab3a0d6c52351781f779bc09f95::SessionAuthService::getSessionAuthInternal(int, std::shared_ptr<const Bedrock::Services::EnvironmentQueryResponse>, brstd::optional<std::string>, std::monostate)`
2. 这个 `getSessionAuthInternal` 的模板签名明确告诉我们：
   - 返回结果是 `std::shared_ptr<const SDL::SessionAuth>`
   - 依赖里包含 `std::optional<std::string>`
   - 这正好吻合 `get_auth_netease_sid` / `netease_sid` 作为上游字符串输入的形态
3. 因此当前最硬的静态结论是：
   - `SelfSignedId` 不是微软 JWT 现场算出来的
   - 它来自 `SessionAuthService` 的内部异步缓存结果
   - 上游关键字符串很可能就是 `netease_sid`

2026-07-20 找到更硬的 provider 写入链：

1. `0x76c8140` 内部唯一调用 `0x76c8d94`：
   - `0x76c872c: bl 0x76c8d94`
2. `0x76c8d94` 内部唯一调用 `0x76e1760`：
   - `0x76c8e24: x0 = x21`
   - `0x76c8e28: x1 = x24`
   - `0x76c8e20: x2 = sp+0x10`
   - `0x76c8e2c: x3 = x23`
   - `0x76c8e30: x4 = x22`
   - `0x76c8e34: bl 0x76e1760`
3. `0x76e1760` 是当前最关键的真实组装点：
   - 开头清空输出对象 `+0x8..+0x58`
   - `0x76e17d0: str xzr, [x0,#0x58]`
   - 如果 `x1` 里有多态对象，会调用其虚表复制到 `x19+0x50`
   - 随后 `0x76e17f8: str x8, [x19,#0x50]`
   - 因为 `x19+0x58` 位于这个 provider wrapper 内部，所以最终被 `0x76c8804` 读取的 provider 来自 `0x76e1760` 的 `x1`
4. 参数回溯：
   - `0x76e1760.x1 = 0x76c8d94.x24`
   - `0x76c8d94.x24 = 0x76c8d94.x0`
   - `0x76c8724: x0 = x29-0x90`
   - `0x76c81e0: stp x28,x27,[x29,#-0x90]`
   - `x28,x27` 来自 `0x76c8170: ldp x28,x27,[x29,#0xb8]`
   - 以 `0x76c8140` 的栈帧换算，`[x29,#0xb8]` 对应调用者 `0x633d064` 之前写的 `[caller_sp,#0x58]`
   - `0x633d050: stp x12,x10,[sp,#0x58]`
   - `x12,x10` 来自 `0x633cff4: ldp x12,x10,[x29,#0xb0]`
5. 所以 provider / `SelfSignedId` 的输入不是 `0x76c8140` 本地生成，而是 `0x633ceb0` 的上游栈参数 `[x29,#0xb0]` 传入，再一路进入 `0x76e1760.x1`，最后落在输出对象 `+0x50/+0x58`。

2026-07-20 `netease_sid` 的静态引用定位：

1. 完整字符串 xref：
   - `getSessionAuthInternal` 函数名字符串 `0x2bcdc0a` 只在 `0x7e1d238` 被引用。
   - `get_auth_netease_sid is not in SERVER thread` 只在 `0x933f070` 被引用。
   - `netease_sid` 字符串 `0x28433b0` 被 `0x76e7c9c`、`0x9bbfe3c`、`0xa77f220` 引用。
2. `0x933efc0..0x933f120` 是脚本/命令 API 的 `get_auth_netease_sid` 包装：
   - 先检查是否在 SERVER thread。
   - 不在 SERVER thread 时使用 `get_auth_netease_sid is not in SERVER thread` 打日志/报错。
   - 这不是 `SessionAuth` 的主生成链。
3. `0x76e7c88..0x76e7cc4` 更像主链：
   - `0x76e7c88: bl 0x9545d98`
   - `0x76e7c8c: bl 0x954d05c`
   - 用返回值构造 `sp+0x140`
   - `0x76e7c9c: x1 = "netease_sid"`
   - `0x76e7ca8..0x76e7cb8` 把 key/value 组合进临时字符串/字典对象
   - `0x76e7cbc: bl 0x9545d98`
   - `0x76e7cc4: bl 0x954d698`
4. 因此 `netease_sid` 的真实读取点当前收束到：
   - 服务对象/客户端对象：`0x9545d98` 返回
   - getter/查询函数：`0x954d698`
   - 读取 key：`"netease_sid"`

2026-07-20 16:41 纠偏：回到 `SelfSignedId` 主链：

1. `netease_sid` 目前只作为旁支记录，不能继续当作 `SelfSignedId` 的已证明来源。
   - 除非重新接回 `SelfSignedId provider` 的硬证据链，否则不要继续扩展它。
2. 当前 `SelfSignedId` 主链断点仍是 provider 输入 pair：
   - `0x76c8140 -> 0x76c872c -> 0x76c8d94 -> 0x76e1760`
   - `0x76e1760.x1` 最终落到后续 JSON 输出会读取的 provider wrapper `+0x50/+0x58`
   - `0x76e1760.x1 = 0x76c8d94.x24 = 0x76c8d94.x0 = 0x76c8140` 局部 pair
   - 该 pair 来自调用者 `0x633ceb4` 的 `[x29,#0xb0]`
3. 重新核对 `0x633ca00 -> 0x633ceb4` 参数后：
   - `0x633c9f4: stp x22,x23,[sp,#0x50]`
   - 进入 `0x633ceb4` 后对应 `[x29,#0xb0]`
   - 所以 `SelfSignedId provider` 输入 pair = `x22,x23`
4. `x22,x23` 来源：
   - `0x633c968: ldr x8,[x20,#0xb0]`
   - `0x633c974: ldr x9,[vtable,#0x740]`
   - `0x633c97c: blr x9`
   - `0x633c980: mov x22,x0`
   - `0x633c984: ldr x0,[x20,#0xb0]`
   - `0x633c98c: ldr x8,[vtable,#0x6f8]`
   - `0x633c990: blr x8`
   - `0x633c994: mov x23,x0`
5. 结论：下一步必须识别 `x20+0xb0` 对象，以及它虚表槽 `+0x740` 和 `+0x6f8` 的实现/字段含义。

2026-07-20 继续回溯 `x20+0xb0`：

1. `0x633c67c` 是 `0x633ca00 -> 0x633ceb4 -> 0x76c8140` 之前的上游组装函数。
   - 函数开头 `0x633c6b0: mov x20,x0`
   - 所以 `0x633c968` 处的 `x20+0xb0` 是该函数第一个参数对象的成员。
2. `0x633c67c` 直接调用点：
   - `0x633dd40`
   - `0x633de9c`
3. `0x633dd14` 和 `0x633de38` 都是包装层，调用 `0x633c67c` 时第一个参数仍是外部传入的 `x0`。
4. 再上一层 `0x633b970`：
   - 函数开头 `0x633b994: mov x19,x0`
   - `0x633b9a8: ldr x0,[x0,#0xb0]`
   - `0x633b9b0: ldr x9,[vtable,#0x1e8]`
   - `0x633b9c0..0x633b9c8` 对 `vtable+0x1e8` 返回对象再调用 `vtable+0x4c0`
   - 后面还会对同一个 `x19+0xb0` 调 `vtable+0x758`
5. 这说明 `+0xb0` 不是普通字符串字段，而是主对象内的核心服务/认证 provider/异步源对象。`SelfSignedId` 的 pair 最终来自这个对象的虚表槽：
   - `+0x740 -> x22`
   - `+0x6f8 -> x23`
6. `0x633b970` 直接调用点：
   - `0x633b4cc`
   - `0x645cc10`

2026-07-20 找到主对象 `+0xb0` 写入点：

1. `0x633b190` 没有普通 `bl` 调用点，`0x12606990 -> 0x633b190` 是 `std::__function::__func<ui::OreUIScreenConfiguration...>` 的函数对象表，不是主对象 vtable；这条不要误认成主类。
2. 真正和主对象 `+0xb0` 相关的构造/初始化函数是 `0x6360284`：
   - `0x63602bc: mov x19,x0`，目标主对象为 `x0`
   - `0x63602a4: mov x20,x5`
   - `0x6360374: str x20,[x19,#0xb0]`
3. 因此 `SelfSignedId` provider 的服务对象来源继续上推为：
   - `main_obj+0xb0 = 0x6360284.x5`
4. 下一步：追 `0x6360284` 的调用点，确认 `x5` 是从哪里传入的。

2026-07-20 修正 `0x6360284`：

1. `0x6360284` 虽然有 `str x20,[x19,#0xb0]`，但它构造的是另一个 `0xb8` 小结构。
2. 调用点显示 `x5` 可能是 `0`、解析出的整数、或临时值：
   - `0x6360054` 的 `x5 = sxtw(w0)`，来自字符串数字解析。
   - `0x6e8399c` / `0x7d29568` 直接 `x5 = xzr`。
3. 因此这个 `+0xb0` 不是 `0x633c968` 处的 `x20+0xb0` 服务对象；只是同偏移、不同类型。
4. 这条路排除，不再作为 `SelfSignedId` 主链来源。
5. 回到正确断点：
   - `0x633c968` 处 `x20+0xb0` 必须是带虚表的对象。
   - 继续应识别它的 vtable/RTTI，尤其是 `SessionAuth` / `CachedAsyncSource<std::shared_ptr<const SDL::SessionAuth>>` 相关类型。

2026-07-20 17:07 纠偏继续：只追 `SelfSignedId` 主链：

1. 重新确认 `0x633c900..0x633ca00`：
   - `0x633c968: ldr x8, [x20,#0xb0]`
   - `0x633c974: ldr x9, [vtable,#0x740]`
   - `0x633c97c: blr x9`
   - `0x633c980: mov x22, x0`
   - `0x633c984: ldr x0, [x20,#0xb0]`
   - `0x633c98c: ldr x8, [vtable,#0x6f8]`
   - `0x633c990: blr x8`
   - `0x633c994: mov x23, x0`
   - `0x633c998..0x633c9a4` 还会调用同一对象 `vtable+0x678`
2. 随后 `0x633c9f4: stp x22,x23,[sp,#0x50]`，进入 `0x633ceb4` 后在 `0x633d050` 转发到 `0x76c8140` 的栈参数，最终走到 `0x76e1760.x1`，成为后续 JSON 输出读取 `SelfSignedId` 的 provider。
3. 因此当前硬断点不是 `netease_sid`，而是识别 `x20+0xb0` 服务对象及其 `+0x740/+0x6f8/+0x678` 槽的含义。

2026-07-20 17:10 重新校验 provider wrapper 语义：

1. `0x76e1760` 的入口参数更准确：
   - `x0` = 输出对象。
   - `x1` = provider wrapper 输入。
   - `x2/x3/x4/x5/x6` = 其它上下文字段/flag。
2. `0x76e1760` 开头清空输出 `+0x8..+0x58`，随后：
   - `0x76e17d4: ldr x8, [x1]`
   - 若 `[x1]` 非空，则调用其虚函数，把内容复制到 `x19+0x50`。
   - `0x76e17f8: str x8, [x19,#0x50]`
   - 后续还会搬 `x1+0x40/+0x48/+0x50...` 到输出 `+0x90/+0x98/+0xa0...`
3. 所以从 `0x633c968` 过来的 `x22/x23` 不是简单的 shared_ptr 二元组；`+0x740` 的返回值是 provider wrapper 的主体输入，`+0x6f8` 至少在 `0x7717af4` 另一个调用点被当作 bool 使用。
4. `0x6349c80` 是一个转发包装：
   - 先取 `[x0,#0xb0]`
   - 调该对象 `vtable+0x740`
   - 再跳到 `0x7ab81f4`
   - `0x7ab81f4` 只是 `add x0,x0,#0x58; ret`
   - 说明 `+0x740` 返回的对象/结构中 `+0x58` 是一个重要子字段。

2026-07-20 17:18 追到 `SessionAuth` payload 构造层：

1. `getSessionAuthInternal` 字符串引用点：
   - `0x7e1d238` 引用完整签名字符串。
   - `0x7e1d2a4 -> 0x7e1ef38` 是 `CachedAsyncSource` 执行/缓存包装。
   - `0x7e1d160 -> 0x7e1e2e4` 会分配 `0x230` 字节异步结果对象。
2. `0x7e1e2e4` 内部调用 `0x7e1e4b0`，`0x7e1e4b0` 最终：
   - 整理 `x1..x7` 和一个栈参数。
   - `0x7e1e5e4: bl 0x7dd164c`
3. `0x7dd164c` 是目前最像 `SDL::SessionAuth` / session auth payload 的构造器：
   - 设置 vtable 为 `0x12760158`。

2026-07-20 继续纠偏：只追 `SelfSignedId`：

1. 重新检查 `0x7425368`：
   - 它会从 `x20+0x38/+0x40` 持有的对象取出对象，再调用该对象 `vtable+0x378`。
   - 返回对象记为 `x22`。
   - `0x74254bc..0x74254e8` 从 `x22+0x310/+0x318/+0x320` 组装一个字符串/状态到 `sp+0x30`。
   - 这个值后续进入 provider/cache 包装，但目前不能直接命名为 `SelfSignedId`；只能说它是 `SelfSignedId provider` 上游输入之一。
2. 继续拆 `getSessionAuthInternal` 路径：
   - `0x7e1e4b0` 整理参数后在 `0x7e1e5e4: bl 0x7dd164c`。
   - `0x7dd164c` 设置对象 vtable：
     - `0x7dd1668..0x7dd1674`: `x0[0] = 0x12760ee0` 初始 vtable。
     - `0x7dd16b8..0x7dd16d0`: 随后切到 `0x12760158`。
   - `llvm-readelf -r` 证明 `0x12760158 -> 0x7dd1acc`，说明这是一个真实 vtable/函数表入口。
3. `0x7dd164c` 字段搬运：
   - `x2` 的 pair/string 被搬到 `SessionAuth +0x70/+0x80`。
   - `x3` 的 vector/列表被搬到 `SessionAuth +0x88`。
   - 栈参数 `[x29+0x60]` 的多态对象被搬到 `SessionAuth +0xc0`。
   - `x4` 字符串/optional 被搬到 `SessionAuth +0xd0/+0xe0`。
   - `x5` 字符串/optional 被搬到 `SessionAuth +0xe8/+0xf8`。
   - `x7` 在 optional 置位时搬到 `SessionAuth +0x100/+0x110`，并设置 `+0x118`。
   - `x6` 的 shared/状态 pair 被搬到 `SessionAuth +0x1d0/+0x1e0`。
4. 当前判断：
   - `SelfSignedId` 仍未证明是这里某个字段本身。
   - 但 `SessionAuth` 构造器是上游真实来源候选，因为它位于 `getSessionAuthInternal`，且其结果正是 `std::shared_ptr<const SDL::SessionAuth>`。
   - 下一步必须把 `main_obj+0xb0` 服务对象的 `vtable+0x740` 返回值接到 `SessionAuth` / `0x12760158` 这一侧，或排除这条连接。
   - 先调用 `0x7dfbe18` 初始化对象前半段。
   - 再搬入多组字符串/optional 字段：
     - `x26 -> object+0xd0/+0xe0`
     - `x25 -> object+0xe8/+0xf8`
     - `x23 -> optional object+0x100/+0x110, flag +0x118`
     - `x21 -> object+0x1d0/+0x1e0...`
4. `0x7dfbe18` 不是字段生成器，而是基类/头部初始化：
   - 调 `0x78f7dc0(x0, x1, 0x5a)`。
   - 然后写 `+0x64/+0x68`。
5. `0x78f7dc0` 是通用基类初始化器：
   - `+0x0` vtable
   - `+0x10` 保存 `x1`
   - `+0x20` 保存 type/id `0x5a`
   - `+0x28/+0x30/+0x38/+0x48/+0x60` 初始化通用成员
   - 异常/析构路径会释放 `+0x58`
6. 当前不要把 `+0x58` 的所有通用写点误判为 `SelfSignedId`。硬证据是：
   - `+0x740` 返回对象。
   - `0x7ab81f4` 暴露这个返回对象的 `+0x58`。
   - 需要继续确认 `0x7dd164c` 构造的这个对象中 `+0x58` 由哪条参数/回调写入。

2026-07-20 17:27 纠偏：当前任务只追 `SelfSignedId`：

1. 不再扩展 `netease_sid`、`IdentityData.Identity`、UUIDv3、`pocket-auth-1-xuid:`，除非它们重新接回 `SelfSignedId provider` 的硬链。
2. 当前硬链重新收束为：
   - `0x633c968`: `x8 = [x20,#0xb0]`
   - `0x633c974..0x633c97c`: 调服务对象 `vtable+0x740`
   - `0x633c980`: `x22 = +0x740` 返回值
   - `0x633c984..0x633c990`: 调同对象 `vtable+0x6f8`
   - `0x633c994`: `x23 = +0x6f8` 返回值
   - `0x633c9f4`: `stp x22,x23,[sp,#0x50]`
   - `0x633d050`: `stp x12,x10,[sp,#0x58]`
   - `0x76c8170`: `ldp x28,x27,[x29,#0xb8]`
   - `0x76c81e0`: `stp x28,x27,[x29,#-0x90]`
   - `0x76c8720`: `x0 = x29-0x90`
   - `0x76c8d94`: `x24 = x0`
   - `0x76c8e28`: `x1 = x24`
   - `0x76c8e34`: `bl 0x76e1760`
3. `0x76c8d94` 已核实只是转发：`x0` 原样作为 `0x76e1760.x1`，没有生成 `SelfSignedId`。
4. 因此现在唯一要破的是：
   - `x20+0xb0` 的对象类型/vtable。
   - 该对象 `vtable+0x740` 的实现如何生成/返回 provider wrapper。
5. `SessionAuthService` 构造区 `0x7df4800..0x7df49bc` 发现：
   - `0x7df4894` 引用字符串 `SessionAuthService`。
   - `0x7df48e8` 引用函数对象表 `0x12766fa8`。
   - `0x7df48f4` 引用函数对象表 `0x12767018`。
   - `0x7df49bc -> 0x7e49734`，疑似创建/注册 `CachedAsyncSource<std::shared_ptr<const SDL::SessionAuth>>` 相关成员。
6. 下一步继续拆 `0x7e49734` 及 `0x7e49b70/0x7e49b84`，目标是识别 `x20+0xb0` 服务对象或其 async source 成员。

2026-07-20 继续追 `SelfSignedId` 主链的新结论：

1. `0x7e49734` 只有一个直接调用点 `0x7df49bc`。
   - 返回值 `x23` 后续写入 `SessionAuthService +0x270`。
   - 同时创建 `0x20` 小闭包写入 `SessionAuthService +0x278`。
2. `SessionAuthService +0x270/+0x278` 是 async/cached source 风格成员：
   - `0x7df4d6c`: 读 `+0x278`
   - `0x7df4d70`: 读 `+0x270`
   - `0x7df5e04`: 读 `+0x270` 后跳其 `vtable+0x18`
   - 析构路径 `0x7e4b170` 会释放 `+0x278`
3. 这条链目前说明 `SessionAuthService` 维护 session auth source，但还没有直接证明它就是 `0x633c968` 的 `x20+0xb0` 服务对象。
4. 重新检查 `0x633b970/0x633c67c` 后，发现更贴近 `SelfSignedId` 的 auth 包读法：
   - `0x633c67c` 开头直接调用主对象 `x0->vtable+0x4c0`，返回 auth 包到 `x26`。
   - `0x633c6dc`: `x0 = x26`
   - `0x633c6e0 -> 0x76e8708`
   - `0x76e8708`: `add x0,#0x10; b 0xa77f940`
   - `0xa77f940`: `add x0,#0x40`
5. 同组 getter：
   - `0xa77f940`: auth data `+0x40`
   - `0xa77f948`: auth data `+0x48`
   - `0xa77f950`: auth data `+0x60`，这就是 `SelfSignedId`
6. 所以 `0x633c67c` 已经在处理同一个 auth data 结构，只是它这里取的是 `+0x40` 相邻字段；最终 JSON 输出取 `+0x60`。
7. 当前最硬结论：
   - `SelfSignedId` 不是 `netease_sid`/UUIDv3 现场生成。
   - 它位于 auth package 的内部 auth data `+0x60`。
   - auth package 来自主对象 `vtable+0x4c0` 或主对象 `+0xb0` 服务返回的 provider wrapper。
   - 下一步应识别主对象 `vtable+0x4c0` 和 `+0xb0` 服务 `vtable+0x740` 的真实实现/字段来源。

2026-07-20 主对象 `vtable+0x4c0` 闭合：

1. `0x6349c80` 反汇编：
   - `ldr x0,[x0,#0xb0]`
   - 取服务对象 vtable
   - 调 `vtable+0x740`
   - 跳 `0x7ab81f4`
2. `0x7ab81f4` 是 `add x0,x0,#0x58; ret`。
3. `0x6349c80` 的数据引用：
   - `scan_u64 0x6349c80 -> va=0x95aaf8` 是文件内未重定位命中。
   - `llvm-readelf -r` 给出真实 relocation：`0x12606c10 -> 0x6349c80`。
   - `0x12606c10 - 0x4c0 = 0x12606750`，因此该 vtable 的 `+0x4c0` 槽就是 `0x6349c80`。
4. 这证明 `0x633c67c` 中的主对象 `vtable+0x4c0` 确实等价于：
   - `main_obj + 0xb0`
   - 调 service `vtable+0x740`
   - 返回该结果对象的 `+0x58`
5. 因此最终 JSON 中 `SelfSignedId` 的硬链可以写成：
   - `main_obj+0xb0` 服务对象
   - `service->vtable+0x740`
   - 返回对象 `+0x58`
   - auth package
   - `auth package +0x10`
   - auth data `+0x60`
   - JSON writer `"SelfSignedId"`
6. 新增临时工具 `tmp/revtools/scan_pair_offset.c` 用于查 `stp/ldp` 覆盖指定偏移。
   - 扫 `+0xb0` 后发现大多数命中是 `x31`，即栈槽 `sp`，不是对象字段。
   - 尚未找到清晰的 `main_obj+0xb0` 普通构造写入点；它可能由更外层依赖注入/移动构造进入。

2026-07-20 `SyncActorPropertyPacket` 逆向结果：

1. `libminecraftpe.so` 字符串直接显示：
   - `SyncActorPropertyPacket`
   - `SyncActorPropertyPacketPayload`
   - `PropertyData`
2. 该 packet 在 native 里对应的字段只有一个核心字段：
   - `PropertyData map[string]any`
3. 序列化方式是 NBT：
   - `io.NBT(&pk.PropertyData, nbt.NetworkLittleEndian)`
4. 结论：
   - `SyncActorPropertyPacket` 不是离散字段包。
   - 它本质上就是一个 NBT 字典包，字段名为 `PropertyData`。
