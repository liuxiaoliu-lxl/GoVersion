package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var versions []GoVersion

type GoFile struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	Kind     string `json:"kind"`
}

type GoVersion struct {
	Version string   `json:"version"`
	Stable  bool     `json:"stable"`
	Files   []GoFile `json:"files"`
}

// 检测当前 OS
func detectCurrentOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}

// 检测当前架构
func detectCurrentArch() string {
	return runtime.GOARCH
}

func downloadAndExtract(filename, version string) error {
	downloadURL := fmt.Sprintf("https://dl.google.com/go/%s", filename)
	installDir := filepath.Join(HomeDir+workDir, "install")
	targetDir := filepath.Join(HomeDir+workDir, "golang")

	// 解析目标文件夹名称
	targetDir = filepath.Join(targetDir, version)

	// 创建安装目录
	if err := os.MkdirAll(installDir, os.ModePerm); err != nil {
		return err
	}

	// 下载文件
	targetFile := filepath.Join(installDir, filename)
	out, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// 解压缩
	file, err := os.Open(targetFile)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}
	tarReader := tar.NewReader(gzipReader)
	for {
		head, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, head.Name)
		if head.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
				return err
			}
		} else {
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}

			// 为文件添加执行权限
			if err := os.Chmod(targetPath, 0755); err != nil {
				return err
			}
		}
	}
	err = os.Remove(targetFile)
	if err != nil {
		fmt.Println("删除文件失败 :", err.Error())
	}
	return nil
}
func buildDownloadPage(mainWindow fyne.Window) fyne.CanvasObject {
	var err error
	if versions == nil || len(versions) == 0 {
		versions, err = fetchGoVersionsFromWeb()
		if err != nil {
			fmt.Println("err : ", err.Error())
		}
	}

	// 分组数据
	osGroups := make(map[string]map[string][]GoFile)
	for _, v := range versions {
		for _, file := range v.Files {
			if file.OS == "" || strings.ToLower(file.Kind) != "archive" {
				continue
			}
			if osGroups[file.OS] == nil {
				osGroups[file.OS] = make(map[string][]GoFile)
			}
			osGroups[file.OS][file.Arch] = append(osGroups[file.OS][file.Arch], file)
		}
	}

	// OS 列表
	var osList []string
	for os := range osGroups {
		osList = append(osList, os)
	}
	sort.Strings(osList)

	// 当前系统
	currentOS := detectCurrentOS()
	currentArch := detectCurrentArch()

	// OS 选择框
	osSelect := widget.NewSelect(osList, nil)
	osSelect.SetSelected(currentOS)

	// Arch 选择框
	archSelect := widget.NewSelect(nil, nil)

	var selectedFile *GoFile
	var files []GoFile

	// **进度条 (默认隐藏)**
	loadingBar := widget.NewProgressBarInfinite()
	loadingBar.Hide()

	// **文件列表**
	fileList := widget.NewList(
		func() int { return len(files) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(fmt.Sprintf("%s (%s)", files[i].Filename, files[i].Arch))
		},
	)
	fileList.OnSelected = func(i widget.ListItemID) {
		selectedFile = &files[i]
	}

	// 更新 Arch 选择框
	updateArchSelect := func(selectedOS string) {
		if archs, exists := osGroups[selectedOS]; exists {
			var archList []string
			for arch := range archs {
				archList = append(archList, arch)
			}
			sort.Strings(archList)
			archSelect.Options = archList
			if len(archList) > 0 {
				archSelect.SetSelected(archList[0])
			}
		}
	}

	// 更新文件列表
	updateFileList := func(selectedOS, selectedArch string) {
		if archs, exists := osGroups[selectedOS]; exists {
			if filesForArch, exists := archs[selectedArch]; exists {
				files = filesForArch
			} else {
				files = nil
			}
		} else {
			files = nil
		}
		fileList.Refresh()
	}

	// **监听 OS 变化**
	osSelect.OnChanged = func(selectedOS string) {
		updateArchSelect(selectedOS) // 更新架构选项
		if archSelect.Selected != "" {
			updateFileList(selectedOS, archSelect.Selected)
		}
	}

	// **监听 Arch 变化**
	archSelect.OnChanged = func(selectedArch string) {
		updateFileList(osSelect.Selected, selectedArch)
	}

	// **初始化 UI**
	updateArchSelect(currentOS)
	if archSelect.Selected == "" {
		archSelect.SetSelected(currentArch) // 选中当前架构
	}
	updateFileList(currentOS, archSelect.Selected)

	// **两级筛选项并排**
	selectContainer := container.NewGridWithColumns(2, osSelect, archSelect)

	// **提前声明 downloadButton**
	var downloadButton *widget.Button

	// **返回按钮**
	returnButton := widget.NewButton("返回", func() {
		mainWindow.SetContent(buildMainPage(mainWindow))
	})

	// **安装按钮**
	downloadButton = widget.NewButton("安装", func() {
		if selectedFile != nil {
			// **开始安装，显示加载状态**
			downloadButton.SetText("安装中...")
			downloadButton.Disable()
			loadingBar.Show()

			go func() {
				fName := strings.Replace(selectedFile.Filename, ".tar.gz", "", 1)
				err := downloadAndExtract(selectedFile.Filename, fName)

				// **安装完成，恢复 UI**
				loadingBar.Hide()
				downloadButton.SetText("安装")
				downloadButton.Enable()

				// 弹窗提示
				if err != nil {
					dialog.ShowError(err, mainWindow)
				} else {
					dialog.ShowInformation("安装成功", "文件已安装", mainWindow)
				}
			}()
		} else {
			dialog.ShowInformation("安装失败", "未选择任何文件", mainWindow)
		}
	})

	// **按钮容器**
	buttonContainer := container.NewGridWithColumns(2, returnButton, downloadButton)

	// **最终页面**
	return container.NewBorder(
		selectContainer,
		container.NewVBox(buttonContainer, loadingBar), // **把 Loading 进度条放到按钮下方**
		nil, nil, fileList,
	)
}
