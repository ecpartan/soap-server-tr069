package web

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anoshenko/rui"
	jrcp2server "github.com/ecpartan/soap-server-tr069/pkg/jrpc2"
	"github.com/ecpartan/soap-server-tr069/tasks/tasker"
	"github.com/fanliao/go-promise"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2/methods"
	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2/mwdto"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
)

/*
	GridLayout {
				style = optionsTable,
				content = [
					TextView { row = 0, column = 0, id = nodename, text = "Cell gap" },
					TextView { row = 0, column = 1, id = nodetype, text = "Table border" },
					TextView { row = 0, column = 2, id = nodeattr, text = "Cell border" },
					TextView { row = 0, column = 3, id = nodevalue, text = "Cell padding" },
				]
			},
*/

const gridLayoutDemoText = `
GridLayout {
	style = demoPage,
	content = [
		EditView {
			row = 0, id = editWindow, type = multiline, read-only = true, wrap = true,
		},
		Resizable {
			row = 1, side = top, background-color = lightgrey, height = 200px,
			content = EditView {
				id = deviceLog, type = multiline, read-only = true, wrap = true,
			}
		},
		Button {
			row = 2, id = findDevice, content = "Find device"
		},
		Resizable {
			row = 3, side = top, background-color = lightgrey, height = 200px,
			content = EditView {
				id = scriptLog, type = multiline, read-only = true, wrap = true,
			}
		},
		Button {
			row = 4, id = executeScript, content = "Execute script"
		},

		GridLayout {
			width = 100%, height = 100%, cell-height = "auto, 1fr", cell-width = "1fr, auto",
			content = [
		FilePicker {
			id = filePicker, accept = "txt, html"
		},
		Button {
			id = fileDownload, row = 0, column = 1, content = "Download file", disabled = true,
		}
		EditView {
			id = selectedFileData, row = 1, column = 0:1, type = multiline, read-only = true, wrap = true,
		}
	]
}
	]
}
`

const popupDemoText = `
GridLayout {
	width = 100%, height = 100%, cell-height = "auto, 1fr",
	content = GridLayout {
		width = 100%, cell-width = "auto, 1fr",
		cell-vertical-align = center, gap = 8px,
		content = [
			Button {
				id = popupShowMessage, margin = 4px, content = "Show users",
			},
			Button {
				id = popupShowQuestion, row = 1, margin = 4px, content = "Show groups",
			},
			Button {
				id = popupUploadFile, row = 2, margin = 4px, content = "Upload file",
			},
			Button {
				id = popupExecScript, row = 3, margin = 4px, content = "Execute script",
			},
			Button {
				id = popupShowDevice, row = 4, margin = 4px, content = "Find device",
			},
		]
	}
}
`

var arr = []string{}

func recurse(lst map[string]any, curr string) {
	for k, v := range lst {

		if mp, ok := v.(map[string]any); ok {

			if len(k) == 0 {
				continue
			}

			if val, ok := mp["Value"]; ok {
				addObj := curr + k + ":" + val.(string)
				arr = append(arr, addObj)

			} else {
				curr += k + "."

				recurse(mp, curr)
				curr = curr[:len(curr)-len(k)-1]
			}

		}
	}
}

var downloadedFile []byte = nil
var downloadedFilename = ""

func (content *webSession) CreateRootView(session rui.Session) rui.View {
	view := rui.CreateViewFromText(session, popupDemoText)
	if view == nil {
		return nil
	}

	rui.Set(view, "popupUploadFile", rui.ClickEvent, func() {
		showPopupUploadFile(view)
	})

	rui.Set(view, "popupExecScript", rui.ClickEvent, func() {
		showPopupExecScript(view)
	})

	rui.Set(view, "popupShowDevice", rui.ClickEvent, func() {
		showPopupDevice(view)
	})

	return view
}

type webPage struct {
	title   string
	creator func(session rui.Session) rui.View
	view    rui.View
}

type webSession struct {
	rootView rui.View
	pages    []webPage
}

func createHelloWorldSession(session rui.Session) rui.SessionContent {
	return new(webSession)
}

func Register(addr string, port int) {
	go func() {
		logger.LogDebug("Start web server ", fmt.Sprintf("%s:%d", addr, port))
		rui.StartApp(fmt.Sprintf("%s:%d", addr, port), createHelloWorldSession, rui.AppParams{
			Title: "SOAP SERVER TR069",
		})
	}()
}

const popupshowDeviceText = `
Popup {
	min-width = 1200px, min-height = 800px, resize = both,
	close-button = true, "outside-close" = false, 
	title = "Enter text", 
	content = GridLayout {
		width = 100%, height = 100%, padding = 12px, gap = 4px, 
    	cell-height = [auto, auto, auto, 1fr],
		content = [
			EditView {
				row = 0, id = editWindow, type = multiline, read-only = true, wrap = true,
			},
			Resizable {
				row = 1, side = top, background-color = lightgrey, height = 400px
				content = EditView {
					id = deviceLog, type = multiline, read-only = true, wrap = true,
				}
			},
			Button {
				row = 2, id = findDevice, content = "Find device"
			},
		],
	},
	buttons = [
		{ title = Cancel, type = cancel },
	],
	show-duration = 0.5, show-opacity = 0,
	show-transform = _{ scale-x = 0.001, scale-y = 0.001 },
}
`

type popupShowDevice struct {
	popup    rui.Popup
	rootView rui.View
}

func showPopupDevice(rootView rui.View) {

	data := new(popupShowDevice)
	data.rootView = rootView

	data.popup = rui.CreatePopupFromText(rootView.Session(), popupshowDeviceText, data)

	if data.popup != nil {
		data.popup.Show()
	}

	popupView := data.popup.View()

	rui.Set(popupView, "findDevice", rui.ClickEvent, func(rui.View) {
		str := rui.GetText(popupView, "editWindow")

		mp := repository.GetCache().Get(str)
		arr = []string{}
		var curr string
		recurse(mp, curr)

		var result string
		for _, line := range arr {
			result += line + "\n"
		}
		rui.Set(popupView, "deviceLog", rui.Text, result)
	})
}

const popupexecScriptText = `
Popup {
	min-width = 1200px, min-height = 800px, resize = both,
	close-button = true, "outside-close" = false, 
	title = "Exec script", 
	content = GridLayout {
		width = 100%, height = 100%, padding = 12px, gap = 4px, 
    	cell-height = [auto, auto, auto, 1fr],
		content = [
			Resizable {
				row = 3, side = top, background-color = lightgrey, height = 200px,
				content = EditView {
					id = scriptLog, type = multiline, read-only = true, wrap = true,
				}
			},
			Button {
				row = 4, id = executeScript, content = "Execute script"
			},
		],
	},
	buttons = [
		{ title = OK, click = ClickOK },
		{ title = Cancel, type = cancel },
	],
	show-duration = 0.5, show-opacity = 0,
	show-transform = _{ scale-x = 0.001, scale-y = 0.001 },
}
`

type popupExecScript struct {
	popup    rui.Popup
	rootView rui.View
}

func showPopupExecScript(rootView rui.View) {

	data := new(popupExecScript)
	data.rootView = rootView

	data.popup = rui.CreatePopupFromText(rootView.Session(), popupexecScriptText, data)

	if data.popup != nil {
		data.popup.Show()
	}

	popupView := data.popup.View()

	rui.Set(popupView, "executeScript", rui.ClickEvent, func(rui.View) {
		str := rui.GetText(popupView, "scriptLog")

		var getScript map[string]any
		err := json.Unmarshal([]byte(str), &getScript)
		if err != nil {
			return
		}

		logger.LogDebug("script", getScript)
		if script, ok := getScript["Script"].(map[string]any); ok {

			task := func() (any, error) {
				var ret []byte
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				err := jrcp2server.Instance.Server.Client.CallResult(ctx, methods.MethodAddScript, mwdto.Mwdto{script, tasker.GetTasker().ExecTasks}, &ret)
				return ret, err

			}

			f := promise.Start(task).OnSuccess(func(result any) {
				logger.LogDebug("Success", result)
			}).OnFailure(func(v any) {
				logger.LogDebug("Failure", v)
			})

			ret, err := f.Get()

			logger.LogDebug("Get", ret)

			if err != nil {
				rui.Set(popupView, "scriptLog", rui.Text, err.Error())
			} else {
				rui.Set(popupView, "scriptLog", rui.Text, string(ret.([]byte)))
			}
		}

	})
}

const filePickerDemoText = `
Popup {
	min-width = 1200px, min-height = 800px, resize = both,
	close-button = true, "outside-close" = false, 
	title = "Exec script", 
	content = GridLayout {
	width = 100%, height = 100%, cell-height = "auto, 1fr", cell-width = "1fr, auto",
	content = [
		FilePicker {
			id = filePicker, accept = "bin"
		},
		Button {
			id = fileDownload, row = 0, column = 1, content = "Download file", disabled = true,
		}
		EditView {
			id = selectedFileData, row = 1, column = 0:1, type = multiline, read-only = true, wrap = true,
		}
	]
},
	buttons = [
		{ title = OK, click = ClickOK },
		{ title = Cancel, type = cancel },
	],
	show-duration = 0.5, show-opacity = 0,
	show-transform = _{ scale-x = 0.001, scale-y = 0.001 },
}
`

type popupUploadFile struct {
	popup    rui.Popup
	rootView rui.View
}

func getListdir() []string {
	files, err := os.ReadDir("./uploads")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return nil
	}

	var fileList []string
	for _, file := range files {
		fileList = append(fileList, file.Name())
	}

	return fileList
}

func showPopupUploadFile(rootView rui.View) {
	data := new(popupUploadFile)
	data.rootView = rootView

	data.popup = rui.CreatePopupFromText(rootView.Session(), filePickerDemoText, data)

	if data.popup != nil {
		data.popup.Show()
	}

	popupView := data.popup.View()

	lst := getListdir()
	var result string
	for _, line := range lst {
		result += line + "\n"
	}
	rui.Set(popupView, "selectedFileData", rui.Text, result)

	rui.Set(popupView, "filePicker", rui.FileSelectedEvent, func(picker rui.FilePicker, files []rui.FileInfo) {
		if len(files) > 0 {
			picker.LoadFile(files[0], func(_ rui.FileInfo, data []byte) {
				if data != nil {
					downloadedFile = data
					downloadedFilename = files[0].Name
					rui.Set(popupView, "fileDownload", rui.Disabled, false)
				} else {
					rui.Set(popupView, "selectedFileData", rui.Text, rui.LastError())
				}
			})
		}
	})

	rui.Set(popupView, "fileDownload", rui.ClickEvent, func() {

		dst, err := os.Create(filepath.Join("uploads", downloadedFilename))
		if err != nil {
			fmt.Println("Error creating destination file:", err)
			return
		}
		defer dst.Close()

		if _, err := dst.Write(downloadedFile); err != nil {
			fmt.Println("Error saving destination file:", err)
		}
		//popupView.Session().DownloadFileData(downloadedFilename, downloadedFile)
	})

}
