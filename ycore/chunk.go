package ycore

import (
	"encoding/gob"
)

const (
	PROJECT_FILE_CHUNK_IDENTIFIER   = 0xAABBEEFF
	ANIMATION_CHUNK_IDENTIFIER      = 0x00000001
	SHADER_CHUNK_IDENTIFIER         = 0x00000002
	SPATIAL_CHUNK_IDENTIFIER        = 0x00000003
	RENDER_MANAGER_CHUNK_IDENTIFIER = 0x00000004
	VERTEX_MANAGER_CHUNK_IDENTIFIER = 0x00000005
	SKIN_MANAGER_CHUNK_IDENTIFIER   = 0x00000006
	END_CHUNK_IDENTIFIER            = 0xFFFFFFFF
)

type Chunk interface {
	Write(w *gob.Encoder) error
	Read(r *gob.Decoder) error
}

type ChunkHeader struct {
	id uint32
}

func (p *ChunkHeader) Write(w *gob.Encoder) error {
	return w.Encode(*p)
}

func (p *ChunkHeader) Read(r *gob.Decoder) error {
	return r.Decode(p)
}

func (c *ChunkHeader) Id() uint32 {
	return c.id
}
