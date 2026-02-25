package yecs

import (
	"errors"
	"sort"
)

var (
	ErrorNoComponentInArchType = errors.New("no component in archtype")
)

type ArchTypeId uint64
type ComponentId uint64
type EntityId uint64

type Type []ComponentId

func (a Type) Len() int {
	return len((a))
}

func (a Type) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Type) Less(i, j int) bool {
	return a[i] < a[j]
}

func (a Type) Sort() {
	sort.Sort(a)
}

type ArchType struct {
	Id    ArchTypeId
	Type  Type
	Edges map[ComponentId]ArchTypeEdge
}

type ArchTypeEdge struct {
	Add    *ArchType
	Remove *ArchType
}

type Record struct {
	ArchType ArchType
	Row      int
}

// maps an archtype to a column, the integer is the column with the data
type ArchTypeSet map[ArchTypeId]ComponentId
type EntityIndex map[EntityId]*Record
type ComponentIndex map[ComponentId]ArchTypeSet

type World struct {
	EntIndx   EntityIndex
	CompIndex ComponentIndex
	Storage   map[ComponentId]any //holds storage for the components as *Column[T] types
}

type Column[T any] struct {
	slice map[ArchTypeId][]T
}

func HasComponent(w *World, ent EntityId, comp ComponentId) bool {
	archType := w.EntIndx[ent].ArchType
	archTypeSet := w.CompIndex[comp]
	_, exists := archTypeSet[archType.Id]
	return exists
}

func GetComponent[T any](w *World, ent EntityId, comp ComponentId) (*T, error) {
	record := w.EntIndx[ent]
	archType := record.ArchType
	archTypeSet := w.CompIndex[comp]
	colId, exists := archTypeSet[archType.Id]
	if !exists {
		return nil, ErrorNoComponentInArchType
	}
	col := w.Storage[colId].(*Column[T])
	return &col.slice[archType.Id][record.Row], nil
}

func MoveEntity[T any](w *World, dst, src *ArchType, ent EntityId, comp ComponentId) *T {
	return nil
}

func AddComponent[T any](w *World, ent EntityId, comp ComponentId) *T {
	record := w.EntIndx[ent]
	archType := record.ArchType
	nextArch := archType.Edges[comp].Add
	return MoveEntity[T](w, nextArch, &archType, ent, comp)
}

func RemoveComponent(w *World, ent EntityId, comp ComponentId) {

}
