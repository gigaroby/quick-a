package main

import (
	"net/http"
	"net/url"

	"honnef.co/go/js/dom"
)

var (
	document = dom.GetWindow().Document()
	canvas   = document.GetElementByID("board").(*dom.HTMLCanvasElement)
	submit   = document.GetElementByID("sub").(*dom.HTMLInputElement)
	clear    = document.GetElementByID("clr").(*dom.HTMLInputElement)

	console = dom.GetWindow().Console()
)

func main() {
	surface := &drawSurface{
		canvas: canvas,
		ctx:    canvas.GetContext2d(),
	}
	surface.Init()
	submit.AddEventListener("click", false, func(d dom.Event) {
		go func() {
			data := surface.Data()
			http.PostForm("image", url.Values{
				"image": []string{data},
			})
		}()
	})
	clear.AddEventListener("click", false, func(d dom.Event) {
		go func() {
			surface.Clear()
		}()
	})
}
