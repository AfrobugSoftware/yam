package yecs

import "yam/y3d"

var (
	TransformComponent   = RegisterComponent[Transform]()
	ANodeComponent       = RegisterComponent[ANode]()
	OBBComponent         = RegisterComponent[y3d.OBB]()
	SphereComponent      = RegisterComponent[y3d.Sphere]()
	BoxComponent         = RegisterComponent[Box]()
	InputComponent       = RegisterComponent[Input]()
	ControlComponent     = RegisterComponent[Control]()
	SpatialComponent     = RegisterComponent[Spatial]()
	RenderStateComponent = RegisterComponent[RenderState]()
	MoveComponent        = RegisterComponent[Move]()
	StateComponent       = RegisterComponent[StateMachine]()
	CameraComponent      = RegisterComponent[Camera]()
	AudioComponent       = RegisterComponent[AudioData]()
	NavigatorComponent   = RegisterComponent[Navigator]()
	TagComponent         = RegisterComponent[Tag]()
	MaterialComponent    = RegisterComponent[Material]()
	LightComponent       = RegisterComponent[Light]()
)
