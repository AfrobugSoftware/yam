package ygl

import (
	"yam/y3d"
	"yam/yecs"

	"github.com/go-gl/gl/v4.3-core/gl"
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
	program := AddProgramSource("assets/shaders/grid.vert", "assets/shaders/grid.frag")

	return &Grid{
		Size:            100,
		CellSize:        0.025,
		ColorThin:       y3d.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1.0},
		ColorThick:      y3d.Vec4{X: 0.0, Y: 0.0, Z: 0.0, W: 1.0},
		BackgroundColor: y3d.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 1.0},
		Program:         program,
	}
}

func (g *Grid) Draw(w *yecs.World, emptyVao uint32) {
	camEntity := w.Query([]yecs.ComponentId{yecs.CameraComponent})
	if len(camEntity) == 0 {
		return
	}
	camera := w.GetComponent(camEntity[0], yecs.CameraComponent).(yecs.Camera)
	view := camera.GetViewTransformation()
	proj := camera.GetProjectionTransformation()
	projView := proj.Mul(view)
	camPos := camera.Pos

	SetActiveProgram(g.Program)

	AssignUniformFloat32(g.Program, "gridSize", g.Size)
	AssignUniformFloat32(g.Program, "gridCellSize", g.CellSize)
	AssignUniformFloat32(g.Program, "gridMinPixelsBetweeenCells", g.MinPixelBetweenCells)
	AssignUniformVec4(g.Program, "gridColorThin", g.ColorThin)
	AssignUniformVec4(g.Program, "gridColorThick", g.ColorThick)
	AssignUniformMat4(g.Program, "projView", projView)
	AssignUniformVec3(g.Program, "camPos", camPos)

	gl.BindVertexArray(emptyVao)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
	SetActiveProgram(0)
}
