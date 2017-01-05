package surface

import "honnef.co/go/js/dom"

// New initializes a new drawing surface given a html canvas element and returns it.
func New(canvas *dom.HTMLCanvasElement) *S {
	s := &S{
		canvas: canvas,
		ctx:    canvas.GetContext2d(),
	}
	s.init()
	return s
}

type S struct {
	canvas *dom.HTMLCanvasElement
	ctx    *dom.CanvasRenderingContext2D

	drawing bool
}

func (s *S) handleEvent(event string, handler func(x, y int)) {
	s.canvas.AddEventListener(event, false, func(de dom.Event) {
		rect := s.canvas.GetBoundingClientRect()
		me := de.(*dom.MouseEvent)
		x := me.ClientX - int(rect.Left)
		y := me.ClientY - int(rect.Top)
		handler(x, y)
	})
}

func (s *S) init() {
	s.ctx.LineWidth = 5
	s.ctx.LineJoin = "round"
	s.ctx.LineCap = "round"
	s.ctx.StrokeStyle = "black"

	s.handleEvent("mousemove", s.handleMove)
	s.handleEvent("mousedown", s.handleDown)
	s.handleEvent("mouseup", s.handleUpOut)
	s.handleEvent("mouseout", s.handleUpOut)
}

func (s *S) handleMove(x, y int) {
	if !s.drawing {
		return
	}
	s.ctx.LineTo(x, y)
	s.ctx.Stroke()
}

func (s *S) handleUpOut(x, y int) {
	s.drawing = false
	s.ctx.ClosePath()
}

func (s *S) handleDown(x, y int) {
	s.ctx.BeginPath()
	s.ctx.MoveTo(x, y)
	s.drawing = true
}

func (s *S) Clear() {
	s.ctx.ClearRect(0, 0, s.canvas.Width, s.canvas.Height)
	// document.getElementById("canvasimg").style.display = "none";
}

func (s *S) Data() string {
	return s.canvas.Call("toDataURL").String()
}
