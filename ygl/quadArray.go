package ygl

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
)

// might remove when we do instancing
const (
	MAX_QUADS             = 1000
	NUM_VERTICES          = 6
	SPRITE_BUFFER_BINDING = 0
)

const (
	POS_VB = iota
	QUAD_ID_VB
	NUM_BUFFERS
)

type QuadArray struct {
	Vbo     uint32
	Buffers [NUM_BUFFERS]uint32
	Ubo     uint32

	block                []byte
	uboBlockSize         int32
	basePosOffset        int32
	widthHeightOffset    int32
	texCoordsOffset      int32
	texWidthHeightOffset int32
}

func CreateQuadArray() *QuadArray {
	maxQuads := MAX_QUADS
	vertices := []float32{
		0.0, 0.0, //bottom left
		0.0, 1.0, //top left
		1.0, 1.0, //top right
		0.0, 0.0, //bottom left
		1.0, 1.0, // top right
		1.0, 0.0, //bottom right
	}
	vertVec := make([]float32, 2*NUM_VERTICES*maxQuads)
	for i := range maxQuads {
		copy(vertVec[i*NUM_VERTICES*2:], vertices[:])
	}

	var vbo, buf, idBuf uint32
	gl.CreateVertexArrays(1, &vbo)
	gl.BindVertexArray(vbo)

	gl.CreateBuffers(1, &buf)
	gl.NamedBufferStorage(buf, int(unsafe.Sizeof(float32(0)))*len(vertVec), gl.Ptr(&vertVec[0]), 0)

	gl.VertexArrayAttribBinding(vbo, POS_VB, 0)
	gl.VertexArrayVertexBuffer(vbo, 0, buf, 0, int32(unsafe.Sizeof(float32(0))*2))
	gl.VertexArrayAttribFormat(vbo, POS_VB, 2, gl.FLOAT, false, 0)
	gl.EnableVertexArrayAttrib(vbo, POS_VB)

	idVec := make([]uint32, maxQuads*NUM_VERTICES)
	for i := range maxQuads {
		for j := range NUM_VERTICES {
			idVec[int(i)*NUM_VERTICES+j] = uint32(i)
		}
	}
	gl.CreateBuffers(1, &idBuf)
	gl.NamedBufferStorage(idBuf, int(unsafe.Sizeof(uint32(0)))*len(idVec), gl.Ptr(&idVec[0]), 0)
	gl.VertexArrayAttribBinding(vbo, QUAD_ID_VB, 1)
	gl.VertexArrayVertexBuffer(vbo, 1, idBuf, 0, int32(unsafe.Sizeof(uint32(0))))
	gl.VertexArrayAttribIFormat(vbo, QUAD_ID_VB, 1, gl.UNSIGNED_INT, 0)
	gl.EnableVertexArrayAttrib(vbo, QUAD_ID_VB)

	q := &QuadArray{
		Vbo: vbo,
	}
	q.Buffers[POS_VB] = buf
	q.Buffers[QUAD_ID_VB] = idBuf
	return q
}

func (q *QuadArray) SetupUBO(shader uint32) {
	blockIdx := gl.GetUniformBlockIndex(shader, gl.Str("QuadInfo\x00"))
	gl.UniformBlockBinding(shader, blockIdx, SPRITE_BUFFER_BINDING)
	names, free := gl.Strs("BasePos\x00", "WidthHeight\x00", "TexCoords\x00", "TexWidthHeight\x00")
	var indices [4]uint32
	gl.GetUniformIndices(shader, 4, names, &indices[0])
	free()
	var offsets [4]int32
	gl.GetActiveUniformsiv(shader, 4, &indices[0], gl.UNIFORM_OFFSET, &offsets[0])
	q.basePosOffset = offsets[0]
	q.widthHeightOffset = offsets[1]
	q.texCoordsOffset = offsets[2]
	q.texWidthHeightOffset = offsets[3]

	gl.CreateBuffers(1, &q.Ubo)
	gl.GetActiveUniformBlockiv(shader, blockIdx, gl.UNIFORM_BLOCK_DATA_SIZE, &q.uboBlockSize)
	q.block = make([]byte, q.uboBlockSize)
	gl.NamedBufferStorage(q.Ubo, int(q.uboBlockSize), gl.Ptr(&q.block[0]), gl.DYNAMIC_STORAGE_BIT)
}

func (q *QuadArray) Draw(numQuads int) {
	if numQuads >= MAX_QUADS {
		numQuads = MAX_QUADS - 1
	}
	glCheck("BEFORE bindVertexArray")
	gl.BindVertexArray(q.Vbo)
	if err := gl.GetError(); err != gl.NO_ERROR {
		log.Printf("[DRAW] GL error BEFORE DrawArrays: 0x%x\n", err)
	}
	gl.DrawArrays(gl.TRIANGLES, 0, int32(numQuads*NUM_VERTICES))
	gl.BindVertexArray(0)
}

func DestroyQuadArray(q *QuadArray) {
	gl.DeleteVertexArrays(1, &q.Vbo)
	gl.DeleteBuffers(1, &q.Ubo)
	gl.DeleteBuffers(2, &q.Buffers[0])
}

func (q *QuadArray) SetQuad(idx int, ndcX, ndcY,
	tileWidthNDC, tileHeightNDC,
	UBase, VBase,
	texUSize, texVSize float32) {
	if idx >= MAX_QUADS {
		return
	}
	basePos := ViewAsFloat32Pairs(q.block, int(q.basePosOffset), MAX_QUADS)
	widthHeight := ViewAsFloat32Pairs(q.block, int(q.widthHeightOffset), MAX_QUADS)
	texCoords := ViewAsFloat32Pairs(q.block, int(q.texCoordsOffset), MAX_QUADS)
	texWidthHeight := ViewAsFloat32Pairs(q.block, int(q.texWidthHeightOffset), MAX_QUADS)

	basePos[idx][0] = ndcX
	basePos[idx][1] = ndcY
	widthHeight[idx][0] = tileWidthNDC
	widthHeight[idx][1] = tileHeightNDC
	texCoords[idx][0] = UBase
	texCoords[idx][1] = VBase
	texWidthHeight[idx][0] = texUSize
	texWidthHeight[idx][1] = texVSize
}

func (q *QuadArray) Update() {
	//move data to ubo
	gl.NamedBufferSubData(q.Ubo, 0, int(q.uboBlockSize), gl.Ptr(&q.block[0]))
	gl.BindBufferBase(gl.UNIFORM_BUFFER, SPRITE_BUFFER_BINDING, q.Ubo)
}

func ViewAsFloat32Pairs(buf []byte, byteOffset, count int) [][2]float32 {
	const pairSize = int(unsafe.Sizeof([2]float32{}))
	required := byteOffset + count*pairSize
	if required > len(buf) {
		panic(fmt.Errorf("ViewAsFloat32Pairs: slice out of bounds, req: %d, len : %d\n", required, len(buf)))
	}
	ptr := unsafe.Pointer(&buf[byteOffset])
	return unsafe.Slice((*[2]float32)(ptr), count)
}
