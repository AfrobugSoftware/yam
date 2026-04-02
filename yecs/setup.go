package yecs

import "yam/y3d"

var (
	TransformComponent    = RegisterComponent[Transform]()
	ANodeComponent        = RegisterComponent[ANode]()
	OBBComponent          = RegisterComponent[y3d.OBB]()
	SphereComponent       = RegisterComponent[y3d.Sphere]()
	AABBComponent         = RegisterComponent[y3d.AABB]()
	InputComponent        = RegisterComponent[Input]()
	ControlComponent      = RegisterComponent[Control]()
	SpriteComponent       = RegisterComponent[Sprite]()
	RenderStateComponent  = RegisterComponent[RenderState]()
	MoveComponent         = RegisterComponent[Move]()
	StateComponent        = RegisterComponent[StateMachine]()
	CameraComponent       = RegisterComponent[Camera]()
	FollowCameraComponent = RegisterComponent[FollowCamera]()
	OrbitCameraComponent  = RegisterComponent[OrbitCamera]()
	SplineCamaraComponent = RegisterComponent[SplineCamera]()
	AudioComponent        = RegisterComponent[AudioData]()
	NavigatorComponent    = RegisterComponent[Navigator]()
)
