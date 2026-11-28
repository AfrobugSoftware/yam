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

func IntegrateRigidBodies(w *ywd.World, dt float32) {
	entites := w.Query(
		ywd.PositionComponent,
		ywd.OrientationComponent,
		ywd.VelocityComponent,
		ywd.AccelerationComponent,
		ywd.ForceComponent,
		ywd.InverseMassComponent,
		ywd.AngularVelocityComponent,
		ywd.TorqueComponent,
		ywd.InertiaTensorComponent,
	)
	var wg sync.WaitGroup
	for _, e := range entites {
		wg.Go(func() {
			position := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)
			orientation := w.GetComponent(e, ywd.OrientationComponent).(y3d.Quaternion)
			velocity := w.GetComponent(e, ywd.VelocityComponent).(y3d.Vec3)
			inverseMass := w.GetComponent(e, ywd.InverseMassComponent).(float32)
			accleration := w.GetComponent(e, ywd.AccelerationComponent).(y3d.Vec3)
			force := w.GetComponent(e, ywd.ForceComponent).(y3d.Vec3)
			angularVelocity := w.GetComponent(e, ywd.AngularVelocityComponent).(y3d.Vec3)
			torque := w.GetComponent(e, ywd.TorqueComponent).(y3d.Vec3)
			inertiaTensor := w.GetComponent(e, ywd.InertiaTensorComponent).(InertiaTensor)

			inertiaTensor.Transform(orientation)
			retAcc := accleration
			retAcc = retAcc.AddScale(force, inverseMass)
			angularAcceleration := inertiaTensor.World.MulVec3(torque)

			velocity = velocity.AddScale(retAcc, dt)
			velocity = y3d.Smul(velocity, float32(math.Pow(float64(0.99), float64(dt))))

			angularVelocity = angularVelocity.AddScale(angularAcceleration, dt)
			angularVelocity = y3d.Smul(angularVelocity, float32(math.Pow(float64(0.99), float64(dt))))

			position = position.AddScale(velocity, dt)
			orientation = orientation.AddScale(angularVelocity, dt)

			force = y3d.Vec3{}
			torque = y3d.Vec3{}

			w.SetComponent(e, ywd.PositionComponent, position)
			w.SetComponent(e, ywd.OrientationComponent, orientation)
			w.SetComponent(e, ywd.VelocityComponent, velocity)
			w.SetComponent(e, ywd.AngularVelocityComponent, angularVelocity)
			w.SetComponent(e, ywd.ForceComponent, force)
			w.SetComponent(e, ywd.TorqueComponent, torque)
		})
	}
	wg.Wait()
}
