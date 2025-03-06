package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"path/filepath"
)

func buildSwitchPage(mainWindow fyne.Window) fyne.CanvasObject {
	// 获取文件夹列表
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

	switchButton := widget.NewButton("切换", func() {
		if selectedValue != "" {
			updateEnvVariable(ExportGoRoot, filepath.Join(HomeDir, workDir, GoRootDirName+"/"+selectedValue+"/go"))
			dialog.ShowInformation("切换成功", "已选择: "+selectedValue, mainWindow)
		} else {
			dialog.ShowInformation("切换失败", "未选择任何项", mainWindow)
		}
	})

	returnButton := widget.NewButton("返回", func() {
		mainWindow.SetContent(buildMainPage(mainWindow))
	})

	buttonContainer := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(ButtonWidth, ButtonHeight), returnButton),
		container.NewGridWrap(fyne.NewSize(ButtonWidth, ButtonHeight), switchButton),
	)

	return container.NewBorder(nil, buttonContainer, nil, nil, list)
}
