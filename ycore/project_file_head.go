package ycore

import "encoding/gob"

type ProjectFileHead struct {
	Name    string
	Version string
	Type    string

	//other things go here
}

func (p *ProjectFileHead) Write(w *gob.Encoder) error {
	return w.Encode(*p)
}

func (p *ProjectFileHead) Read(r *gob.Decoder) error {
	return r.Decode(p)
}
