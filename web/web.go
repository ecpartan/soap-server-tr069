package web

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anoshenko/rui"
	logger "github.com/ecpartan/soap-server-tr069/log"
	jrcp2server "github.com/ecpartan/soap-server-tr069/pkg/jrpc2"
	"github.com/ecpartan/soap-server-tr069/pkg/jrpc2/methods"
	repository "github.com/ecpartan/soap-server-tr069/repository/cache"
	"github.com/fanliao/go-promise"
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
	]
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

func (content *webSession) CreateRootView(session rui.Session) rui.View {
	view := rui.CreateViewFromText(session, gridLayoutDemoText)
	if view == nil {

		return nil
	}

	rui.Set(view, "findDevice", rui.ClickEvent, func(rui.View) {
		str := rui.GetText(view, "editWindow")

		mp := repository.GetCache().Get(str)
		arr = []string{}
		var curr string
		recurse(mp, curr)

		var result string
		for _, line := range arr {
			result += line + "\n"
		}
		rui.Set(view, "deviceLog", rui.Text, result)
	})

	rui.Set(view, "executeScript", rui.ClickEvent, func(rui.View) {
		str := rui.GetText(view, "scriptLog")

		var getScript map[string]any
		err := json.Unmarshal([]byte(str), &getScript)
		if err != nil {
			return
		}
		if script, ok := getScript["Script"].(map[string]any); ok {

			task := func() (any, error) {

				var ret []byte
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				err := jrcp2server.Instance.Server.Client.CallResult(ctx, methods.MethodAddScript, script, &ret)
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
				rui.Set(view, "scriptLog", rui.Text, err.Error())
			} else {
				rui.Set(view, "scriptLog", rui.Text, string(ret.([]byte)))
			}
		}

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
		rui.StartApp(fmt.Sprintf("%s:%d", addr, port), createHelloWorldSession, rui.AppParams{
			Title: "SOAP SERVER TR069",
		})
	}()
}
