package yecs

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

var (
	ErrorNoComponentInArchType = errors.New("no component in archtype")
)

type ArchetypeId uint64
type ComponentId uint64
type EntityId uint64

var (
	componentMu       sync.RWMutex
	componentRegistry = map[reflect.Type]ComponentId{}
	componentCounter  ComponentId
)

type Component interface{}
type System interface {
	Query() []ComponentId
	Update(w *World, dt float64, entities []EntityId)
}

func RegisterComponent[T Component]() ComponentId {
	componentMu.Lock()
	defer componentMu.Unlock()

	t := reflect.TypeOf((*T)(nil)).Elem()
	if id, ok := componentRegistry[t]; ok {
		return id
	}
	componentCounter++
	componentRegistry[t] = componentCounter
	return componentCounter
}

func ComponentIDOf[T Component]() ComponentId {
	componentMu.RLock()
	defer componentMu.RUnlock()

	t := reflect.TypeOf((*T)(nil)).Elem()
	id, ok := componentRegistry[t]
	if !ok {
		panic(fmt.Sprintf("component %s not registered", t.Name()))
	}
	return id
}

type archetypeKey []ComponentId

func newArchetypeKey(ids []ComponentId) archetypeKey {
	key := make(archetypeKey, len(ids))
	copy(key, ids)
	sort.Slice(key, func(i, j int) bool { return key[i] < key[j] })
	return key
}

func (k archetypeKey) String() string {
	return fmt.Sprintf("%v", []ComponentId(k))
}

func (k archetypeKey) contains(id ComponentId) bool {
	for _, c := range k {
		if c == id {
			return true
		}
	}
	return false
}

func (k archetypeKey) containsAll(ids []ComponentId) bool {
	for _, id := range ids {
		if !k.contains(id) {
			return false
		}
	}
	return true
}

type Archetype struct {
	id         ArchetypeId
	key        archetypeKey
	entities   []EntityId
	components map[ComponentId][]Component // column storage per component
}

func newArchetype(id ArchetypeId, key archetypeKey) *Archetype {
	return &Archetype{
		id:         id,
		key:        key,
		entities:   []EntityId{},
		components: make(map[ComponentId][]Component),
	}
}

func (a *Archetype) addEntity(entity EntityId, comps map[ComponentId]Component) int {
	row := len(a.entities)
	a.entities = append(a.entities, entity)
	for cid, comp := range comps {
		a.components[cid] = append(a.components[cid], comp)
	}
	return row
}

func (a *Archetype) removeEntity(row int) (swappedEntity EntityId, wasSwapped bool) {
	last := len(a.entities) - 1

	if row != last {
		// Swap with last
		a.entities[row] = a.entities[last]
		for cid := range a.components {
			a.components[cid][row] = a.components[cid][last]
		}
		swappedEntity = a.entities[row]
		wasSwapped = true
	}

	// Truncate
	a.entities = a.entities[:last]
	for cid := range a.components {
		a.components[cid] = a.components[cid][:last]
	}
	return
}

func (a *Archetype) getComponent(row int, cid ComponentId) Component {
	col, ok := a.components[cid]
	if !ok || row >= len(col) {
		return nil
	}
	return col[row]
}

func (a *Archetype) setComponent(row int, cid ComponentId, comp Component) {
	a.components[cid][row] = comp
}

type entityRecord struct {
	archetype *Archetype
	row       int
}

type World struct {
	mu sync.RWMutex

	nextEntity    EntityId
	nextArchetype ArchetypeId

	entities   map[EntityId]*entityRecord
	archetypes map[string]*Archetype // keyed by archetypeKey.String()
	systems    []System
}

func NewWorld() *World {
	return &World{
		entities:   make(map[EntityId]*entityRecord),
		archetypes: make(map[string]*Archetype),
	}
}

func (w *World) NewEntity() EntityId {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.nextEntity++
	id := w.nextEntity
	w.entities[id] = nil
	return id
}

func (w *World) DestroyEntity(entity EntityId) {
	w.mu.Lock()
	defer w.mu.Unlock()

	record, ok := w.entities[entity]
	if !ok || record == nil {
		return
	}

	swapped, wasSwapped := record.archetype.removeEntity(record.row)
	if wasSwapped {
		// Update the swapped entity's row
		w.entities[swapped].row = record.row
	}
	delete(w.entities, entity)
}

func (w *World) gatherComponents(record *entityRecord) map[ComponentId]Component {
	comps := make(map[ComponentId]Component)
	for cid, col := range record.archetype.components {
		comps[cid] = col[record.row]
	}
	return comps
}

func (w *World) getOrCreateArchetype(key archetypeKey) *Archetype {
	k := key.String()
	if arch, ok := w.archetypes[k]; ok {
		return arch
	}
	w.nextArchetype++
	arch := newArchetype(w.nextArchetype, key)
	w.archetypes[k] = arch
	return arch
}

func (w *World) AddComponent(entity EntityId, cid ComponentId, comp Component) {
	w.mu.Lock()
	defer w.mu.Unlock()

	record := w.entities[entity]
	var currentComps map[ComponentId]Component
	var newKey archetypeKey

	if record == nil {
		// Entity has no archetype yet
		currentComps = map[ComponentId]Component{cid: comp}
		newKey = newArchetypeKey([]ComponentId{cid})
	} else {
		// Gather existing components
		currentComps = w.gatherComponents(record)
		currentComps[cid] = comp

		ids := make([]ComponentId, 0, len(currentComps))
		for id := range currentComps {
			ids = append(ids, id)
		}
		newKey = newArchetypeKey(ids)

		// Remove from old archetype
		swapped, wasSwapped := record.archetype.removeEntity(record.row)
		if wasSwapped {
			w.entities[swapped].row = record.row
		}
	}

	// Find or create target archetype
	arch := w.getOrCreateArchetype(newKey)
	row := arch.addEntity(entity, currentComps)
	w.entities[entity] = &entityRecord{archetype: arch, row: row}
}

func (w *World) RemoveComponent(entity EntityId, cid ComponentId) {
	w.mu.Lock()
	defer w.mu.Unlock()

	record := w.entities[entity]
	if record == nil {
		return
	}

	currentComps := w.gatherComponents(record)
	delete(currentComps, cid)

	// Remove from old archetype
	swapped, wasSwapped := record.archetype.removeEntity(record.row)
	if wasSwapped {
		w.entities[swapped].row = record.row
	}

	if len(currentComps) == 0 {
		w.entities[entity] = nil
		return
	}

	ids := make([]ComponentId, 0, len(currentComps))
	for id := range currentComps {
		ids = append(ids, id)
	}

	arch := w.getOrCreateArchetype(newArchetypeKey(ids))
	row := arch.addEntity(entity, currentComps)
	w.entities[entity] = &entityRecord{archetype: arch, row: row}
}

func (w *World) GetComponent(entity EntityId, cid ComponentId) Component {
	w.mu.RLock()
	defer w.mu.RUnlock()

	record := w.entities[entity]
	if record == nil {
		return nil
	}
	return record.archetype.getComponent(record.row, cid)
}

func (w *World) SetComponent(entity EntityId, cid ComponentId, comp Component) {
	w.mu.Lock()
	defer w.mu.Unlock()

	record := w.entities[entity]
	if record == nil {
		return
	}
	//bug here, if it does not have to component, it would create one
	record.archetype.setComponent(record.row, cid, comp)
}

// returns all entities that has at least the given component ids
func (w *World) Query(cids []ComponentId) []EntityId {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var result []EntityId
	for _, arch := range w.archetypes {
		if arch.key.containsAll(cids) {
			result = append(result, arch.entities...)
		}
	}
	return result
}

func (w *World) AddSystem(s System) {
	w.systems = append(w.systems, s)
}

func (w *World) Update(dt float64) {
	for _, s := range w.systems {
		entities := w.Query(s.Query())
		s.Update(w, dt, entities)
	}
}
