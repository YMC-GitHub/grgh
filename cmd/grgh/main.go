package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 版本信息
const Version = "v1.0.0"

var usage = `Git发布工具 - 迁移文件并推送到GitHub

用法: grgh --cmd <命令> [选项]

命令：
  migrate           迁移发布文件到临时目录
  clone             克隆远程仓库到临时目录
  push              Git推送（需先执行migrate或clone）
  clean             清理临时发布目录
  sync              同步流程：智能处理（clone或init）+ 迁移 + 推送
  docs              处理文档文件（重命名 README）

选项：
  --cmd             命令名称 (必需)
  --gh-name         Git用户名称 (默认: YeMincheng, 环境变量: GRGH_NAME)
  --gh-email        Git用户邮箱 (默认: ymc.github@gmail.com, 环境变量: GRGH_EMAIL)
  --gh-repo         GitHub仓库 (默认: ymc-github/pdfi, 环境变量: GRGH_REPO)
  --gh-msg          提交信息 (默认: init: add all files)
  --gh-branch       Git分支名称 (默认: main)
  --clean           清理临时目录 (默认: true)
  --force           强制推送 (默认: false)
  --ws              工作目录路径 (可选，默认从仓库名生成)
  --release-cmd     发布命令名称 (可选，默认从仓库名提取)
  --extra-files     额外文件或目录（逗号分隔，如: "./config,.env"）
  --exclude-dirs    排除的目录（逗号分隔，默认: ".git,.github,node_modules"）
  --use-lang        指定默认语言 (zh/en/jp/kr 等，默认: zh)
  -h, --help        显示帮助信息
  -v, --version     显示版本信息

环境变量：
  GRGH_NAME         Git用户名称
  GRGH_EMAIL        Git用户邮箱
  GRGH_REPO         GitHub仓库

示例：
  # 使用环境变量
  export GRGH_NAME="YourName"
  export GRGH_EMAIL="your@email.com"
  export GRGH_REPO="YMC-GitHub/iufx"
  grgh --cmd sync --use-lang zh
  
  # 命令行参数优先级更高
  grgh --cmd sync --gh-repo "YMC-GitHub/iufx" --gh-name "MyName"
`

type Config struct {
	// Git 配置
	Name   string
	Email  string
	Repo   string
	Msg    string
	Branch string
	Clean  bool
	Force  bool
	Ws     string

	// 发布配置
	ReleaseCmd  string
	ExtraFiles  string
	ExcludeDirs string
	UseLang     string
}

func main() {
	config, cmd := parseFlags()

	if config == nil {
		return
	}

	switch cmd {
	case "migrate":
		if err := runMigrate(*config); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 迁移失败: %v\n", err)
			os.Exit(1)
		}
	case "clone":
		if err := runClone(*config); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 克隆失败: %v\n", err)
			os.Exit(1)
		}
	case "push":
		if err := runPush(*config); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 推送失败: %v\n", err)
			os.Exit(1)
		}
	case "clean":
		if err := runClean(*config); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 清理失败: %v\n", err)
			os.Exit(1)
		}
	case "sync":
		if err := runSync(*config); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 同步失败: %v\n", err)
			os.Exit(1)
		}
	case "docs":
		if err := runDocs(*config); err != nil {
			fmt.Fprintf(os.Stderr, "❌ 文档处理失败: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("❌ 未知命令: %s\n", cmd)
		os.Exit(1)
	}
}

func parseFlags() (*Config, string) {
	config := Config{
		ExcludeDirs: ".git,.github,node_modules,dist,build,target",
		UseLang:     "zh",
	}

	// 从环境变量读取默认值
	envName := os.Getenv("GRGH_NAME")
	envEmail := os.Getenv("GRGH_EMAIL")
	envRepo := os.Getenv("GRGH_REPO")

	// 设置默认值（优先环境变量）
	defaultName := "YeMincheng"
	defaultEmail := "ymc.github@gmail.com"
	defaultRepo := "ymc-github/pdfi"

	if envName != "" {
		defaultName = envName
	}
	if envEmail != "" {
		defaultEmail = envEmail
	}
	if envRepo != "" {
		defaultRepo = envRepo
	}

	var (
		cmd     string
		help    bool
		version bool
	)

	flag.StringVar(&cmd, "cmd", "", "Command: migrate, clone, push, clean, sync, docs")
	flag.StringVar(&config.Name, "gh-name", defaultName, "Git user name (env: GRGH_NAME)")
	flag.StringVar(&config.Email, "gh-email", defaultEmail, "Git user email (env: GRGH_EMAIL)")
	flag.StringVar(&config.Repo, "gh-repo", defaultRepo, "GitHub repository (env: GRGH_REPO)")
	flag.StringVar(&config.Msg, "gh-msg", "init: add all files", "Commit message")
	flag.StringVar(&config.Branch, "gh-branch", "main", "Git branch name")
	flag.BoolVar(&config.Clean, "clean", true, "Clean up temporary directory after push")
	flag.BoolVar(&config.Force, "force", false, "Force push to remote repository")
	flag.StringVar(&config.Ws, "ws", "", "Working directory path")
	flag.StringVar(&config.ReleaseCmd, "release-cmd", "", "Release command name")
	flag.StringVar(&config.ExtraFiles, "extra-files", "", "Extra files or directories")
	flag.StringVar(&config.ExcludeDirs, "exclude-dirs", ".git,.github,node_modules,dist,build,target", "Directories to exclude")
	flag.StringVar(&config.UseLang, "use-lang", "zh", "Default language for README")
	flag.BoolVar(&help, "h", false, "Help")
	flag.BoolVar(&help, "help", false, "Help")
	flag.BoolVar(&version, "v", false, "Version")
	flag.BoolVar(&version, "version", false, "Version")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	if help {
		flag.Usage()
		return nil, ""
	}
	if version {
		fmt.Println(Version)
		return nil, ""
	}
	if cmd == "" {
		fmt.Println("❌ 请指定 --cmd 参数\n")
		flag.Usage()
		return nil, ""
	}

	validCommands := map[string]bool{
		"migrate": true, "clone": true, "push": true,
		"clean": true, "sync": true, "docs": true,
	}
	if !validCommands[cmd] {
		fmt.Printf("❌ 无效命令: %s\n", cmd)
		return nil, ""
	}

	repoBaseName := extractRepoBaseName(config.Repo)

	if config.ReleaseCmd == "" {
		config.ReleaseCmd = repoBaseName
	}

	if (cmd == "push" || cmd == "clean" || cmd == "sync" || cmd == "clone" || cmd == "docs") && config.Ws == "" {
		config.Ws = fmt.Sprintf("./release_tmp/%s", repoBaseName)
	}

	// 构建 SSH URL
	if cmd != "clean" && cmd != "docs" && config.Repo != "" {
		// 如果已经是 SSH 或 HTTPS 格式，保持不变
		if !strings.Contains(config.Repo, "git@") && !strings.Contains(config.Repo, "https://") {
			config.Repo = fmt.Sprintf("git@github.com:%s.git", config.Repo)
		}
	}

	// 显示使用的配置（调试用）
	if cmd != "clean" && cmd != "docs" {
		fmt.Printf("📋 配置: Name=%s, Email=%s, Repo=%s\n", config.Name, config.Email, config.Repo)
	}

	return &config, cmd
}

func extractRepoBaseName(repo string) string {
	repo = strings.TrimSuffix(repo, ".git")
	if strings.Contains(repo, ":") && strings.Contains(repo, "@") {
		parts := strings.Split(repo, ":")
		if len(parts) > 1 {
			repo = parts[1]
		}
	}
	if strings.Contains(repo, "://") {
		parts := strings.Split(repo, "/")
		if len(parts) >= 2 {
			repo = strings.Join(parts[len(parts)-2:], "/")
		}
	}
	repo = strings.TrimSuffix(repo, ".git")
	parts := strings.Split(repo, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "repo"
}

func runDocs(config Config) error {
	fmt.Println("📚 开始处理文档文件...")

	wsAbs, err := filepath.Abs(config.Ws)
	if err != nil {
		return fmt.Errorf("获取路径失败: %w", err)
	}

	if _, err := os.Stat(wsAbs); os.IsNotExist(err) {
		return fmt.Errorf("目录不存在: %s", wsAbs)
	}

	fmt.Printf("📂 工作目录: %s\n", wsAbs)

	langFile := fmt.Sprintf("README.%s.md", config.UseLang)
	langPath := filepath.Join(wsAbs, langFile)

	if _, err := os.Stat(langPath); os.IsNotExist(err) {
		return fmt.Errorf("语言文件不存在: %s", langFile)
	}

	readmePath := filepath.Join(wsAbs, "README.md")
	if err := copyFile(langPath, readmePath); err != nil {
		return fmt.Errorf("复制失败: %w", err)
	}

	fmt.Printf("✅ 已将 %s 设为 README.md\n", langFile)
	return nil
}

func runSync(config Config) error {
	fmt.Println("🚀 开始同步流程...")

	wsPath := config.Ws
	if wsPath == "" {
		repoBaseName := extractRepoBaseName(config.Repo)
		wsPath = fmt.Sprintf("./release_tmp/%s", repoBaseName)
		config.Ws = wsPath
	}

	// 检查并准备仓库
	wsAbs, _ := filepath.Abs(wsPath)
	if _, err := os.Stat(wsAbs); os.IsNotExist(err) {
		fmt.Printf("📁 目录不存在: %s\n", wsAbs)
		if err := runClone(config); err != nil {
			fmt.Printf("⚠️ 克隆失败，创建新仓库: %v\n", err)
			if err := os.MkdirAll(wsAbs, 0755); err != nil {
				return err
			}
			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)
			os.Chdir(wsAbs)
			runCommand("git", "init")
			runCommand("git", "branch", "-m", config.Branch)
			runCommand("git", "remote", "add", "origin", config.Repo)
		}
	}

	if err := runMigrate(config); err != nil {
		return err
	}

	if err := runDocs(config); err != nil {
		fmt.Printf("⚠️ 文档处理失败: %v\n", err)
	}

	if err := runPush(config); err != nil {
		return err
	}

	fmt.Println("\n🎉 同步完成！")
	return nil
}

func runMigrate(config Config) error {
	fmt.Println("📦 开始迁移发布文件...")

	if config.Ws == "" {
		repoBaseName := extractRepoBaseName(config.Repo)
		config.Ws = fmt.Sprintf("./release_tmp/%s", repoBaseName)
	}

	if err := os.MkdirAll(config.Ws, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 拷贝目录
	dirs := []struct{ src, dst string }{
		{fmt.Sprintf("./cmd/%s", config.ReleaseCmd), filepath.Join(config.Ws, "cmd", config.ReleaseCmd)},
		{fmt.Sprintf("./ghwf/%s", config.ReleaseCmd), filepath.Join(config.Ws, ".github/workflows")},
	}

	for _, d := range dirs {
		if err := copyPath(d.src, d.dst, config.ExcludeDirs); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	// 拷贝文件
	files := []struct{ src, dst string }{
		{fmt.Sprintf("./Dockerfile.%s", config.ReleaseCmd), filepath.Join(config.Ws, fmt.Sprintf("Dockerfile.%s", config.ReleaseCmd))},
		{"./LICENSE", filepath.Join(config.Ws, "LICENSE")},
		{"./LICENSE-APACHE", filepath.Join(config.Ws, "LICENSE-APACHE")},
		{"./LICENSE-MIT", filepath.Join(config.Ws, "LICENSE-MIT")},
		{"./.Dockerignore", filepath.Join(config.Ws, ".Dockerignore")},
		{"./.editorconfig", filepath.Join(config.Ws, ".editorconfig")},
		{"./.gitignore", filepath.Join(config.Ws, ".gitignore")},
	}

	for _, f := range files {
		if err := copyFile(f.src, f.dst); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	// 拷贝文档
	docsSrc := fmt.Sprintf("./docs/%s", config.ReleaseCmd)
	if entries, err := os.ReadDir(docsSrc); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				src := filepath.Join(docsSrc, entry.Name())
				dst := filepath.Join(config.Ws, entry.Name())
				copyFile(src, dst)
			}
		}
	}

	fmt.Printf("✅ 迁移完成: %s\n", config.Ws)
	return nil
}

func runClone(config Config) error {
	fmt.Println("📥 开始克隆远程仓库...")

	wsPath := config.Ws
	if wsPath == "" {
		repoBaseName := extractRepoBaseName(config.Repo)
		wsPath = fmt.Sprintf("./release_tmp/%s", repoBaseName)
	}

	// 如果目录已存在且是 Git 仓库，直接更新
	if _, err := os.Stat(filepath.Join(wsPath, ".git")); err == nil {
		fmt.Printf("🔄 更新仓库: %s\n", wsPath)
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(wsPath)
		runCommand("git", "pull", "origin", config.Branch)
		return nil
	}

	// 删除已存在但非 Git 的目录
	if _, err := os.Stat(wsPath); err == nil {
		os.RemoveAll(wsPath)
	}

	// 克隆仓库
	if err := runCommand("git", "clone", config.Repo, wsPath); err != nil {
		return fmt.Errorf("克隆失败: %w", err)
	}

	fmt.Println("✅ 克隆完成")
	return nil
}

func runPush(config Config) error {
	fmt.Println("🚀 开始 Git 推送...")

	wsAbs, err := filepath.Abs(config.Ws)
	if err != nil {
		return err
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if _, err := os.Stat(wsAbs); os.IsNotExist(err) {
		return fmt.Errorf("目录不存在: %s", wsAbs)
	}

	fmt.Printf("📂 切换到: %s\n", wsAbs)
	if err := os.Chdir(wsAbs); err != nil {
		return err
	}

	// 确保是 Git 仓库
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		runCommand("git", "init")
	}

	// 配置用户
	runCommand("git", "config", "user.name", config.Name)
	runCommand("git", "config", "user.email", config.Email)

	// 设置分支
	currentBranch, _ := exec.Command("git", "branch", "--show-current").Output()
	if strings.TrimSpace(string(currentBranch)) != config.Branch {
		runCommand("git", "branch", "-m", config.Branch)
	}

	// 设置远程仓库
	if err := runCommand("git", "remote", "get-url", "origin"); err != nil {
		runCommand("git", "remote", "add", "origin", config.Repo)
	} else {
		runCommand("git", "remote", "set-url", "origin", config.Repo)
	}

	// 添加并提交
	fmt.Println("📦 添加文件...")
	runCommand("git", "add", ".")

	// 检查是否有变更
	status, _ := exec.Command("git", "status", "--porcelain").Output()
	if len(status) > 0 {
		fmt.Printf("💾 提交: %s\n", config.Msg)
		if err := runCommand("git", "commit", "-m", config.Msg); err != nil {
			return fmt.Errorf("提交失败: %w", err)
		}
	} else {
		fmt.Println("✅ 无变更需要提交")
	}

	// 推送
	fmt.Printf("🚀 推送到 %s...\n", config.Branch)
	pushCmd := []string{"push", "-u", "origin", config.Branch}
	if config.Force {
		pushCmd = []string{"push", "-f", "-u", "origin", config.Branch}
	}

	if err := runCommand("git", pushCmd...); err != nil {
		return fmt.Errorf("推送失败: %w", err)
	}

	fmt.Println("✅ 推送成功")

	if config.Clean {
		fmt.Printf("🧹 清理: %s\n", config.Ws)
		os.RemoveAll(config.Ws)
	}

	return nil
}

func runClean(config Config) error {
	fmt.Println("🧹 清理临时目录...")

	wsPath := config.Ws
	if wsPath == "" {
		repoBaseName := extractRepoBaseName(config.Repo)
		wsPath = fmt.Sprintf("./release_tmp/%s", repoBaseName)
	}

	wsAbs, _ := filepath.Abs(wsPath)
	if _, err := os.Stat(wsAbs); os.IsNotExist(err) {
		fmt.Printf("目录不存在: %s\n", wsAbs)
		return nil
	}

	fmt.Printf("删除: %s\n", wsAbs)
	if err := os.RemoveAll(wsAbs); err != nil {
		return err
	}

	fmt.Println("✅ 清理完成")
	return nil
}

func copyPath(src, dst, excludeDirs string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return copyFile(src, dst)
	}

	return copyDir(src, dst, excludeDirs)
}

func copyDir(src, dst, excludeDirs string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	excludeMap := make(map[string]bool)
	for _, d := range strings.Split(excludeDirs, ",") {
		excludeMap[strings.TrimSpace(d)] = true
	}

	for _, entry := range entries {
		if entry.IsDir() && excludeMap[entry.Name()] {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			copyDir(srcPath, dstPath, excludeDirs)
		} else {
			copyFile(srcPath, dstPath)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}