package ygame

import (
	"yam/y3d"
	"yam/yecs"
	"yam/ygl"

	"github.com/go-gl/gl/v4.5-core/gl"
)

type Grid struct {
	Size                 float32
	CellSize             float32
	MinPixelBetweenCells float32
	ColorThin            y3d.Vec4
	ColorThick           y3d.Vec4
	BackgroundColor      y3d.Vec4
	Program              uint32
}

func NewGrid() *Grid {
	v, err := ygl.CreateShaderFromFile("assets/shaders/grid.vert", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	f, err := ygl.CreateShaderFromFile("assets/shaders/grid.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}
	program, err := ygl.CreateProgram([]uint32{v, f})
	if err != nil {
		panic(err)
	}
	return &Grid{
		Size:            100,
		CellSize:        0.025,
		ColorThin:       y3d.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1.0},
		ColorThick:      y3d.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
		BackgroundColor: y3d.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 1.0},
		Program:         program,
	}
}

func (g *Grid) Draw(w *yecs.World) {
	camEntity := w.Query([]yecs.ComponentId{yecs.CameraComponent})
	if len(camEntity) == 0 {
		return
	}
	camera := w.GetComponent(camEntity[0], yecs.CameraComponent).(yecs.Camera)
	view := camera.GetViewTransformation()
	proj := camera.GetProjectionTransformation()
	projView := proj.Mul(view)
	camPos := camera.Pos

	ygl.AssignUniformFloat32(g.Program, "gridSize", g.Size)
	ygl.AssignUniformFloat32(g.Program, "gridCellSize", g.CellSize)
	ygl.AssignUniformFloat32(g.Program, "gridMinPixelsBetweeenCells", g.MinPixelBetweenCells)
	ygl.AssignUniformVec4(g.Program, "gridColorThin", g.ColorThin)
	ygl.AssignUniformVec4(g.Program, "gridColorThick", g.ColorThick)
	ygl.AssignUniformMat4(g.Program, "projView", projView)
	ygl.AssignUniformVec3(g.Program, "camPos", camPos)

	gl.ClearColor(g.BackgroundColor.X, g.BackgroundColor.Y, g.BackgroundColor.Z, g.BackgroundColor.W)
	ygl.SetActiveProgram(g.Program)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	ygl.SetActiveProgram(0)
}
