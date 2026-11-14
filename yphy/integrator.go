package yphy

type Integrator struct{}

func NewIntegrator() *Integrator {
	return &Integrator{}
}

func (i *Integrator) LinearIntegrator(r *RigidBody, dt float32) {
	r.Position = r.Position.AddScale(r.Velocity, dt)
}
