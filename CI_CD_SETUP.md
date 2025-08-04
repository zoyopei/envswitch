# GitHub CI/CD 配置总结

## 📁 项目结构

EnvSwitch 项目现已配置完整的 GitHub CI/CD 流水线，包含以下组件：

```
EnvSwitch/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml              # 持续集成工作流
│   │   ├── release.yml         # 发版工作流
│   │   └── codeql.yml          # 代码安全分析
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md       # Bug报告模板
│   │   └── feature_request.md  # 功能请求模板
│   ├── PULL_REQUEST_TEMPLATE.md # PR模板
│   └── dependabot.yml          # 依赖自动更新
├── scripts/
│   └── install.sh              # 一键安装脚本
├── .gitignore                 # Git忽略文件
├── .golangci.yml              # Go代码检查配置
├── Makefile                   # 构建脚本
├── LICENSE                    # MIT许可证
├── CONTRIBUTING.md            # 贡献指南
└── README.md                  # 项目说明（已更新）
```

## 🔄 CI/CD 工作流程

### 1. 持续集成 (CI) - `.github/workflows/ci.yml`

**触发条件：**
- Push 到 `main` 或 `develop` 分支
- 创建 Pull Request 到 `main` 或 `develop` 分支

**包含任务：**
- ✅ **多版本测试**: Go 1.21.x 和 1.22.x
- ✅ **依赖缓存**: 加速构建过程
- ✅ **单元测试**: 运行 `./internal/...` 测试
- ✅ **集成测试**: 运行 API 和 Web 页面测试
- ✅ **代码检查**: 使用 golangci-lint
- ✅ **安全扫描**: 使用 Gosec
- ✅ **跨平台构建**: Linux、Windows、macOS (AMD64/ARM64)
- ✅ **测试覆盖率**: 自动上传到 Codecov

### 2. 发版流程 (Release) - `.github/workflows/release.yml`

**触发条件：**
- 推送带有 `v*` 格式的 Git 标签（如 `v1.0.0`）

**自动化任务：**
- ✅ **预发版测试**: 确保代码质量
- ✅ **多平台构建**: 
  - Linux (AMD64/ARM64)
  - Windows (AMD64)
  - macOS (AMD64/ARM64)
- ✅ **版本信息注入**: 在二进制文件中嵌入版本号
- ✅ **创建压缩包**: tar.gz (Linux/macOS) 和 zip (Windows)
- ✅ **生成校验和**: SHA256 校验文件
- ✅ **自动发布**: 创建 GitHub Release

### 3. 代码安全分析 (CodeQL) - `.github/workflows/codeql.yml`

**触发条件：**
- Push 到主分支
- Pull Request
- 每周一自动扫描

**分析内容：**
- ✅ **Go代码分析**: 检测安全漏洞
- ✅ **JavaScript分析**: Web界面代码检查
- ✅ **SARIF报告**: 详细的安全分析报告

## 🚀 发版流程

### 创建新版本

1. **准备发布**
   ```bash
   # 确保代码已提交并推送
   git add .
   git commit -m "feat: prepare for v1.0.0 release"
   git push origin main
   ```

2. **创建版本标签**
   ```bash
   # 创建带注释的标签
   git tag -a v1.0.0 -m "Release version 1.0.0

   ### New Features
   - Complete environment management system
   - Web interface for project management
   - CLI tools for environment switching
   - Automatic backup and rollback functionality

   ### Improvements  
   - Enhanced error handling
   - Improved documentation
   - Better test coverage"

   # 推送标签到远程仓库
   git push origin v1.0.0
   ```

3. **自动化流程**
   - GitHub Actions 自动检测到标签推送
   - 运行完整的测试套件
   - 构建多平台二进制文件
   - 发布 GitHub Release

### 版本命名规范

遵循 [语义化版本](https://semver.org/lang/zh-CN/) (SemVer):

- `v1.0.0` - 主版本.次版本.修订版本
- `v1.0.0-alpha.1` - 预发布版本
- `v1.0.0-beta.1` - 测试版本

## 📦 发布产物

每次发版会自动生成以下文件：

### 二进制文件
- `envswitch-v1.0.0-linux-amd64.tar.gz`
- `envswitch-v1.0.0-linux-arm64.tar.gz`
- `envswitch-v1.0.0-darwin-amd64.tar.gz`
- `envswitch-v1.0.0-darwin-arm64.tar.gz`
- `envswitch-v1.0.0-windows-amd64.zip`

### 其他文件
- `checksums.txt` - SHA256 校验和
- `install.sh` - 一键安装脚本

## 🔧 本地开发工作流

### 使用 Makefile

```bash
# 查看所有可用命令
make help

# 构建项目
make build

# 运行测试
make test

# 生成测试覆盖率报告
make test-coverage

# 代码检查
make lint

# 格式化代码
make format

# 跨平台构建
make cross-compile

# 创建发布包
make release
```

### 使用 Go 命令

```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 构建项目
go build -o envswitch .

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o envswitch-linux-amd64 .
```

## 🛡️ 质量保证

### 自动化检查
- **代码覆盖率**: 目标 > 80%
- **静态分析**: golangci-lint 检查
- **安全扫描**: Gosec 安全分析
- **依赖检查**: Dependabot 自动更新

### 测试策略
- **单元测试**: 测试独立组件
- **集成测试**: 测试组件交互
- **端到端测试**: 测试完整流程
- **性能测试**: 基准测试

## 📋 Issue 和 PR 管理

### Issue 模板
- **Bug 报告**: 结构化的错误报告
- **功能请求**: 新功能建议模板

### PR 检查清单
- [ ] 所有测试通过
- [ ] 代码覆盖率不降低
- [ ] 遵循代码规范
- [ ] 更新相关文档
- [ ] 添加适当的测试

## 🔄 依赖管理

### Dependabot 配置
- **Go模块**: 每周一自动检查更新
- **GitHub Actions**: 每周一检查新版本

### 手动更新
```bash
# 更新所有依赖
go get -u ./...
go mod tidy

# 检查过期依赖
go list -u -m all
```

## 🚦 状态徽章

README.md 中包含以下状态徽章：

- [![CI](https://github.com/zoyopei/EnvSwitch/workflows/CI/badge.svg)](https://github.com/zoyopei/EnvSwitch/actions/workflows/ci.yml)
- [![Release](https://github.com/zoyopei/EnvSwitch/workflows/Release/badge.svg)](https://github.com/zoyopei/EnvSwitch/actions/workflows/release.yml)
- [![Go Report Card](https://goreportcard.com/badge/github.com/zoyopei/EnvSwitch)](https://goreportcard.com/report/github.com/zoyopei/EnvSwitch)
- [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 💡 最佳实践

### 提交信息规范
使用 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/):

```
feat: 添加新功能
fix: 修复错误
docs: 更新文档
style: 代码格式
refactor: 重构代码
test: 添加测试
chore: 其他更改
```

### 分支策略
- `main`: 主分支，稳定版本
- `develop`: 开发分支，新功能集成
- `feature/*`: 功能分支
- `hotfix/*`: 紧急修复分支

### 安全最佳实践
- 不在代码中硬编码敏感信息
- 使用 GitHub Secrets 管理密钥
- 定期更新依赖以修复安全漏洞
- 启用分支保护规则

## 📞 支持

如果您在使用 CI/CD 流程中遇到问题：

1. 查看 [GitHub Actions 日志](https://github.com/zoyopei/EnvSwitch/actions)
2. 阅读 [CONTRIBUTING.md](CONTRIBUTING.md)
3. 创建 [Issue](https://github.com/zoyopei/EnvSwitch/issues)
4. 参与 [Discussions](https://github.com/zoyopei/EnvSwitch/discussions)

---

**配置完成时间**: 2025年7月29日  
**配置版本**: v1.0.0  
**维护者**: [@zoyopei](https://github.com/zoyopei) 