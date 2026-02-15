package yutil

import "yam/ygame"

type SparseSet struct {
	dense  []ygame.Actor
	sparse []uint32
}
