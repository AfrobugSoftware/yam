package ygl

import (
	"fmt"
	"log"
	"unsafe"
	"yam/y3d"
	"yam/yecs"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	MESS_PASS = iota
	LIGHT_PASS
)

const (
	ALBEDO = iota
	WORLD_POSITION
	NORMAL
	SPECULAR
)

const (
	LIGHT_UBO_BINDING_INDEX    = 0
	SPATIAL_SSBO_BINDING_INDEX = 1
	DRAWCOMMAND_BINDING_INDEX  = 2
)

type DrawCommand struct {
	Count         uint32
	InstanceCount uint32
	FirstIndex    uint32
	BaseVertex    int32
	BaseIntance   uint32
}

type DeferredRenderer struct {
	Gbuf            *Framebuffer
	PassTechnique   []uint32
	TotalLights     uint32
	LightUBO        uint32
	MeshVBO         uint32
	IndxBuffer      uint32
	LightBlockSize  int32
	SpatialSSBO     uint32
	DrawCommandSSBO uint32

	emptyVao uint32
}

func CreateDeferredRenderer(width, height, dwidth, dheight int32) *DeferredRenderer {
	dr := &DeferredRenderer{}
	dr.Gbuf = CreateFrameBuffer(width, height, false)
	//albedo, world position and normal, for now
	dr.Gbuf.Bind()
	for range 3 {
		tex := CreateRenderTexture(width, height, gl.RGB32F)
		dr.Gbuf.AttachTexture(tex)
	}
	depthTex := CreateDepthTexture(dwidth, dheight)
	dr.Gbuf.AttachDepthTexture(depthTex)
	if !dr.Gbuf.CheckComplete() {
		DestoryFrameBuffer(dr.Gbuf)
		panic("could not setup g-buffer")
	}
	dr.Gbuf.DrawBuffers()
	dr.PassTechnique = append(dr.PassTechnique, AddProgramSource("assets/shaders/mesh.vert", "assets/shaders/gBufferWrite.frag")) //mesh pass
	dr.PassTechnique = append(dr.PassTechnique, AddProgramSource("assets/shaders/screenQuad.vert", "assets/shaders/glight.frag")) //light pass
	dr.SetupLightUBO()
	dr.Gbuf.Unbind()
	gl.CreateVertexArrays(1, &dr.emptyVao)
	return dr
}

func (dr *DeferredRenderer) UpdateSpatialsSSBO(w *yecs.World) {
	spatials := w.Query([]yecs.ComponentId{yecs.SpatialComponent, yecs.TransformComponent})
	buffer := make([]y3d.Mat4, len(spatials))
	for i, e := range spatials {
		transform := w.GetComponent(e, yecs.TransformComponent).(yecs.Transform)
		buffer[i] = transform.World
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(&buffer[0])), len(buffer)*16*int(unsafe.Sizeof(float32(0))))
	gl.NamedBufferSubData(dr.SpatialSSBO, 0, len(b), gl.Ptr(&b[0]))
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, SPATIAL_SSBO_BINDING_INDEX, dr.SpatialSSBO)
}

func (dr *DeferredRenderer) Draw(w *yecs.World) {
	dr.Gbuf.Bind()
	gl.UseProgram(dr.PassTechnique[MESS_PASS])
	camera := w.Query([]yecs.ComponentId{yecs.CameraComponent})
	spatials := w.Query([]yecs.ComponentId{yecs.SpatialComponent,
		yecs.TransformComponent,
		yecs.RenderStateComponent,
		yecs.MaterialSurfaceComponent})
	if len(camera) == 0 {
		log.Println("no camera attached to scene")
		return
	}
	MainCam := w.GetComponent(camera[0], yecs.CameraComponent).(yecs.Camera)
	view := MainCam.GetViewTransformation()
	proj := MainCam.GetProjectionTransformation()
	projView := proj.Mul(view)
	err := AssignUniformMat4(dr.PassTechnique[MESS_PASS], "projView", projView)
	if err != nil {
		log.Println(err)
	}
	for _, e := range spatials {
		s := w.GetComponent(e, yecs.SpatialComponent).(yecs.Spatial)
		if w.HasComponent(e, yecs.BoxComponent) {
			box := w.GetComponent(e, yecs.BoxComponent).(yecs.Box)
			if MainCam.CullView(box, projView) {
				continue
			}
		}
		r := w.GetComponent(e, yecs.RenderStateComponent).(yecs.RenderState)
		gl.BindVertexArray(s.VertArray)
		m := w.GetComponent(e, yecs.MaterialSurfaceComponent).(yecs.MaterialSurface)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, m.Diffuse)
		err := AssignUniformInt32(dr.PassTechnique[MESS_PASS], "material.diffuse", 0)
		if err != nil {
			log.Println(err)
		}
		t := w.GetComponent(e, yecs.TransformComponent).(yecs.Transform)
		err = AssignUniformMat4(dr.PassTechnique[MESS_PASS], "world", t.World)
		if err != nil {
			log.Println(err)
		}
		r.SetupRenderState()
		gl.DrawElements(gl.TRIANGLES, s.IndxCount, gl.UNSIGNED_SHORT, nil)
		gl.BindVertexArray(0)
	}
	dr.Gbuf.Unbind()

	gl.UseProgram(dr.PassTechnique[LIGHT_PASS])
	lights := w.Query([]yecs.ComponentId{yecs.LightComponent})
	uploadSize := len(lights) * int(unsafe.Sizeof(yecs.Light{}))
	if uploadSize > int(dr.LightBlockSize) {
		uploadSize = int(dr.LightBlockSize)
	}
	gl.BindBuffer(gl.UNIFORM_BUFFER, dr.LightUBO)
	ptr := gl.MapBufferRange(gl.UNIFORM_BUFFER, 0, uploadSize, gl.MAP_WRITE_BIT|gl.MAP_PERSISTENT_BIT|gl.MAP_COHERENT_BIT)
	if ptr == nil {
		panic(fmt.Errorf("cannot get mapped buffer"))
	}
	dst := unsafe.Slice((*[16]float32)(ptr), len(lights))
	for i := range lights {
		light := w.GetComponent(lights[i], yecs.LightComponent).(yecs.Light)
		copy(dst[i][0:4], light.Diffuse.ToSlice())
		copy(dst[i][4:8], light.Ambient.ToSlice())
		copy(dst[i][8:12], light.Specular.ToSlice())
		copy(dst[i][12:16], light.Pos.ToSlice())
	}
	gl.UnmapBuffer(gl.UNIFORM_BUFFER)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, 0, dr.LightUBO)
	SetActiveTex(dr.Gbuf.Color[0], 0)
	SetActiveTex(dr.Gbuf.Color[1], 1)
	SetActiveTex(dr.Gbuf.Color[2], 2)
	AssignUniformInt32(dr.PassTechnique[LIGHT_PASS], "surface.diffuse", 0)
	AssignUniformInt32(dr.PassTechnique[LIGHT_PASS], "surface.position", 1)
	AssignUniformInt32(dr.PassTechnique[LIGHT_PASS], "surface.normal", 2)
	AssignUniformVec3(dr.PassTechnique[LIGHT_PASS], "cameraPos", MainCam.Pos)
	AssignUniformInt32(dr.PassTechnique[LIGHT_PASS], "lightCount", int32(len(lights)))

	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.BLEND)
	gl.BindVertexArray(dr.emptyVao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.UseProgram(0)
}

func (dr *DeferredRenderer) SetupLightUBO() {
	p := dr.PassTechnique[LIGHT_PASS]
	blockId := gl.GetUniformBlockIndex(p, gl.Str("LightSet\x00"))
	gl.UniformBlockBinding(p, blockId, 0)

	gl.CreateBuffers(1, &dr.LightUBO)
	gl.GetActiveUniformBlockiv(p, blockId, gl.UNIFORM_BLOCK_DATA_SIZE, &dr.LightBlockSize)
	gl.NamedBufferStorage(dr.LightUBO, int(dr.LightBlockSize), nil, gl.MAP_WRITE_BIT|gl.MAP_PERSISTENT_BIT|gl.MAP_COHERENT_BIT)
}

func ValidateProgram(shader uint32) {
	gl.ValidateProgram(shader)

	var status int32
	gl.GetProgramiv(shader, gl.VALIDATE_STATUS, &status)
	if status == gl.FALSE {
		var logLen int32
		gl.GetProgramiv(shader, gl.INFO_LOG_LENGTH, &logLen)
		logBuf := make([]byte, logLen)
		gl.GetProgramInfoLog(shader, logLen, nil, &logBuf[0])
		log.Fatalf("[SHADER] validation failed: %s", string(logBuf))
	}
	log.Printf("[SHADER] validation OK")
}
