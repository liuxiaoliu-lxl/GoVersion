package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ExportGoPath = "export GOPATH="
	ExportGoRoot = "export GOROOT="
	ExportPath   = "export PATH="
	PathAppend   = "$PATH:$GOPATH:$GOROOT/bin"

	GoPathDirName  = "/gopath"
	GoRootDirName  = "/golang"
	InstallDirName = "/install"

	DefaultWorkDir = "/go_dev"
)

var HomeDir = ""
var EnvPath = "/.go_dev"

func InitOS(workDir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("无法获取用户目录:", err)
		return
	}
	HomeDir = homeDir

	if workDir == "" {
		workDir = DefaultWorkDir
	}

	// 用户可传入此参数
	goPath := filepath.Join(homeDir, workDir, GoPathDirName)
	goRoot := filepath.Join(homeDir, workDir, GoRootDirName)
	install := filepath.Join(homeDir, workDir, InstallDirName)

	// 确保目录存在
	ensureDirExists(goPath)
	ensureDirExists(goRoot)
	ensureDirExists(install)

	// Step 1: 检查 .go_dev 文件是否存在，如果不存在则创建
	goDevPath := HomeDir + EnvPath
	if _, err := os.Stat(goDevPath); os.IsNotExist(err) {
		file, err := os.Create(goDevPath)
		if err != nil {
			fmt.Println("创建 .go_dev 失败:", err)
			return
		}
		file.Close()
		fmt.Println("已创建 .go_dev 文件")
	}

	// 追加三行环境变量
	addEnvVariable(ExportGoPath, goPath)
	addEnvVariable(ExportGoRoot, goRoot)
	addEnvVariable(ExportPath, PathAppend)

	// Step 2: 检查 .bash_profile 是否包含指定代码，如果没有则追加
	bashProfilePath := homeDir + "/.bash_profile"
	codeSnippet := "[[ -s ~/.go_dev ]] && source ~/.go_dev"

	if fileExists(bashProfilePath) {
		if !containsLine(bashProfilePath, codeSnippet) {
			appendToFile(bashProfilePath, codeSnippet)
			fmt.Println("已追加代码到 .bash_profile")
		} else {
			fmt.Println(".bash_profile 已包含该代码")
		}
	} else {
		// 如果 .bash_profile 不存在，创建并写入
		createAndWriteFile(bashProfilePath, codeSnippet)
		fmt.Println("已创建 .bash_profile 并写入代码")
	}
}

func addEnvVariable(prefix, newLine string) {
	filename := HomeDir + EnvPath
	lines := readFileLines(filename)

	// 检查是否已存在相同的前缀
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return // 如果存在则直接返回，不进行修改
		}
	}

	// 如果不存在，则追加新行
	lines = append(lines, prefix+newLine)
	writeFileLines(filename, lines)
}

func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("创建目录失败:", err)
		} else {
			fmt.Println("已创建目录:", dir)
		}
	}
}

func updateEnvVariable(prefix, newLine string) {
	filename := HomeDir + EnvPath
	lines := readFileLines(filename)
	updated := false
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = prefix + newLine
			updated = true
			break
		}
	}
	if !updated {
		lines = append(lines, prefix+newLine)
	}
	writeFileLines(filename, lines)
}

func writeFileLines(filename string, lines []string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("无法打开文件进行写入:", err)
		return
	}
	defer file.Close()

	for _, line := range lines {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			fmt.Println("写入文件失败:", err)
		}
	}
}

func readFileLines(filename string) []string {
	var lines []string
	file, err := os.Open(filename)
	if err != nil {
		return lines
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	return lines
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// containsLine 检查文件是否包含特定的行
func containsLine(filename, target string) bool {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("打开文件失败:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == target {
			return true
		}
	}
	return false
}

// appendToFile 追加内容到文件
func appendToFile(filename, content string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("打开文件追加失败:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString("\n" + content + "\n")
	if err != nil {
		fmt.Println("写入文件失败:", err)
	}
}

// createAndWriteFile 创建文件并写入内容
func createAndWriteFile(filename, content string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("创建文件失败:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(content + "\n")
	if err != nil {
		fmt.Println("写入文件失败:", err)
	}
}

func ListFolders(root string) ([]string, error) {
	if HomeDir == "" {
		return nil, errors.New("homeDir error")
	}
	root = HomeDir + root
	var folders []string
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
		}
	}
	return folders, err
}
