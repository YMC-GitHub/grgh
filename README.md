# grgh - Git发布工具

**一键迁移文件并推送到GitHub**，支持智能同步、文档处理、强制推送等功能，专为简化发布流程设计。

## ✨ 核心功能
1. **`migrate`** 迁移发布文件到临时目录
2. **`clone`** 克隆远程仓库到临时目录
3. **`push`** Git推送（需先执行migrate或clone）
4. **`clean`** 清理临时发布目录
5. **`sync`** 同步流程：智能处理 + 迁移 + 推送（一键完成）
6. **`docs`** 处理文档文件（多语言README重命名）

## 📌 完整参数说明

| 参数 | 说明 | 默认值 | 环境变量 |
|------|------|--------|----------|
| --cmd | 命令名称（必需） | - | - |
| --gh-name | Git用户名称 | YeMincheng | GRGH_NAME |
| --gh-email | Git用户邮箱 | ymc.github@gmail.com | GRGH_EMAIL |
| --gh-repo | GitHub仓库 | ymc-github/pdfi | GRGH_REPO |
| --gh-msg | 提交信息 | init: add all files | - |
| --gh-branch | Git分支名称 | main | - |
| --clean | 清理临时目录 | true | - |
| --force | 强制推送 | false | - |
| --ws | 工作目录路径 | 自动生成 | - |
| --release-cmd | 发布命令名称 | 自动提取 | - |
| --extra-files | 额外文件或目录 | - | - |
| --exclude-dirs | 排除的目录 | .git,.github,node_modules,dist,build,target | - |
| --use-lang | 指定默认语言 | zh | - |
| -h, --help | 显示帮助信息 | - | - |
| -v, --version | 显示版本信息 | - | - |

## 🚀 最常用命令示例

### 1. 一键同步（推荐）
```bash
# 使用环境变量配置
export GRGH_NAME="YourName"
export GRGH_EMAIL="your@email.com"
export GRGH_REPO="YourGitHub/your-repo"

# 执行同步
grgh --cmd sync --use-lang zh
```

### 2. 分步操作
```bash
# 先迁移文件
grgh --cmd migrate --gh-repo "YourGitHub/your-repo"

# 再推送
grgh --cmd push --gh-msg "feat: update files"

# 清理临时目录
grgh --cmd clean
```

### 3. 强制推送（谨慎使用）
```bash
grgh --cmd push --force --gh-msg "force push"
```

### 4. 自定义配置
```bash
grgh --cmd sync \
  --gh-repo "YourGitHub/your-repo" \
  --gh-name "YourName" \
  --gh-email "your@email.com" \
  --gh-branch "master" \
  --gh-msg "release: new version" \
  --use-lang en
```

### 5. 处理文档文件
```bash
# 将 README.zh.md 设为 README.md
grgh --cmd docs --ws ./release_tmp/your-repo --use-lang zh
```

## 📁 项目结构说明

工具期望的源文件结构：
```
项目根目录/
├── cmd/
│   └── {release-cmd}/     # 命令源码
├── ghwf/
│   └── {release-cmd}/     # GitHub workflows配置
├── docs/
│   └── {release-cmd}/     # 文档文件（.md）
├── Dockerfile.{release-cmd}
├── LICENSE
├── LICENSE-APACHE
├── LICENSE-MIT
├── .Dockerignore
├── .editorconfig
└── .gitignore
```

## 🎯 同步流程详解

`sync` 命令会自动执行以下步骤：

1. **智能仓库处理**
   - 如果本地目录存在且是git仓库 → 执行 `git pull`
   - 如果本地目录不存在 → 执行 `git clone`
   - 如果克隆失败 → 创建新仓库并初始化

2. **迁移文件**
   - 复制命令源码到 `cmd/{release-cmd}/`
   - 复制GitHub workflows到 `.github/workflows/`
   - 复制文档文件到根目录
   - 复制配置文件（Dockerfile、LICENSE等）

3. **文档处理**
   - 根据 `--use-lang` 参数选择语言文件
   - 将 `README.{lang}.md` 复制为 `README.md`

4. **Git提交与推送**
   - 配置用户信息
   - 添加所有文件
   - 提交变更
   - 推送到远程仓库

## 🐳 环境变量配置

推荐使用环境变量避免重复输入：

```bash
# 设置环境变量（永久配置）
echo 'export GRGH_NAME="YourName"' >> ~/.bashrc
echo 'export GRGH_EMAIL="your@email.com"' >> ~/.bashrc
echo 'export GRGH_REPO="YourGitHub/your-repo"' >> ~/.bashrc
source ~/.bashrc

# 或临时设置
export GRGH_NAME="YourName"
export GRGH_EMAIL="your@email.com"
export GRGH_REPO="YourGitHub/your-repo"
```

**优先级**：命令行参数 > 环境变量 > 默认值

## 🔧 高级用法

### 排除特定目录
```bash
grgh --cmd sync --exclude-dirs ".git,.github,node_modules,dist,logs,temp"
```

### 添加额外文件
```bash
grgh --cmd sync --extra-files "./config,.env,./scripts"
```

### 自定义工作目录
```bash
grgh --cmd migrate --ws "/tmp/my-release" --gh-repo "YourGitHub/repo"
```

### 不清理临时目录（便于调试）
```bash
grgh --cmd push --clean=false
```

## 📊 命令速查表

| 命令 | 说明 | 是否需要ws | 需要网络 |
|------|------|-----------|----------|
| sync | 一键同步（推荐） | 自动 | ✅ |
| migrate | 只迁移文件 | 自动 | ❌ |
| clone | 只克隆仓库 | 自动 | ✅ |
| push | 只推送变更 | ✅ | ✅ |
| docs | 只处理文档 | ✅ | ❌ |
| clean | 只清理目录 | 自动 | ❌ |

## ⚠️ 注意事项

1. **首次使用前**请配置环境变量或使用命令行参数
2. **强制推送** (`--force`) 会覆盖远程历史，谨慎使用
3. **工作目录**默认创建在 `./release_tmp/{仓库名}/`
4. **SSH密钥**需要预先配置好GitHub SSH访问权限
5. **仓库格式**支持：
   - 简洁格式：`username/repo`（自动转换为SSH）
   - SSH格式：`git@github.com:username/repo.git`
   - HTTPS格式：`https://github.com/username/repo.git`

## 🛡️ 安全特性

- 自动排除 `.git` 目录避免冲突
- 推送前检查文件变更，避免空提交
- 支持强制推送确认机制
- 临时目录自动清理（可选关闭）

## 📝 常见问题

**Q: 提示权限拒绝？**  
A: 检查GitHub SSH密钥配置：`ssh -T git@github.com`

**Q: 推送失败怎么办？**  
A: 先执行 `grgh --cmd clean` 清理，再重新 `sync`

**Q: 如何查看详细日志？**  
A: 工具会实时输出所有命令的执行结果

**Q: 支持Windows吗？**  
A: 支持，但建议使用Git Bash或WSL环境
