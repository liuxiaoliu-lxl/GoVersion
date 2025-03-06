package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"net/url"
)

func main() {
	// 初始化操作系统环境
	InitOS(workDir)

	// 创建 fyne 应用
	myApp := app.New()
	mainWindow := myApp.NewWindow("GoVersion")
	mainWindow.Resize(fyne.NewSize(WindowWidth, WindowHeight)) // 调整窗口大小

	// 创建主页面内容
	mainList := buildMainPage(mainWindow)

	mainWindow.SetContent(container.NewMax(mainList))
	mainWindow.ShowAndRun()
}

func buildMainPage(mainWindow fyne.Window) fyne.CanvasObject {
	mainList := widget.NewList(
		func() int {
			return 4 // 这里改为 4，新增“新特性”项
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			texts := []string{"切换", "安装", "卸载", "新特性"}
			o.(*widget.Label).SetText(texts[i])
		},
	)

	mainList.OnSelected = func(id widget.ListItemID) {
		switch id {
		case 0:
			mainWindow.SetContent(buildSwitchPage(mainWindow))
		case 1:
			mainWindow.SetContent(buildDownloadPage(mainWindow))
		case 2:
			mainWindow.SetContent(buildDeletePage(mainWindow))
		case 3:
			openGoReleasePage()
		}
	}

	return container.NewMax(mainList)
}

// 打开 Go Release 页面
func openGoReleasePage() {
	link, _ := url.Parse("https://go.dev/doc/devel/release")
	_ = fyne.CurrentApp().OpenURL(link)
}
