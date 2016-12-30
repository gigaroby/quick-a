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
	// previous x and y
	px, py  int
	drawing bool
}

func (s *S) handleEvent(event string, handler func(px, py int, cx, cy int)) {
	s.canvas.AddEventListener(event, false, func(de dom.Event) {
		rect := s.canvas.GetBoundingClientRect()
		me := de.(*dom.MouseEvent)
		cx := me.ClientX - int(rect.Left)
		cy := me.ClientY - int(rect.Top)
		handler(s.px, s.py, cx, cy)
		s.px = cx
		s.py = cy
	})
}

func (s *S) init() {
	s.handleEvent("mousemove", s.handleMove)
	s.handleEvent("mousedown", s.handleDown)
	s.handleEvent("mouseup", s.handleUpOut)
	s.handleEvent("mouseout", s.handleUpOut)
}

func (s *S) handleMove(px, py int, cx, cy int) {
	if !s.drawing {
		return
	}
	s.drawTo(px, py, cx, cy)
}

func (s *S) handleUpOut(px, py int, cx, cy int) {
	s.drawing = false
}

func (s *S) handleDown(px, py int, cx, cy int) {
	s.drawing = true
}

func (s *S) drawTo(px, py int, cx, cy int) {
	s.ctx.BeginPath()
	s.ctx.MoveTo(px, py)
	s.ctx.LineTo(cx, cy)
	s.ctx.StrokeStyle = "black"
	s.ctx.LineWidth = 5
	s.ctx.Stroke()
	s.ctx.ClosePath()
}

func (s *S) Clear() {
	s.ctx.ClearRect(0, 0, s.canvas.Width, s.canvas.Height)
	// document.getElementById("canvasimg").style.display = "none";
}

func (s *S) Data() string {
	return s.canvas.Call("toDataURL").String()
}
