package yphy

type ParticleContactResolver struct {
	Iterations int
	Contacts   []Contact
}

func NewParticleContactResolver(iterations int) *ParticleContactResolver {
	return &ParticleContactResolver{
		Iterations: iterations,
	}
}
