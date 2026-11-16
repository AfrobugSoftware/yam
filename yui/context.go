package yui

type Context struct {
	Fonts        []*Font
	Layers       []*QuadArray
	PixelPerInch float32
	Width        float32
	Height       float32
}

func NewContext() *Context {
	return &Context{
		Fonts:  make([]*Font, 0),
		Layers: make([]*QuadArray, 0),
	}
}
