# wx_channel 定制化合并策略文档

> 此文档用于指导 AI 助手自动处理上游合并，减少人工干预。

---

## 1. 项目基本信息

| 项目 | 地址 |
|------|------|
| 上游仓库 (upstream) | `https://github.com/nobiyou/wx_channel.git` |
| 我的 Fork (origin) | `https://github.com/longit123/wx_channel.git` |
| 定制分支 | `custom` |
| 基准分支 | `main` |

---

## 2. 合并流程

### 标准合并步骤
```bash
# 1. 获取上游更新
git fetch upstream

# 2. 尝试自动合并
git merge upstream/main --no-edit

# 3. 如有冲突，按本文档规则处理
# 4. 推送到我的 Fork
git push origin custom
```

---

## 3. 定制化修改记录

### 3.1 新增文件（无冲突风险）

> 这些文件是我们新增的，上游不会有，合并时自动保留。

| 文件路径 | 用途 | 状态 |
|----------|------|------|
| `pkg/sunnynet/Resource/nfapi/dll/win32/nfapi.dll` | Windows 32位网络驱动DLL（从Go模块缓存复制） | 新增 |
| `pkg/sunnynet/Resource/nfapi/dll/x64/nfapi.dll` | Windows 64位网络驱动DLL（从Go模块缓存复制） | 新增 |
| `.trae/MERGE_STRATEGY.md` | 合并策略文档 | 新增 |
| `internal/services/audio_extractor.go` | 音频提取服务（m4a 纯Go抽轨 / mp3 ffmpeg转码） | 新增 |
| `internal/api/audio.go` | 音频能力查询 API（`/api/v1/audio/capabilities`） | 新增 |

**说明**：上游 `.gitignore` 忽略了 `*.dll` 文件，导致源码构建时缺少这些 DLL。我们从 Go 模块缓存复制到本地，解决 Windows 源码构建问题。

### 3.2 修改文件（可能冲突）

> 这些文件我们修改了上游代码，需要记录修改策略。

| 文件路径 | 修改内容摘要 | 冲突策略 | 状态 |
|----------|--------------|----------|------|
| `internal/database/database.go` | 将 SQLite 驱动从 `mattn/go-sqlite3` 改为 `modernc.org/sqlite`（纯Go实现） | **保留本地** | 已修改 |
| `go.mod` | 移除 `mattn/go-sqlite3`（避免符号冲突）；新增 `gomedia` direct 依赖（音频抽轨） | **人工确认** | 已修改 |
| `go.sum` | 随 go.mod 变化自动更新 | **自动处理** | 已修改 |
| `internal/app/app.go` | 创建 AudioExtractor 单例，注入 BatchHandler 和 APIRouter | **保留本地** | 已修改 |
| `internal/handlers/batch.go` | BatchTask 加音频字段，下载完成后提取音频钩子 | **保留本地** | 已修改 |
| `internal/router/api_routes.go` | 注册音频能力查询 API 路由 | **保留本地** | 已修改 |
| `internal/router/router.go` | `isPublicAPIPath` 白名单加 `/api/v1/audio/capabilities` | **保留本地** | 已修改 |
| `internal/database/models.go` | DownloadRecord 加 `AudioPath`/`AudioFormat` 字段 | **保留本地** | 已修改 |
| `internal/database/migrations.go` | v15 migration 加音频字段 | **保留本地** | 已修改 |
| `internal/database/download_repository.go` | SQL 读写音频字段 | **保留本地** | 已修改 |
| `internal/assets/inject/batch_download.js` | 批量下载面板加「同时下载音频」复选框与格式下拉 | **保留本地** | 已修改 |
| `internal/config/config.go` | `cloud_enabled`/`hub_sync` 默认改为 false，避免向第三方推送数据 | **保留本地** | 已修改 |
| `internal/cloud/connector.go` | `cloud_enabled` 为 false 时 Start 早返回 | **保留本地** | 已修改 |
| `config.yaml.example` | 云端管理默认禁用示例 | **保留本地** | 已修改 |
| `config.yaml.full` | 云端管理默认禁用完整示例 | **保留本地** | 已修改 |

**说明**：
- 上游同时使用了 `mattn/go-sqlite3`（CGO）和 `glebarez/sqlite`（纯Go），在 Windows MinGW 环境下会产生 SQLite 符号冲突
- 我们统一使用纯 Go SQLite 驱动，解决 Windows 源码构建问题
- 如果上游更新 go.mod，需要人工确认是否保留我们的修改
- 音频提取功能（2026-06-29 新增）：m4a 用纯 Go gomedia 抽轨或 ffmpeg copy，mp3 需 ffmpeg 转码；前端通过 `/api/v1/audio/capabilities` 查询 ffmpeg 是否可用以决定默认格式
- 云端默认禁用（2026-06-29）：上游默认启用云端推送，我方改为默认关闭，避免未经用户同意向第三方服务器推送浏览/下载记录

### 3.3 配置文件（自动保留本地）

> 这些配置文件冲突时，自动保留我们的版本。

| 文件路径 | 用途 | 状态 |
|----------|------|------|
| `.env` | 环境配置 | 自动保留本地 |
| `config/*.yaml` | 自定义配置 | 自动保留本地 |

---

## 4. 冲突处理策略

### 策略优先级

| 策略 | 说明 | 适用场景 |
|------|------|----------|
| **保留本地 (ours)** | 冲突时使用我们的版本 | 配置文件、定制功能 |
| **使用上游 (theirs)** | 冲突时使用上游版本 | 通用 bugfix、性能优化 |
| **人工确认** | 需要用户判断 | 功能性改动、逻辑变更 |
| **智能合并** | 尝试结合双方 | 可兼容的改动 |

### 自动处理规则

以下情况 **无需人工确认**，AI 直接处理：

1. **新增文件冲突** → 保留本地（上游不会有我们的新文件）
2. **纯配置冲突** → 保留本地
3. **注释/文档冲突** → 使用上游（保持文档最新）
4. **相同修改** → 自动合并（Git 会处理）

以下情况 **需要人工确认**：

1. **核心逻辑冲突** → 列出差异，用户选择
2. **API 接口变更** → 可能影响依赖项目
3. **依赖版本冲突** → go.mod / package.json

---

## 5. 冲突处理记录

> 记录每次冲突的解决方案，供下次参考。

| 日期 | 文件 | 冲突内容 | 解决方案 | 备注 |
|------|------|----------|----------|------|
| 2026-06-18 | `internal/database/database.go` | SQLite 驱动名 `sqlite3` 应改为 `sqlite` | 保留本地修改 | `modernc.org/sqlite` 注册名是 `sqlite`，不是 `sqlite3` |
| 2026-06-18 | `internal/database/database_test.go` | 同上 | 同上 | 测试文件同步修改 |
| 2026-06-27 | `go.mod` | 上游删除 `glebarez/sqlite`/`gorm`/`gorilla/mux`（随 hub_server 移除），保留 `mattn/go-sqlite3`；我方需保留 `modernc.org/sqlite` 纯 Go 方案 | 保留本地：go.mod direct 区只保留 `modernc.org/sqlite`，移除 `mattn/go-sqlite3`/`gorm.io/gorm`，其余 indirect 交由 `go mod tidy` 整理 | 上游已无符号冲突根因（glebarez 删除），但我方延续纯 Go 构建策略；用户确认保留方案 A |
| 2026-06-27 | `go.sum` | 随 go.mod 变化 | 自动合并 + `go mod tidy` | 无冲突标记 |
| 2026-06-27 | `rsrc_windows_386.syso` / `rsrc_windows_amd64.syso` / `wx_channel.exe` | modify/delete：我方在 `9d68216` 删除构建产物，上游更新了这些二进制 | 保持删除（`git rm`） | 沿用我方「移除构建产物」决策，不接受上游二进制 |
| 2026-06-27 | `hub_server/` 整个目录 | 上游 `828b24b chore: remove migrated hub server` 删除整个 hub_server | 接受上游删除 | 无冲突（我方未改动该目录内文件）；物理残留由 `git clean -fd` 清理 |
| 2026-06-29 | 多文件（见 3.1/3.2 节） | 新增音频提取功能 + 云端默认禁用 | 分两个 commit 提交至 `custom` 分支 | 非冲突解决，为功能开发记录。音频功能：2 新增 + 8 修改 + `go.mod`/`go.sum`；云端禁用：4 修改 |

---

## 6. 依赖项目关联

> 本项目被以下项目依赖，合并时需考虑兼容性。

| 项目名称 | 依赖方式 | 关注点 |
|----------|----------|--------|
| （待填写） | go.mod replace | API 兼容性 |

---

## 7. 快速指令参考

```bash
# 查看当前状态
git status
git branch -a
git remote -v

# 同步上游
git fetch upstream
git log --oneline upstream/main -5  # 查看上游最近更新

# 合并上游
git merge upstream/main --no-edit

# 查看冲突
git diff --name-only --diff-filter=U

# 解决冲突后
git add .
git commit -m "merge upstream"
git push origin custom

# 回滚合并（如果出问题）
git merge --abort
```

---

## 8. AI 助手处理流程

当用户发送此文档并请求合并时，AI 应按以下流程处理：

```
1. git fetch upstream
2. git log --oneline upstream/main -10  → 分析上游更新内容
3. git merge upstream/main --no-edit    → 尝试自动合并

4. 如果无冲突 → git push origin custom → 完成

5. 如果有冲突：
   a. git diff --name-only --diff-filter=U → 获取冲突文件列表
   b. 根据本文档第3节规则，自动处理可自动处理的冲突
   c. 对于需要人工确认的冲突：
      - 列出冲突详情（diff 格式）
      - 提供选项：[保留本地] [使用上游] [智能合并]
      - 等待用户确认
   d. 用户确认后，执行 git checkout --ours/--theirs 或手动编辑
   e. git add . && git commit && git push

6. 更新本文档第5节（冲突处理记录）
```

---

## 9. 功能开发提交规范与验证要点

> 功能开发后的提交规范，避免重蹈覆辙。

### 9.1 提交流程

1. 改代码
2. 更新本文件第 3.1（新增文件）/ 3.2（修改文件）/ 第 5 节（提交记录）
3. **`git diff` 验证文档确实写入工作区**（不能只看 Edit 工具返回的成功提示）
4. commit 后用 **`Select-String` 验证 commit 内容**（见 9.2）
5. 分功能 commit + push（一个功能一个 commit，不混入无关改动）

### 9.2 验证工具选择（Windows）

| 工具 | 可靠性 | 说明 |
|------|--------|------|
| `git diff` | ✅ 可靠 | 验证工作区 vs 暂存区/commit 的差异，首选 |
| `Select-String` | ✅ 可靠 | PowerShell 内置，正确处理 UTF-8 中文 |
| `findstr` | ❌ 不可靠 | 对 UTF-8 中文内容匹配失败，**禁止用于含中文的文件** |

### 9.3 经验教训记录

| 日期 | 教训 | 改进措施 |
|------|------|----------|
| 2026-06-29 | Edit 工具返回"成功"但内容未真正进入工作区/commit；`findstr` 误报 commit 内容缺失导致误判 | 提交前必须 `git diff` 验证写入；验证 commit 内容用 `Select-String` 而非 `findstr` |
| 2026-06-29 | Read 工具显示版本号为 1.4，实际磁盘仍为 1.3（显示与磁盘不一致） | 关键字段以 `git diff`/`git show` 为准，不轻信 Read 的显示内容 |

---

## 附录：待填写事项

> 随着定制化开发进行，请及时更新以下内容：

- [x] 第3.1节：新增文件列表
- [x] 第3.2节：修改文件列表及策略
- [ ] 第6节：依赖项目信息

---

*文档版本: 1.4*
*创建日期: 2026-06-15*
*最后更新: 2026-06-29*