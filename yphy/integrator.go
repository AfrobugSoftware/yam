package yphy

type Integrator struct{}

func NewIntegrator() *Integrator {
	return &Integrator{}
}

func (i *Integrator) LinearIntegrator(dt float32) {}
