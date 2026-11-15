package yui

type Context struct {
	Fonts        []*Font
	PixelPerInch float32
	Width        float32
	Height       float32
}

func NewContext() *Context {
	return &Context{
		Fonts: make([]*Font, 0),
	}
}
