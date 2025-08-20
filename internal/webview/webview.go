package webview

import (
	webview "github.com/webview/webview_go"
)

func Run() {
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("Web button command run")
	w.SetSize(800, 600, webview.HintFixed)
	w.Navigate("http://localhost:8080")
	w.Run()
}
