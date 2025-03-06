package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"os"
	"path/filepath"
)

// 构建删除页面
func buildDeletePage(mainWindow fyne.Window) fyne.CanvasObject {
	// 获取 Go 版本目录列表
	res, err := ListFolders(workDir + GoRootDirName)
	if err != nil {
		fmt.Println("err : ", err.Error())
		return nil
	}

	var selectedValue string

	list := widget.NewList(
		func() int {
			return len(res)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(res[i])
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(res) {
			selectedValue = res[id]
		}
	}

	deleteButton := widget.NewButton("删除", func() {
		if selectedValue != "" {
			targetPath := filepath.Join(HomeDir, workDir, GoRootDirName, selectedValue)
			err := os.RemoveAll(targetPath)
			if err != nil {
				dialog.ShowError(err, mainWindow)
			} else {
				dialog.ShowInformation("删除成功", "已删除: "+selectedValue, mainWindow)
				mainWindow.SetContent(buildDeletePage(mainWindow)) // 刷新页面
			}
		} else {
			dialog.ShowInformation("删除失败", "未选择任何项", mainWindow)
		}
	})

	returnButton := widget.NewButton("返回", func() {
		mainWindow.SetContent(buildMainPage(mainWindow))
	})

	buttonContainer := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(ButtonWidth, ButtonHeight), returnButton),
		container.NewGridWrap(fyne.NewSize(ButtonWidth, ButtonHeight), deleteButton),
	)

	return container.NewBorder(nil, buttonContainer, nil, nil, list)
}
