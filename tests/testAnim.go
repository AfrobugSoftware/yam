package yam

import (
	"strings"
	"yam/y3d"
	"yam/ygame"

	"github.com/veandco/go-sdl2/sdl"
)

func CreateLevel() *ygame.TileMap {
	var level strings.Builder
	//64x14
	level.WriteString(`................................................................`)
	level.WriteString(`................................................................`)
	level.WriteString(`................................................................`)
	level.WriteString(`................................................................`)
	level.WriteString(`................................................................`)
	level.WriteString(`............................########............................`)
	level.WriteString(`....................#######.........#####.......................`)
	level.WriteString(`................####............................................`)
	level.WriteString(`#########################################....#################..`)
	level.WriteString(`.........................................#......................`)
	level.WriteString(`...........................................##...................`)
	level.WriteString(`..............................................#####.............`)
	level.WriteString(`................................................................`)
	level.WriteString(`................................................................`)
	return ygame.NewTileMap(level.String(), 64, 14, 64, 64, &ygame.Camara{})
}

var l *ygame.TileMap

func NewTestAnim() {
	g, err := ygame.NewGame("test anim", width, height)
	if err != nil {
		panic(err)
	}
	err = g.ResourceManager.LoadSurfaceBundle("redhat", []string{"assets/redhatidle.png",
		"assets/redhatrun.png",
		"assets/redhatjump.png"})
	if err != nil {
		panic(err)
	}
	an := &ygame.AnimSprite{}
	an.Pos = y3d.Vec3{
		X: float64(g.Renderer.Width) / 2,
		Y: float64(g.Renderer.Height) / 2,
	}
	an.FrameHeight = 64
	an.FrameWidth = 64
	an.FPS = 10

	an.Texture, err = g.ResourceManager.GetTextureBundle("redhat")
	if err != nil {
		panic(err)
	}
	an.SelectSheet(0)
	an.NumFrames = 10

	an2 := &ygame.AnimSprite{}
	inputComponent := &ygame.InputComponent{
		MaxForwardSpeed:      150,
		MaxAngularSpeed:      5,
		ForwardKey:           sdl.SCANCODE_RIGHT,
		BackwardKey:          sdl.SCANCODE_LEFT,
		ClockwiseRotationKey: -1,
		CounterClockwiseKey:  -1,
	}
	inputComponent.Owner = an2
	an2.Components = append(an2.Components, inputComponent)
	an2.StateMachine = &ygame.StateMachine{
		PrevState: 0,
		CurState:  0,
		OnEnter:   EnterAnimStateTable,
		OnLeave:   ExitAnimStateTable,
		OnUpdate:  UpdateAnimStateTable,
		Owner:     an2,
	}
	an2.Pos = y3d.Vec3{
		X: float64(g.Renderer.Width)/2 + 128,
		Y: float64(g.Renderer.Height) / 2,
	}
	an2.FrameHeight = 64
	an2.FrameWidth = 64
	an2.FPS = 10
	an2.Texture, err = g.ResourceManager.GetTextureBundle("redhat")
	if err != nil {
		panic(err)
	}
	an2.StateMachine.ChangeState(Idle)
	l = CreateLevel()
	g.AddActor(an)
	g.AddActor(an2)
	g.AddLevel(l)
	g.Run()
}
