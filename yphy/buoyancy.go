package yphy

import (
	"yam/y3d"
	"yam/ywd"
)

type BuoyancyForce struct {
	World         *ywd.World
	MaxDepth      float32
	Volume        float32
	WaterHeight   float32
	LiquidDensity float32
}

func NewBuoyancyForce(world *ywd.World, maxDepth, volume, waterHeight, liquidDensity float32) *BuoyancyForce {
	return &BuoyancyForce{
		World:         world,
		MaxDepth:      maxDepth,
		Volume:        volume,
		WaterHeight:   waterHeight,
		LiquidDensity: liquidDensity,
	}
}

func (b *BuoyancyForce) Force(w *ywd.World, e ywd.EntityId, dt float32) {
	position := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)
	force := w.GetComponent(e, ywd.ForceComponent).(y3d.Vec3)
	depth := position.Y
	if depth >= b.WaterHeight+b.MaxDepth {
		return
	}
	f := y3d.Vec3{}
	if depth <= b.WaterHeight-b.MaxDepth {
		f = y3d.Vec3{X: 0, Y: b.LiquidDensity * b.Volume, Z: 0}
	} else {
		fy := b.LiquidDensity * b.Volume * (depth - b.MaxDepth - b.WaterHeight) / (2 * b.MaxDepth)
		f = y3d.Vec3{X: 0, Y: fy, Z: 0}
	}
	force = y3d.Add(force, f)
	w.SetComponent(e, ywd.ForceComponent, force)
}
