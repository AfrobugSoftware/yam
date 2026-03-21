package main

import (
	"fmt"
	"yam/y3d"
	"yam/yecs"
)

func main() {
	fmt.Println("yam yum!")

	world := yecs.NewWorld()
	ent1 := world.NewEntity()
	ent2 := world.NewEntity()

	world.AddComponent(ent1, yecs.TransformComponent, yecs.Transform{
		Position: y3d.Vec3{X: 1, Y: 2, Z: 3},
	})

	world.AddComponent(ent2, yecs.TransformComponent, yecs.Transform{
		Position: y3d.Vec3{X: 4, Y: 5, Z: 6},
	})

	trans1 := world.GetComponent(ent1, yecs.TransformComponent).(yecs.Transform)
	trans2 := world.GetComponent(ent2, yecs.TransformComponent).(yecs.Transform)

	fmt.Printf("Ent1 X:%f\n", trans1.Position.X)
	fmt.Printf("Ent2 X:%f\n", trans2.Position.X)

	world.SetComponent(ent2, yecs.TransformComponent, yecs.Transform{
		Position: y3d.Vec3{X: 10, Y: 5, Z: 6},
	})
	trans2 = world.GetComponent(ent2, yecs.TransformComponent).(yecs.Transform)
	fmt.Printf("Again Ent2 X:%f\n", trans2.Position.X)

	world.RemoveComponent(ent1, yecs.TransformComponent)
	world.DestroyEntity(ent1)
}
