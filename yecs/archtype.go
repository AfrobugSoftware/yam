package yecs

import (
	"errors"
	"fmt"
	"log"
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

	storageMu     sync.Mutex
	storageBuffer = map[ComponentId]any{} //contain all the storage
)

type Component any
type System interface {
	Query() []ComponentId
	Update(w *World, dt float64, entities []EntityId)
}

type Storage[T any] struct {
	Store map[ArchetypeId][]T
}

func RegisterComponent[T Component]() ComponentId {
	componentMu.Lock()
	defer componentMu.Unlock()

	t := reflect.TypeFor[*T]()
	if id, ok := componentRegistry[t]; ok {
		return id
	}
	componentCounter++
	componentRegistry[t] = componentCounter
	storageBuffer[componentCounter] = Storage[T]{
		Store: make(map[ArchetypeId][]T),
	}
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
	id       ArchetypeId
	key      archetypeKey
	entities []EntityId
}

func newArchetype(id ArchetypeId, key archetypeKey) *Archetype {
	return &Archetype{
		id:       id,
		key:      key,
		entities: []EntityId{},
	}
}

// storage methods
func appendToStorage[T any](data T, comp ComponentId, a *Archetype) {
	storageMu.Lock()
	defer storageMu.Unlock()

	str, ok := storageBuffer[comp]
	if !ok {
		storageBuffer[comp] = Storage[T]{
			Store: make(map[ArchetypeId][]T),
		}
		str = storageBuffer[comp]
	}
	store := reflect.ValueOf(str)
	mapVal := store.Field(0)
	key := reflect.ValueOf(a.id)

	slc := mapVal.MapIndex(key)
	if !slc.IsValid() {
		sliceType := mapVal.Type().Elem()
		slc = reflect.MakeSlice(sliceType, 0, 10)
	}
	newSlc := reflect.Append(slc, reflect.ValueOf(data))
	mapVal.SetMapIndex(key, newSlc)
}

func removeFromStorage(row, last int, comp ComponentId, a *Archetype) {
	storageMu.Lock()
	defer storageMu.Unlock()

	str, ok := storageBuffer[comp]
	if !ok {
		return
	}
	store := reflect.ValueOf(str)
	mapVal := store.Field(0)
	key := reflect.ValueOf(a.id)
	slc := mapVal.MapIndex(key)
	if row != last {
		slc.Index(row).Set(slc.Index(last))
	}
	mapVal.SetMapIndex(key, slc.Slice(0, last))
}

func getStorageLen(comp ComponentId, a *Archetype) int {
	storageMu.Lock()
	defer storageMu.Unlock()

	store := reflect.ValueOf(storageBuffer[comp])
	if store.Kind() != reflect.Struct {
		panic(fmt.Errorf("invalid type used for component storage"))
	}
	slc := store.Field(0).MapIndex(reflect.ValueOf(a.id))
	return slc.Len()
}

func getFromStorage(row int, comp ComponentId, a *Archetype) Component {
	storageMu.Lock()
	defer storageMu.Unlock()

	store := reflect.ValueOf(storageBuffer[comp])
	if store.Kind() != reflect.Struct {
		panic(fmt.Errorf("invalid type used for component storage"))
	}
	slc := store.Field(0).MapIndex(reflect.ValueOf(a.id))
	return slc.Index(row).Interface()
}

func SetToStorage(row int, data any, comp ComponentId, a *Archetype) {
	storageMu.Lock()
	defer storageMu.Unlock()
	str, ok := storageBuffer[comp]
	if !ok {
		log.Printf("no component: %d", comp)
		return
	}
	store := reflect.ValueOf(str)
	mapVal := store.Field(0)
	key := reflect.ValueOf(a.id)
	if store.Kind() != reflect.Struct {
		panic(fmt.Errorf("invalid type used for component storage"))
	}
	slc := store.Field(0).MapIndex(key)
	slc.Index(row).Set(reflect.ValueOf(data))
	mapVal.SetMapIndex(key, slc)
}

func gatherComponentsFromStorage(a *Archetype, row int) map[ComponentId]Component {
	comps := make(map[ComponentId]Component)
	for _, cid := range a.key {
		store := reflect.ValueOf(storageBuffer[cid])
		if store.Kind() != reflect.Struct {
			panic(fmt.Errorf("invalid type used for component storage"))
		}
		slc := store.Field(0).MapIndex(reflect.ValueOf(a.id))
		comps[cid] = slc.Index(row).Interface()
	}
	return comps
}

func (a *Archetype) addEntity(entity EntityId, comps map[ComponentId]Component) int {
	row := len(a.entities)
	a.entities = append(a.entities, entity)
	for cid, comp := range comps {
		appendToStorage(comp, cid, a)
	}
	return row
}

func (a *Archetype) removeEntity(row int) (swappedEntity EntityId, wasSwapped bool) {
	last := len(a.entities) - 1
	// Swap with last
	if row != last {
		a.entities[row] = a.entities[last]
		for _, cid := range a.key {
			removeFromStorage(row, last, cid, a)
		}
		swappedEntity = a.entities[row]
		wasSwapped = true
	}
	a.entities = a.entities[:last]
	return
}

func (a *Archetype) getComponent(row int, cid ComponentId) Component {
	if row >= getStorageLen(cid, a) {
		return nil
	}
	return getFromStorage(row, cid, a)
}

func (a *Archetype) setComponent(row int, cid ComponentId, comp Component) {
	SetToStorage(row, comp, cid, a)
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
	return gatherComponentsFromStorage(record.archetype, record.row)
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
	//assuming the the component we want to set is in this archetype
	//possible bug here, if it does not have to component, it would create one
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

func (w *World) ProcessInput(keyState []uint8) {
	entities := w.Query([]ComponentId{InputComponent})
	for _, e := range entities {
		in := Input{
			KeyState: []uint8{},
		}
		copy(in.KeyState, keyState)
		w.SetComponent(e, InputComponent, in)
	}
}
