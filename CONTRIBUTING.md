# 贡献指南

感谢您对 EnvSwitch 项目的兴趣！我们欢迎各种形式的贡献，包括但不限于：

- 🐛 Bug 报告
- 💡 功能建议
- 📝 文档改进
- 🧪 代码贡献
- 🔍 代码审查

## 开始之前

在开始贡献之前，请确保您已经：

1. 阅读了 [README.md](README.md) 文件
2. 查看了现有的 [Issues](https://github.com/zoyopei/EnvSwitch/issues) 和 [Pull Requests](https://github.com/zoyopei/EnvSwitch/pulls)
3. 了解了项目的代码结构和约定

## 开发环境设置

### 系统要求

- Go 1.21 或更高版本
- Git
- 支持 Make 的系统（可选）

### 设置步骤

1. **Fork 并克隆仓库**
   ```bash
   git clone https://github.com/YOUR_USERNAME/EnvSwitch.git
   cd EnvSwitch
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **验证安装**
   ```bash
   go run . --help
   ```

4. **运行测试**
   ```bash
   go test ./...
   ```

## 代码规范

### Go 代码风格

我们遵循标准的 Go 代码风格和最佳实践：

- 使用 `gofmt` 格式化代码
- 使用 `goimports` 管理导入
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html) 指导原则
- 使用 `golangci-lint` 进行代码检查

### 代码检查

在提交代码前，请运行以下命令：

```bash
# 格式化代码
go fmt ./...
goimports -w .

# 运行代码检查
golangci-lint run

# 运行测试
go test -v ./...
```

### 提交信息规范

我们使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**类型（Type）：**
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响代码运行的变动）
- `refactor`: 重构（既不是新增功能，也不是修改bug的代码变动）
- `test`: 增加测试
- `chore`: 构建过程或辅助工具的变动

**示例：**
```
feat(cli): add new environment switch command

Add support for switching environments with a single command.
This includes validation of environment names and automatic
backup creation.

Closes #123
```

## 测试指南

### 测试类型

1. **单元测试** (`*_test.go`)
   - 测试单个函数或方法
   - 使用 Go 标准的 `testing` 包

2. **集成测试** (`test/integration_test.go`)
   - 测试多个组件的集成
   - 测试 API 端点

3. **端到端测试** (`test/e2e_test.go`)
   - 测试完整的用户流程
   - 包括 CLI 和 Web 界面测试

### 编写测试

- 为新功能添加相应的测试
- 确保测试覆盖边界情况
- 使用有意义的测试名称
- 添加必要的测试文档

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行单元测试
go test ./internal/...

# 运行集成测试
go test ./test/ -run="TestAPI"

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Pull Request 流程

1. **创建 Issue**（可选但推荐）
   - 对于重大更改，建议先创建 Issue 讨论

2. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **开发和测试**
   - 编写代码
   - 添加/更新测试
   - 确保所有测试通过

4. **提交更改**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **推送分支**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **创建 Pull Request**
   - 填写 PR 模板
   - 提供清晰的描述
   - 链接相关的 Issues

7. **代码审查**
   - 响应审查意见
   - 根据反馈修改代码

8. **合并**
   - 在审查通过后，维护者将合并 PR

## Pull Request 检查清单

在提交 PR 之前，请确保：

- [ ] 代码通过所有测试
- [ ] 新功能有相应的测试
- [ ] 代码通过 linting 检查
- [ ] 更新了相关文档
- [ ] 提交信息遵循规范
- [ ] PR 描述清晰
- [ ] 没有合并冲突

## 报告 Bug

### Bug 报告模板

请使用 [Bug 报告模板](.github/ISSUE_TEMPLATE/bug_report.md) 创建详细的 Bug 报告。

### 包含的信息

- 清晰的问题描述
- 重现步骤
- 预期行为
- 实际行为
- 环境信息（OS、Go 版本等）
- 相关的日志或截图

## 功能请求

### 功能请求模板

请使用 [功能请求模板](.github/ISSUE_TEMPLATE/feature_request.md) 提交新功能建议。

### 提供的信息

- 功能描述
- 使用场景
- 期望的解决方案
- 备选方案
- 实现建议

## 文档贡献

我们重视文档的质量，欢迎以下类型的文档贡献：

- 修复错别字和语法错误
- 改进现有文档的清晰度
- 添加使用示例
- 翻译文档
- 添加 API 文档

## 社区

### 讨论

- [GitHub Discussions](https://github.com/zoyopei/EnvSwitch/discussions) - 一般讨论和问题
- [Issues](https://github.com/zoyopei/EnvSwitch/issues) - Bug 报告和功能请求

### 行为准则

请遵循 [Contributor Covenant](https://www.contributor-covenant.org/) 行为准则。我们致力于营造一个开放、包容的社区环境。

## 发布流程

项目使用自动化的发布流程：

1. 创建带有版本号的 Git 标签（如 `v1.0.0`）
2. GitHub Actions 自动构建多平台二进制文件
3. 创建 GitHub Release
4. 构建和推送 Docker 镜像

## 获得帮助

如果您在贡献过程中遇到问题，请：

1. 查看现有的 Issues 和 Discussions
2. 创建新的 Discussion 询问问题
3. 发送邮件至 [support@example.com](mailto:support@example.com)

## 致谢

感谢所有为 EnvSwitch 项目做出贡献的开发者！

### 贡献者

<!-- ALL-CONTRIBUTORS-LIST:START -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

---

再次感谢您的贡献！🎉 