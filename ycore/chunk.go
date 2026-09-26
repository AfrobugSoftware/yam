package ycore

import (
	"encoding/binary"
	"io"
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
	Write(w io.Writer) error
	Read(r io.Reader) error
}

type ChunkHeader struct {
	ClassId  uint32
	ObjectId uint64
}

func (p *ChunkHeader) Write(w io.Writer) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf, p.ClassId)
	binary.LittleEndian.PutUint64(buf, p.ObjectId)
	return nil
}

func (p *ChunkHeader) Read(r io.Reader) error {
	err := binary.Read(r, binary.LittleEndian, &p.ClassId)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.LittleEndian, &p.ObjectId)
	if err != nil {
		return err
	}
	return nil
}

func (c *ChunkHeader) ObjId() uint64 {
	return c.ObjectId
}
