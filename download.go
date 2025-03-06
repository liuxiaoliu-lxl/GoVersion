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
	//// 获取原有版本数据
	//versions, err := fetchGoVersions()
	//if err != nil {
	//	dialog.ShowError(err, mainWindow)
	//	return nil
	//}

	var err error
	if versions == nil || len(versions) == 0 {
		// 获取网页上抓取的版本数据
		versions, err = fetchGoVersionsFromWeb()
		if err != nil {
			fmt.Println("err : ", err.Error())
		}
	}

	osGroups := make(map[string][]GoFile)
	for _, v := range versions {
		for _, file := range v.Files {
			if file.OS != "" && strings.ToLower(file.Kind) == "archive" {
				osGroups[file.OS] = append(osGroups[file.OS], file)
			}
		}
	}

	var osList []string
	for os := range osGroups {
		osList = append(osList, os)
	}
	sort.Strings(osList)

	for i := 0; i < len(osList); i++ {
		if osList[i] == "macOS" {
			osList[i], osList[0] = osList[0], osList[i]
		}
		if osList[i] == "Linux" {
			osList[i], osList[1] = osList[1], osList[i]
		}
	}

	osSelect := widget.NewSelect(osList, nil)
	var selectedFile *GoFile
	fileList := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {},
	)

	osSelect.OnChanged = func(selectedOS string) {
		files := osGroups[selectedOS]
		fileList.Length = func() int { return len(files) }
		fileList.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(fmt.Sprintf("%s (%s)", files[i].Filename, files[i].Arch))
		}
		fileList.OnSelected = func(i widget.ListItemID) {
			selectedFile = &files[i]
		}
		fileList.Refresh()
	}

	returnButton := widget.NewButton("返回", func() {
		mainWindow.SetContent(buildMainPage(mainWindow))
	})

	downloadButton := widget.NewButton("下载", func() {
		if selectedFile != nil {
			go func() {
				fName := strings.Replace(selectedFile.Filename, ".tar.gz", "", 1)
				if err := downloadAndExtract(selectedFile.Filename, fName); err != nil {
					dialog.ShowError(err, mainWindow)
				} else {
					dialog.ShowInformation("下载成功", "文件已安装", mainWindow)
				}
			}()
		} else {
			dialog.ShowInformation("下载失败", "未选择任何文件", mainWindow)
		}
	})

	buttonContainer := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(ButtonWidth, ButtonHeight), returnButton),
		container.NewGridWrap(fyne.NewSize(ButtonWidth, ButtonHeight), downloadButton),
	)

	return container.NewBorder(osSelect, buttonContainer, nil, nil, fileList)
}
