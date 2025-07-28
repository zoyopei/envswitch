# EnvSwitch - 环境管理切换工具

[![CI](https://github.com/zoyopei/EnvSwitch/workflows/CI/badge.svg)](https://github.com/zoyopei/EnvSwitch/actions/workflows/ci.yml)
[![Release](https://github.com/zoyopei/EnvSwitch/workflows/Release/badge.svg)](https://github.com/zoyopei/EnvSwitch/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoyopei/EnvSwitch)](https://goreportcard.com/report/github.com/zoyopei/EnvSwitch)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/zoyopei/EnvSwitch.svg)](https://github.com/zoyopei/EnvSwitch/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/ghcr.io/zoyopei/envswitch)](https://github.com/zoyopei/EnvSwitch/pkgs/container/envswitch)

一个用Go语言实现的环境管理切换命令行工具，支持多项目、多环境配置管理，可以快速切换不同环境的配置文件。同时提供Web界面进行可视化管理。

## 🚀 功能特性

### 核心功能
- **项目管理**：创建、列出、查看、删除项目
- **环境管理**：在项目下创建、修改、删除环境配置
- **文件切换**：根据环境配置替换系统中的指定文件
- **Web界面**：提供HTTP服务，支持Web端管理
- **备份恢复**：自动备份原文件，支持一键回滚
- **安全保障**：文件操作前验证权限，确保安全性

### 技术特点
- **轻量级**：单一二进制文件，无外部依赖
- **跨平台**：支持Windows、Linux、macOS
- **数据持久化**：JSON文件存储，简单可靠
- **RESTful API**：完整的API接口
- **实时状态**：WebSocket支持实时状态更新

## 📦 安装

### 推荐：一键安装脚本 (Linux/macOS)
```bash
curl -sfL https://github.com/zoyopei/EnvSwitch/releases/latest/download/install.sh | sh
```

### 下载预编译二进制文件
从 [Releases](https://github.com/zoyopei/EnvSwitch/releases) 页面下载适合您系统的二进制文件。

### 使用 Docker
```bash
# 拉取镜像
docker pull ghcr.io/zoyopei/envswitch:latest

# 运行容器
docker run -d -p 8080:8080 \
  -v $(pwd)/data:/home/envswitch/data \
  -v $(pwd)/backups:/home/envswitch/backups \
  ghcr.io/zoyopei/envswitch:latest
```

### 从源码构建
```bash
git clone https://github.com/zoyopei/EnvSwitch.git
cd EnvSwitch
go mod tidy
go build -o envswitch .
```

### 使用 Go Install
```bash
go install github.com/zoyopei/EnvSwitch@latest
```

## 🔧 快速开始

### 1. 初始化配置

```bash
# 首次运行会自动创建配置文件
envswitch project list
```

### 2. 创建项目

```bash
# 创建一个新项目
envswitch project create myapp --description="我的应用项目"

# 设置为默认项目
envswitch project set-default myapp
```

### 3. 创建环境

```bash
# 在项目中创建开发环境
envswitch env create myapp dev --description="开发环境" --tags="development,local"

# 创建生产环境
envswitch env create myapp prod --description="生产环境" --tags="production"
```

### 4. 添加文件配置

```bash
# 为开发环境添加配置文件
envswitch env add-file myapp dev ./config/dev.json ./app/config.json --description="开发配置文件"

# 为生产环境添加配置文件
envswitch env add-file myapp prod ./config/prod.json ./app/config.json --description="生产配置文件"
```

### 5. 切换环境

```bash
# 切换到开发环境
envswitch switch myapp dev

# 或者使用默认项目（如果已设置）
envswitch switch dev

# 查看当前状态
envswitch status
```

### 6. 启动Web服务

```bash
# 启动Web界面（默认端口8080）
envswitch server

# 指定端口
envswitch server --port 9090

# 后台运行
envswitch server --daemon
```

然后在浏览器中访问 `http://localhost:8080`

## 📋 CLI命令参考

### 项目管理

```bash
# 创建项目
envswitch project create <name> [--description="描述"]

# 列出所有项目
envswitch project list

# 查看项目详情
envswitch project show <name>

# 删除项目
envswitch project delete <name> [--force]

# 设置默认项目
envswitch project set-default <name>
```

### 环境管理

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
envswitch env delete <project> <env-name> [--force]

# 添加文件配置
envswitch env add-file <project> <env-name> <source> <target> [--description="描述"]

# 移除文件配置
envswitch env remove-file <project> <env-name> <file-id>
```

### 环境切换

```bash
# 切换到指定环境
envswitch switch <project> <env-name>

# 快速切换（使用默认项目）
envswitch switch <env-name>

# 预览模式（不实际执行）
envswitch switch <env-name> --dry-run

# 查看当前环境状态
envswitch status

# 回滚到切换前状态
envswitch rollback [backup-id] [--force]
```

### Web服务

```bash
# 启动Web服务
envswitch server [--port=8080] [--daemon]
```

## 🌐 Web API

### 项目相关
- `GET /api/projects` - 获取所有项目
- `POST /api/projects` - 创建项目
- `GET /api/projects/{id}` - 获取项目详情
- `PUT /api/projects/{id}` - 更新项目
- `DELETE /api/projects/{id}` - 删除项目

### 环境相关
- `GET /api/projects/{project-id}/environments` - 获取项目的所有环境
- `POST /api/projects/{project-id}/environments` - 创建环境
- `GET /api/environments/{id}` - 获取环境详情
- `PUT /api/environments/{id}` - 更新环境
- `DELETE /api/environments/{id}` - 删除环境

### 切换相关
- `POST /api/switch` - 切换环境
- `GET /api/status` - 获取当前状态
- `POST /api/rollback` - 回滚

## 📁 目录结构

```
envswitch/
├── cmd/                    # CLI命令实现
│   ├── root.go            # 根命令
│   ├── project.go         # 项目管理命令
│   ├── env.go             # 环境管理命令
│   ├── switch.go          # 切换命令
│   └── server.go          # Web服务命令
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── storage/          # 数据存储
│   ├── project/          # 项目逻辑
│   ├── file/            # 文件操作
│   └── web/             # Web服务
├── web/                  # Web界面资源
│   ├── static/          # 静态文件
│   └── templates/       # HTML模板
├── data/                # 数据存储目录
├── backups/             # 备份目录
└── config.json          # 配置文件
```

## ⚙️ 配置文件

配置文件位置：
- 当前目录：`./config.json`
- 用户目录：`~/.envswitch/config.json`

默认配置：
```json
{
  "data_dir": "data",
  "backup_dir": "backups",
  "web_port": 8080,
  "default_project": ""
}
```

## 🔒 安全考虑

### 文件操作安全
- 文件路径验证，防止路径遍历攻击
- 权限检查，确保有足够权限操作目标文件
- 原子操作，确保文件替换的原子性
- 自动备份，切换前备份原文件

### Web服务安全
- CSRF防护
- 输入验证和消毒
- 可配置的访问控制

## 🛠️ 开发

### 环境要求
- Go 1.21+
- Git

### 本地开发

```bash
# 克隆项目
git clone https://github.com/your-org/envswitch.git
cd envswitch

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 开发模式运行
go run main.go

# 构建
go build -o envswitch
```

### 目录说明
- `cmd/` - CLI命令定义
- `internal/` - 核心业务逻辑
- `web/` - Web界面相关文件
- `docs/` - 项目文档

## 📝 使用示例

### 管理多个Node.js环境

```bash
# 创建项目
envswitch project create webapp --description="Web应用项目"

# 创建环境
envswitch env create webapp dev --description="开发环境"
envswitch env create webapp test --description="测试环境"
envswitch env create webapp prod --description="生产环境"

# 添加package.json配置
envswitch env add-file webapp dev ./configs/dev/package.json ./package.json
envswitch env add-file webapp test ./configs/test/package.json ./package.json
envswitch env add-file webapp prod ./configs/prod/package.json ./package.json

# 添加环境变量文件
envswitch env add-file webapp dev ./configs/dev/.env ./.env
envswitch env add-file webapp test ./configs/test/.env ./.env
envswitch env add-file webapp prod ./configs/prod/.env ./.env

# 切换到开发环境
envswitch switch webapp dev

# 切换到生产环境
envswitch switch webapp prod
```

### 管理数据库配置

```bash
# 创建数据库项目
envswitch project create database --description="数据库配置管理"

# 创建环境
envswitch env create database local --description="本地数据库"
envswitch env create database staging --description="预发布数据库"
envswitch env create database production --description="生产数据库"

# 添加数据库配置文件
envswitch env add-file database local ./db-configs/local.conf ./etc/database.conf
envswitch env add-file database staging ./db-configs/staging.conf ./etc/database.conf
envswitch env add-file database production ./db-configs/production.conf ./etc/database.conf

# 切换数据库环境
envswitch switch database production
```

## 🛠️ 开发

### 本地开发环境设置
```bash
git clone https://github.com/zoyopei/EnvSwitch.git
cd EnvSwitch
go mod download

# 运行应用
go run . --help

# 启动开发服务器
go run . server --port 8080
```

### 构建
```bash
# 本地构建
make build

# 交叉编译
make cross-compile

# 使用 Go 直接构建
go build -o envswitch .
```

### 运行测试
```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 直接使用 Go
go test ./...
```

### 代码质量检查
```bash
# 运行代码检查
make lint

# 格式化代码
make format
```

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建 feature 分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 贡献指南

- 请确保所有测试通过
- 遵循 Go 代码规范
- 添加适当的单元测试
- 更新相关文档
- Pull Request 应包含清晰的描述

详细信息请查看 [CONTRIBUTING.md](CONTRIBUTING.md)。

### CI/CD

本项目使用 GitHub Actions 进行持续集成和部署：

- **CI**: 自动运行测试、代码检查和构建
- **Release**: 自动构建多平台二进制文件并发布
- **CodeQL**: 安全代码分析
- **Docker**: 自动构建和推送 Docker 镜像

## 🔒 安全

### 漏洞报告

如果您发现安全漏洞，请不要在公开的 GitHub Issues 中报告。请发送邮件至：[security@example.com](mailto:security@example.com)

### 安全特性

- 文件路径验证防止目录遍历攻击
- 自动备份机制防止数据丢失
- 配置文件权限检查
- Web 界面 CSRF 保护（计划中）

## 🗺️ 路线图

- [ ] 支持配置文件模板
- [ ] 添加环境变量管理
- [ ] 支持远程配置存储
- [ ] 集成更多第三方工具
- [ ] 添加插件系统
- [ ] GUI 桌面应用

## 📄 许可证

本项目基于 MIT 许可证开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 鸣谢

- [Cobra](https://github.com/spf13/cobra) - CLI 框架
- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [Viper](https://github.com/spf13/viper) - 配置管理

## 📞 支持

- 📚 [文档](https://github.com/zoyopei/EnvSwitch/wiki)
- 🐛 [问题反馈](https://github.com/zoyopei/EnvSwitch/issues)
- 💬 [讨论](https://github.com/zoyopei/EnvSwitch/discussions)
- 📫 [邮件支持](mailto:support@example.com)

---

如果您觉得这个项目有用，请给它一个 ⭐️！ 