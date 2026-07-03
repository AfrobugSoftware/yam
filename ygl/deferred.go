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
	SCREEN_TEST
	GLTF_TEST
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
	MeshVBO         uint32
	IndxBuffer      uint32
	LightBlockSize  int32
	SpatialSSBO     uint32
	DrawCommandSSBO uint32
	IGrid           *Grid
	DrawGrid        bool

	LightUBO    uint32
	LightBuffer [][16]float32

	TransformUBO    uint32
	TransfromBuffer [][16]float32

	ClearColor []float32
	emptyVao   uint32
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
	dr.Gbuf.DrawBuffers()

	if !dr.Gbuf.CheckComplete() {
		DestoryFrameBuffer(dr.Gbuf)
		panic("could not setup g-buffer")
	}
	dr.PassTechnique = append(dr.PassTechnique, AddProgramSource("assets/shaders/mesh.vert", "assets/shaders/gbufferWrite.frag"))     //mesh pass
	dr.PassTechnique = append(dr.PassTechnique, AddProgramSource("assets/shaders/screenQuad.vert", "assets/shaders/glight.frag"))     //light pass
	dr.PassTechnique = append(dr.PassTechnique, AddProgramSource("assets/shaders/screenQuad.vert", "assets/shaders/screenQuad.frag")) //screen test
	dr.PassTechnique = append(dr.PassTechnique, AddProgramSource("assets/shaders/gltf.vert", "assets/shaders/gltf.frag"))             //screen test
	dr.SetupTransformUBO()
	dr.SetupLightUBO()

	dr.Gbuf.Unbind()
	gl.CreateVertexArrays(1, &dr.emptyVao)
	dr.IGrid = NewGrid()
	dr.DrawGrid = true
	return dr
}

func (dr *DeferredRenderer) DrawGLTF(w *yecs.World) {
	entities := w.Query([]yecs.ComponentId{yecs.MeshEntryComponent, yecs.TransformComponent})
	meshId := -1
	var mesh *yecs.Mesh
	gl.UseProgram(dr.PassTechnique[GLTF_TEST])
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, e := range entities {
		me := w.GetComponent(e, yecs.MeshEntryComponent).([]yecs.MeshEntry)
		for _, m := range me {
			if meshId != m.MeshId {
				mesh = w.GetMesh(m.MeshId)
				if mesh == nil {
					panic(fmt.Errorf("no mesh but this entry but has Id"))
				}
				meshId = m.MeshId
				mesh.Bind()
			}
			gl.DrawElementsBaseVertex(gl.TRIANGLES,
				int32(m.NumIndices),
				gl.UNSIGNED_SHORT,
				gl.Ptr(uintptr(m.BaseIndex)), int32(m.BaseVertex))
		}
	}
	if mesh != nil {
		mesh.Unbind()
	}
	gl.UseProgram(0)
}

func (dr *DeferredRenderer) LightPass(w *yecs.World, camPos y3d.Vec3) {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(dr.PassTechnique[LIGHT_PASS])

	lights := w.Query([]yecs.ComponentId{yecs.LightComponent})
	for i := range lights {
		if i >= len(dr.LightBuffer) {
			break
		}
		light := w.GetComponent(lights[i], yecs.LightComponent).(yecs.Light)
		lighData := light.ToUBO()
		copy(dr.LightBuffer[i][0:16], lighData[:])
	}
	gl.BindTextureUnit(0, dr.Gbuf.Color[0])
	gl.BindTextureUnit(1, dr.Gbuf.Color[1])
	gl.BindTextureUnit(2, dr.Gbuf.Color[2])

	c := camPos.ToSlice()
	gl.Uniform4fv(3, 1, &c[0])
	gl.Uniform1i(4, int32(len(lights)))

	gl.Enable(gl.DEPTH_TEST)
	gl.Disable(gl.BLEND)

	gl.BindVertexArray(dr.emptyVao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
	gl.UseProgram(0)
}

func (dr *DeferredRenderer) MeshPass(w *yecs.World) {
	dr.Gbuf.Bind()
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	if !dr.DrawGrid {
		dr.IGrid.Draw(w, dr.emptyVao)
	}
	gl.UseProgram(dr.PassTechnique[MESS_PASS])
	camera := w.Query([]yecs.ComponentId{yecs.CameraComponent})
	spatials := w.Query([]yecs.ComponentId{yecs.MeshEntryComponent,
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
	gl.UniformMatrix4fv(0, 1, false, &projView[0])
	meshId := -1
	var mesh *yecs.Mesh

	//upload world
	// for id, e := range spatials {
	// 	t := w.GetComponent(e, yecs.TransformComponent).(yecs.Transform)
	// 	copy(dr.TransfromBuffer[id][:], t.World[:])
	// }

	for _, e := range spatials {
		me := w.GetComponent(e, yecs.MeshEntryComponent).([]yecs.MeshEntry)
		r := w.GetComponent(e, yecs.RenderStateComponent).(yecs.RenderState)
		r.SetupRenderState()
		ms := w.GetComponent(e, yecs.MaterialSurfaceComponent).(yecs.MaterialSurface)
		gl.BindTextureUnit(0, ms.Diffuse)

		//gl.Uniform1ui(1, uint32(id))
		t := w.GetComponent(e, yecs.TransformComponent).(yecs.Transform)
		gl.UniformMatrix4fv(1, 1, false, &t.World[0])
		normalMat := t.World.CalulateNormalMatrix()
		gl.UniformMatrix4fv(2, 1, false, &normalMat[0])
		for _, m := range me {
			if meshId != m.MeshId {
				mesh = w.GetMesh(m.MeshId)
				if mesh == nil {
					panic(fmt.Errorf("no mesh but this entry but has Id"))
				}
				meshId = m.MeshId
				mesh.Bind()
			}

			gl.DrawElementsBaseVertex(gl.TRIANGLES,
				int32(m.NumIndices),
				gl.UNSIGNED_INT,
				gl.Ptr(uintptr(m.BaseIndex)), int32(m.BaseVertex))
		}
	}
	gl.UseProgram(0)
	dr.Gbuf.Unbind()
	dr.LightPass(w, MainCam.Pos)
}

func (dr *DeferredRenderer) Draw(w *yecs.World) {
	dr.MeshPass(w)
}

func (dr *DeferredRenderer) SetupLightUBO() {
	p := dr.PassTechnique[LIGHT_PASS]
	blockId := gl.GetUniformBlockIndex(p, gl.Str("LightSet\x00"))
	gl.UniformBlockBinding(p, blockId, 0)

	gl.CreateBuffers(1, &dr.LightUBO)
	gl.GetActiveUniformBlockiv(p, blockId, gl.UNIFORM_BLOCK_DATA_SIZE, &dr.LightBlockSize)
	gl.NamedBufferStorage(dr.LightUBO, int(dr.LightBlockSize), nil, gl.MAP_WRITE_BIT|gl.MAP_PERSISTENT_BIT|gl.MAP_COHERENT_BIT)

	gl.BindBuffer(gl.UNIFORM_BUFFER, dr.LightUBO)
	ptr := gl.MapBuffer(gl.UNIFORM_BUFFER, gl.WRITE_ONLY)
	if ptr == nil {
		panic(fmt.Errorf("cannot get mapped buffer"))
	}
	dr.LightBuffer = unsafe.Slice((*[16]float32)(ptr), 100)
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, 0, dr.LightUBO)
}

func (dr *DeferredRenderer) SetupTransformUBO() {
	blocksize := 16000
	gl.CreateBuffers(1, &dr.TransformUBO)
	gl.NamedBufferStorage(dr.TransformUBO, int(blocksize), nil, gl.MAP_WRITE_BIT|gl.MAP_PERSISTENT_BIT|gl.MAP_COHERENT_BIT)
	gl.BindBuffer(gl.UNIFORM_BUFFER, dr.TransformUBO)

	ptr := gl.MapBuffer(gl.UNIFORM_BUFFER, gl.WRITE_ONLY)
	if ptr == nil {
		panic(fmt.Errorf("cannot get mapped buffer"))
	}
	dr.TransfromBuffer = unsafe.Slice((*[16]float32)(ptr), 1000)
	gl.BindBuffer(gl.UNIFORM_BUFFER, 0)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, 1, dr.TransformUBO)
}

func (dr *DeferredRenderer) ShutDown() {
	DestoryFrameBuffer(dr.Gbuf)
	gl.DeleteBuffers(1, &dr.DrawCommandSSBO)

	gl.BindBuffer(gl.UNIFORM_BUFFER, dr.LightUBO)
	gl.UnmapBuffer(gl.UNIFORM_BUFFER)
	gl.DeleteBuffers(1, &dr.LightUBO)

	gl.DeleteVertexArrays(1, &dr.emptyVao)
}

func (dr *DeferredRenderer) ClearBuffers() {
	gl.ClearBufferfv(gl.COLOR, 0, &dr.ClearColor[0])
	gl.ClearBufferfi(gl.DEPTH_STENCIL, 0, 1.0, 0)
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
