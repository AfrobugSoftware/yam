package yui

type Context struct {
	Fonts []*Font
}

func NewContext() *Context {
	return &Context{
		Fonts: make([]*Font, 0),
	}
}
