package yphy

import (
	"math"
	"sync"
	"yam/y3d"
	"yam/ywd"
)

func IntegrateParticles(w *ywd.World, dt float32) {
	entites := w.Query(
		ywd.PositionComponent,
		ywd.VelocityComponent,
		ywd.AccelerationComponent,
		ywd.ForceComponent,
		ywd.InverseMassComponent,
	)
	var wg sync.WaitGroup
	for _, e := range entites {
		wg.Go(func() {
			position := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)
			velocity := w.GetComponent(e, ywd.VelocityComponent).(y3d.Vec3)
			inverseMass := w.GetComponent(e, ywd.InverseMassComponent).(float32)
			accleration := w.GetComponent(e, ywd.AccelerationComponent).(y3d.Vec3)
			force := w.GetComponent(e, ywd.ForceComponent).(y3d.Vec3)

			position = position.AddScale(velocity, dt)
			retAcc := accleration
			retAcc = retAcc.AddScale(force, inverseMass)
			velocity = velocity.AddScale(retAcc, dt)
			velocity = y3d.Smul(velocity, float32(math.Pow(float64(0.99), float64(dt))))
			force = y3d.Vec3{}

			w.SetComponent(e, ywd.PositionComponent, position)
			w.SetComponent(e, ywd.VelocityComponent, velocity)
			w.SetComponent(e, ywd.ForceComponent, force)
		})
	}
	wg.Wait()
}
