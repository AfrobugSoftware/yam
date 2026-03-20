package yecs

import "yam/y3d"

var (
	TransfromComponent = RegisterComponent[Transform]()
	ANodeComponent     = RegisterComponent[ANode]()
	OBBComponent       = RegisterComponent[y3d.OBB]()
	ABBComponent       = RegisterComponent[y3d.AABB]()
	InputComponent     = RegisterComponent[Input]()
	ControlComponent   = RegisterComponent[Control]()
)
