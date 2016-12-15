package main

import "honnef.co/go/js/dom"

type drawSurface struct {
	canvas *dom.HTMLCanvasElement
	ctx    *dom.CanvasRenderingContext2D
	// previous x and y
	px, py  int
	drawing bool
}

func (d *drawSurface) handleEvent(event string, handler func(px, py int, cx, cy int)) {
	d.canvas.AddEventListener(event, false, func(de dom.Event) {
		me := de.(*dom.MouseEvent)
		cx := me.ClientX - int(canvas.OffsetLeft())
		cy := me.ClientY - int(canvas.OffsetTop())
		handler(d.px, d.py, cx, cy)
		d.px = cx
		d.py = cy
	})
}

func (d *drawSurface) Init() {
	d.handleEvent("mousemove", d.handleMove)
	d.handleEvent("mousedown", d.handleDown)
	d.handleEvent("mouseup", d.handleUpOut)
	d.handleEvent("mouseout", d.handleUpOut)
}

func (d *drawSurface) handleMove(px, py int, cx, cy int) {
	if !d.drawing {
		return
	}
	d.drawTo(px, py, cx, cy)
}

func (d *drawSurface) handleUpOut(px, py int, cx, cy int) {
	d.drawing = false
}

func (d *drawSurface) handleDown(px, py int, cx, cy int) {
	d.drawing = true
}

func (d *drawSurface) drawTo(px, py int, cx, cy int) {
	d.ctx.BeginPath()
	d.ctx.MoveTo(px, py)
	d.ctx.LineTo(cx, cy)
	d.ctx.StrokeStyle = "black"
	d.ctx.LineWidth = 2
	d.ctx.Stroke()
	d.ctx.ClosePath()
}

func (d *drawSurface) Clear() {
	d.ctx.ClearRect(0, 0, d.canvas.Width, d.canvas.Height)
	// document.getElementById("canvasimg").style.display = "none";
}

func (d *drawSurface) Data() string {
	return d.canvas.Call("toDataURL").String()
}
