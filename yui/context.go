package yui

import "yam/y3d"

type Context struct {
	Fonts        []*Font
	Layers       []*QuadArray
	PixelPerInch float32
	Width        float32
	Height       float32
	Root         *UIElement

	proj y3d.Mat4
	view y3d.Mat4
}

func NewContext(width, height float32) *Context {
	c := &Context{
		Width:  width,
		Height: height,
		Fonts:  make([]*Font, 0),
		Layers: make([]*QuadArray, 0),
		Root: NewUIElement(y3d.Vec2{}, y3d.Vec2{X: width, Y: height}).Set(UI_FIT | UI_INVISIBLE).
			SetPadding(y3d.Vec4{X: 2, Y: 2, Z: 2, W: 2}),
	}
	c.prepare(width, height, 0.1, 1000)
	return c
}

func (io *Context) prepare(width, height, near, far float32) {
	io.view = y3d.Identity
	io.proj = y3d.Identity
	io.proj[0] = 2.0 / float32(width)
	io.proj[5] = 2.0 / float32(height)
	io.proj[10] = 1.0 / (far - near)
	io.proj[11] = -near * (1.0 / (far - near))
	io.proj[15] = 1.0
	//flip the y axis for 2D rendering``
	io.view[5] = -1.0
	io.view[12] = -float32(width) + float32(width)*0.5
	io.view[13] = float32(height) - float32(height)*0.5
	io.view[14] = near + 0.1
}

func (io *Context) ReSize(size y3d.Vec2) {
	io.Width = size.X
	io.Height = size.Y
	io.prepare(io.Width, io.Height, 0.1, 1000.0)
}

func (io *Context) Draw() {
	Layout(io.Root)
	vp := io.proj.Mul(io.view)
	Draw(io.Root, vp)
}

func (io *Context) Update(dt float32) {

}
