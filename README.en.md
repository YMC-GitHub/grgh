# grgh - Git Release Tool

**One-click file migration and push to GitHub**, supporting smart synchronization, document processing, force push, and more, designed to simplify release workflows.

## ✨ Core Features
1. **`migrate`** Migrate release files to temporary directory
2. **`clone`** Clone remote repository to temporary directory
3. **`push`** Git push (requires migrate or clone first)
4. **`clean`** Clean up temporary release directory
5. **`sync`** Sync workflow: smart handling + migration + push (all-in-one)
6. **`docs`** Process document files (multi-language README renaming)

## 📌 Complete Parameters

| Parameter | Description | Default | Environment Variable |
|-----------|-------------|---------|---------------------|
| --cmd | Command name (required) | - | - |
| --gh-name | Git user name | YeMincheng | GRGH_NAME |
| --gh-email | Git user email | ymc.github@gmail.com | GRGH_EMAIL |
| --gh-repo | GitHub repository | ymc-github/pdfi | GRGH_REPO |
| --gh-msg | Commit message | init: add all files | - |
| --gh-branch | Git branch name | main | - |
| --clean | Clean up temporary directory | true | - |
| --force | Force push to remote | false | - |
| --ws | Working directory path | Auto-generated | - |
| --release-cmd | Release command name | Auto-extracted | - |
| --extra-files | Extra files or directories | - | - |
| --exclude-dirs | Directories to exclude | .git,.github,node_modules,dist,build,target | - |
| --use-lang | Default language for README | zh | - |
| -h, --help | Show help information | - | - |
| -v, --version | Show version information | - | - |

## 🚀 Most Common Command Examples

### 1. One-click Sync (Recommended)
```bash
# Configure with environment variables
export GRGH_NAME="YourName"
export GRGH_EMAIL="your@email.com"
export GRGH_REPO="YourGitHub/your-repo"

# Execute sync
grgh --cmd sync --use-lang en
```

### 2. Step-by-step Operations
```bash
# Migrate files first
grgh --cmd migrate --gh-repo "YourGitHub/your-repo"

# Then push
grgh --cmd push --gh-msg "feat: update files"

# Clean up temporary directory
grgh --cmd clean
```

### 3. Force Push (Use with caution)
```bash
grgh --cmd push --force --gh-msg "force push"
```

### 4. Custom Configuration
```bash
grgh --cmd sync \
  --gh-repo "YourGitHub/your-repo" \
  --gh-name "YourName" \
  --gh-email "your@email.com" \
  --gh-branch "master" \
  --gh-msg "release: new version" \
  --use-lang en
```

### 5. Process Document Files
```bash
# Set README.en.md as README.md
grgh --cmd docs --ws ./release_tmp/your-repo --use-lang en
```

## 📁 Project Structure Expected

The tool expects the following source file structure:
```
Project Root/
├── cmd/
│   └── {release-cmd}/     # Command source code
├── ghwf/
│   └── {release-cmd}/     # GitHub workflows config
├── docs/
│   └── {release-cmd}/     # Documentation files (.md)
├── Dockerfile.{release-cmd}
├── LICENSE
├── LICENSE-APACHE
├── LICENSE-MIT
├── .Dockerignore
├── .editorconfig
└── .gitignore
```

## 🎯 Sync Workflow Details

The `sync` command automatically executes the following steps:

1. **Smart Repository Handling**
   - If local directory exists and is a git repository → execute `git pull`
   - If local directory doesn't exist → execute `git clone`
   - If clone fails → create new repository and initialize

2. **File Migration**
   - Copy command source to `cmd/{release-cmd}/`
   - Copy GitHub workflows to `.github/workflows/`
   - Copy documentation files to root directory
   - Copy config files (Dockerfile, LICENSE, etc.)

3. **Document Processing**
   - Select language file based on `--use-lang` parameter
   - Copy `README.{lang}.md` to `README.md`

4. **Git Commit & Push**
   - Configure user information
   - Add all files
   - Commit changes
   - Push to remote repository

## 🐳 Environment Variables Setup

Recommended to use environment variables to avoid repeated input:

```bash
# Set environment variables (permanent)
echo 'export GRGH_NAME="YourName"' >> ~/.bashrc
echo 'export GRGH_EMAIL="your@email.com"' >> ~/.bashrc
echo 'export GRGH_REPO="YourGitHub/your-repo"' >> ~/.bashrc
source ~/.bashrc

# Or temporary setup
export GRGH_NAME="YourName"
export GRGH_EMAIL="your@email.com"
export GRGH_REPO="YourGitHub/your-repo"
```

**Priority**: Command line arguments > Environment variables > Default values

## 🔧 Advanced Usage

### Exclude Specific Directories
```bash
grgh --cmd sync --exclude-dirs ".git,.github,node_modules,dist,logs,temp"
```

### Add Extra Files
```bash
grgh --cmd sync --extra-files "./config,.env,./scripts"
```

### Custom Working Directory
```bash
grgh --cmd migrate --ws "/tmp/my-release" --gh-repo "YourGitHub/repo"
```

### Keep Temporary Directory (for debugging)
```bash
grgh --cmd push --clean=false
```

## 📊 Command Quick Reference

| Command | Description | Requires ws | Requires Network |
|---------|-------------|-------------|------------------|
| sync | One-click sync (recommended) | Auto | ✅ |
| migrate | Migrate files only | Auto | ❌ |
| clone | Clone repository only | Auto | ✅ |
| push | Push changes only | ✅ | ✅ |
| docs | Process documents only | ✅ | ❌ |
| clean | Clean directory only | Auto | ❌ |

## ⚠️ Important Notes

1. **Configure before first use** - set environment variables or use command line arguments
2. **Force push** (`--force`) overwrites remote history, use with caution
3. **Working directory** defaults to `./release_tmp/{repository-name}/`
4. **SSH keys** must be pre-configured for GitHub SSH access
5. **Repository formats** supported:
   - Simple format: `username/repo` (auto-converts to SSH)
   - SSH format: `git@github.com:username/repo.git`
   - HTTPS format: `https://github.com/username/repo.git`

## 🛡️ Security Features

- Automatically excludes `.git` directory to prevent conflicts
- Checks for file changes before push to avoid empty commits
- Force push confirmation mechanism
- Automatic temporary directory cleanup (optional)

## 📝 FAQ

**Q: Permission denied error?**  
A: Check GitHub SSH key configuration: `ssh -T git@github.com`

**Q: Push failed, what to do?**  
A: First run `grgh --cmd clean`, then retry `sync`

**Q: How to see detailed logs?**  
A: The tool outputs all command execution results in real-time

**Q: Windows support?**  
A: Yes, but Git Bash or WSL environment is recommended
