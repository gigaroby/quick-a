package main

import (
	"fmt"

	"time"

	"github.com/gigaroby/quick-a/frontend/session"
	"github.com/gigaroby/quick-a/frontend/surface"
	"honnef.co/go/js/dom"
)

var (
	document = dom.GetWindow().Document()
	canvas   = document.GetElementByID("board").(*dom.HTMLCanvasElement)

	messages = document.GetElementByID("messages").(*dom.HTMLPreElement)

	game  = document.GetElementByID("game").(*dom.HTMLDivElement)
	wait  = document.GetElementByID("wait").(*dom.HTMLDivElement)
	final = document.GetElementByID("final").(*dom.HTMLDivElement)

	fReload  = document.GetElementByID("f_reload").(*dom.HTMLInputElement)
	fCorrect = document.GetElementByID("f_correct").(*dom.HTMLSpanElement)
	fWrong   = document.GetElementByID("f_wrong").(*dom.HTMLSpanElement)
	fMessage = document.GetElementByID("f_message").(*dom.HTMLSpanElement)

	wInstructions = document.GetElementByID("w_instructions").(*dom.HTMLSpanElement)
	wMessage      = document.GetElementByID("w_message").(*dom.HTMLSpanElement)
	wReady        = document.GetElementByID("w_ready").(*dom.HTMLInputElement)

	predictions  = document.GetElementByID("predictions").(*dom.HTMLSpanElement)
	instructions = document.GetElementByID("instructions").(*dom.HTMLSpanElement)
	countdown    = document.GetElementByID("countdown").(*dom.HTMLSpanElement)

	clear = document.GetElementByID("clr").(*dom.HTMLInputElement)
)

func showError(err error) {
	messages.SetInnerHTML(fmt.Sprintf("[%s] %s", time.Now().String(), err.Error()))
}

func main() {
	surface := surface.New(canvas)
	sess, err := session.New(6, surface, instructions, predictions, countdown, wInstructions, wMessage, fCorrect, fWrong, fMessage, game, wait, final)
	if err != nil {
		showError(err)
		return
	}

	dom.GetWindow().AddEventListener("resize", false, func(d dom.Event) {
		surface.Resize()
	})

	wReady.AddEventListener("click", false, func(d dom.Event) {
		go func() {
			sess.NextRound()
		}()
	})

	clear.AddEventListener("click", false, func(d dom.Event) {
		surface.Clear()
		messages.SetInnerHTML("")
	})

	fReload.AddEventListener("click", false, func(d dom.Event) {
		dom.GetWindow().Location().Call("reload")
	})
}
