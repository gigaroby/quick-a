package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"encoding/json"

	"time"

	"github.com/gigaroby/quick-a/frontend/surface"
	"github.com/gigaroby/quick-a/model"
	"honnef.co/go/js/dom"
)

var (
	document = dom.GetWindow().Document()
	canvas   = document.GetElementByID("board").(*dom.HTMLCanvasElement)
	submit   = document.GetElementByID("sub").(*dom.HTMLInputElement)
	clear    = document.GetElementByID("clr").(*dom.HTMLInputElement)

	messages  = document.GetElementByID("messages").(*dom.HTMLPreElement)
	countdown = document.GetElementByID("cd").(*dom.HTMLPreElement)

	tableContainer = document.GetElementByID("tc").(*dom.HTMLDivElement)
)

var (
	tc = `<table>
	  <thead>
        <tr>
		  <td>category</td>
		  <td>precision</td>
		</tr>
	  </thead>
	  <tbody>
	  	{{ range $idx, $elem := . }}
		  <tr>
		    <td>{{ $elem.CategoryName }}</td>
		    <td>{{ $elem.Confidence }}</td>
		  </tr>
        {{ end }}
	  </tbody>
	</table>`
	table = template.Must(template.New("table").Parse(tc))
)

func showError(err error) {
	messages.SetInnerHTML(err.Error())
}

func renderTable(pred model.Predictions) {
	buf := new(bytes.Buffer)
	table.Execute(buf, pred)
	tableContainer.SetInnerHTML(buf.String())
}

func main() {
	println("ready")
	surface := surface.New(canvas)
	submit.AddEventListener("click", false, func(d dom.Event) {
		go func() {
			res, err := http.PostForm("classify", url.Values{
				"image":    []string{surface.Data()},
				"category": []string{"cat"},
			})
			if err != nil {
				showError(err)
				return
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				showError(fmt.Errorf("[%s] failed to contact backend, response code was %d %s", time.Now().String(), res.StatusCode, http.StatusText(res.StatusCode)))
				return
			}

			predictions := model.Predictions{}
			if err = json.NewDecoder(res.Body).Decode(&predictions); err != nil {
				showError(err)
				return
			}
			renderTable(predictions)
		}()
	})
	clear.AddEventListener("click", false, func(d dom.Event) {
		go func() {
			surface.Clear()
		}()
	})
}
