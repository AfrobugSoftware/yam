package yecs

import "yam/y3d"

var (
	TransformComponent   = RegisterComponent[Transform]()
	ANodeComponent       = RegisterComponent[ANode]()
	OBBComponent         = RegisterComponent[y3d.OBB]()
	AABBComponent        = RegisterComponent[y3d.AABB]()
	InputComponent       = RegisterComponent[Input]()
	ControlComponent     = RegisterComponent[Control]()
	SpriteComponent      = RegisterComponent[Sprite]()
	RenderStateComponent = RegisterComponent[RenderState]()
	MoveComponent        = RegisterComponent[Move]()
	StateComponent       = RegisterComponent[StateMachine]()
	CameraComponent      = RegisterComponent[Camera]()
	AudioComponent       = RegisterComponent[AudioData]()
)
