# EnvSwitch - 环境管理切换工具设计文档

## 1. 项目概述

EnvSwitch 是一个用于管理多项目、多环境配置的命令行工具，支持快速切换不同环境的配置文件。

### 1.1 核心功能
- **项目管理**：创建、列出、查看、删除项目
- **环境管理**：在项目下创建、修改、删除环境配置
- **文件切换**：根据环境配置替换系统中的指定文件
- **Web界面**：提供HTTP服务，支持Web端管理

### 1.2 技术栈
- **后端**：Go 1.21+
- **CLI框架**：cobra
- **Web框架**：gin
- **数据存储**：JSON文件 + SQLite（可选）
- **前端**：HTML + CSS + JavaScript（简单Web界面）

## 2. 数据结构设计

### 2.1 项目（Project）
```go
type Project struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Environments []Environment `json:"environments"`
}
```

### 2.2 环境（Environment）
```go
type Environment struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Tags        []string  `json:"tags"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    LastSwitchAt *time.Time `json:"last_switch_at,omitempty"`
    Files       []FileConfig `json:"files"`
}
```

### 2.3 文件配置（FileConfig）
```go
type FileConfig struct {
    ID          string `json:"id"`
    SourcePath  string `json:"source_path"`  // 模板文件路径
    TargetPath  string `json:"target_path"`  // 目标替换路径
    BackupPath  string `json:"backup_path"`  // 备份文件路径
    Description string `json:"description"`
}
```

### 2.4 配置管理（Config）
```go
type Config struct {
    DataDir     string `json:"data_dir"`
    BackupDir   string `json:"backup_dir"`
    WebPort     int    `json:"web_port"`
    DefaultProject string `json:"default_project"`
}
```

## 3. 目录结构

```
EnvSwitch/
├── cmd/                    # 命令行入口
│   ├── root.go            # 根命令
│   ├── project.go         # 项目管理命令
│   ├── env.go             # 环境管理命令
│   ├── switch.go          # 切换命令
│   └── server.go          # Web服务命令
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── storage/          # 数据存储
│   ├── project/          # 项目逻辑
│   ├── environment/      # 环境逻辑
│   ├── file/            # 文件操作
│   └── web/             # Web服务
├── web/                  # Web界面资源
│   ├── static/          # 静态文件
│   └── templates/       # HTML模板
├── data/                # 数据存储目录
└── backups/             # 备份目录
```

## 4. CLI命令设计

### 4.1 项目管理命令
```bash
# 创建项目
envswitch project create <name> [--description="描述"]

# 列出所有项目
envswitch project list

# 查看项目详情
envswitch project show <name>

# 删除项目
envswitch project delete <name>

# 设置默认项目
envswitch project set-default <name>
```

### 4.2 环境管理命令
```bash
# 创建环境
envswitch env create <project> <env-name> [--description="描述"] [--tags="tag1,tag2"]

# 列出环境
envswitch env list [project]

# 查看环境详情
envswitch env show <project> <env-name>

# 修改环境
envswitch env update <project> <env-name> [--description="新描述"] [--tags="tag1,tag2"]

# 删除环境
envswitch env delete <project> <env-name>

# 添加文件配置
envswitch env add-file <project> <env-name> <source> <target> [--description="描述"]

# 移除文件配置
envswitch env remove-file <project> <env-name> <file-id>
```

### 4.3 切换命令
```bash
# 切换到指定环境
envswitch switch <project> <env-name>

# 快速切换（使用默认项目）
envswitch switch <env-name>

# 查看当前环境状态
envswitch status

# 回滚到切换前状态
envswitch rollback
```

### 4.4 Web服务命令
```bash
# 启动Web服务
envswitch server [--port=8080]

# 后台启动Web服务
envswitch server --daemon [--port=8080]
```

## 5. API设计

### 5.1 REST API端点

#### 项目相关
- `GET /api/projects` - 获取所有项目
- `POST /api/projects` - 创建项目
- `GET /api/projects/{id}` - 获取项目详情
- `PUT /api/projects/{id}` - 更新项目
- `DELETE /api/projects/{id}` - 删除项目

#### 环境相关
- `GET /api/projects/{project-id}/environments` - 获取项目的所有环境
- `POST /api/projects/{project-id}/environments` - 创建环境
- `GET /api/environments/{id}` - 获取环境详情
- `PUT /api/environments/{id}` - 更新环境
- `DELETE /api/environments/{id}` - 删除环境

#### 文件配置相关
- `POST /api/environments/{env-id}/files` - 添加文件配置
- `PUT /api/files/{id}` - 更新文件配置
- `DELETE /api/files/{id}` - 删除文件配置

#### 切换相关
- `POST /api/switch` - 切换环境
- `GET /api/status` - 获取当前状态
- `POST /api/rollback` - 回滚

### 5.2 WebSocket
- `/ws` - 实时状态更新和日志推送

## 6. 数据存储

### 6.1 文件存储结构
```
data/
├── config.json           # 全局配置
├── projects/            # 项目数据
│   ├── project1.json
│   └── project2.json
└── state.json          # 当前状态信息
```

### 6.2 备份策略
- 每次切换前自动备份当前文件
- 备份文件按时间戳命名
- 支持回滚到任意备份点
- 定期清理过期备份

## 7. 安全考虑

### 7.1 文件操作安全
- 文件路径验证，防止路径遍历攻击
- 权限检查，确保有足够权限操作目标文件
- 原子操作，确保文件替换的原子性

### 7.2 Web服务安全
- CSRF防护
- 输入验证和消毒
- 访问控制（可选的认证机制）

## 8. 错误处理

### 8.1 常见错误场景
- 文件不存在或无权限访问
- 磁盘空间不足
- 配置文件格式错误
- 网络服务启动失败

### 8.2 错误恢复
- 自动回滚机制
- 详细的错误日志
- 用户友好的错误消息

## 9. 性能优化

### 9.1 文件操作优化
- 增量备份
- 文件变更检测
- 并发处理多个文件

### 9.2 Web服务优化
- 静态资源缓存
- API响应缓存
- 连接池管理

## 10. 扩展性

### 10.1 插件系统
- 支持自定义文件处理器
- 支持第三方集成

### 10.2 配置模板
- 预定义的环境模板
- 配置导入导出功能

## 11. 测试策略

### 11.1 单元测试
- 核心逻辑的单元测试
- Mock外部依赖

### 11.2 集成测试
- 端到端的切换流程测试
- API集成测试

### 11.3 性能测试
- 大量文件切换的性能测试
- 并发访问的压力测试 