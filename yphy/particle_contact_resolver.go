package yphy

type ParticleContactResolver struct {
	Iterations int
	Contacts   []ParticleContact
}

func NewParticleContactResolver(iterations int) *ParticleContactResolver {
	return &ParticleContactResolver{
		Iterations: iterations,
	}
}
