package ycore

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"reflect"
	"slices"
	"sort"
	"sync"
)

const (
	NullEntity EntityId = 0
)

const (
	MAX_ENTITES    = 100000
	MAX_COMPONENTS = 1000
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

type Component any
type System interface {
	Init()
	Query() []ComponentId
	Update(w *World, dt float64, entities []EntityId)
	Shutdown()
}

type Storage[T any] struct {
	Store []T
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
	id            ArchetypeId
	key           archetypeKey
	entities      []EntityId
	storageBuffer map[ComponentId]any //contain all the storage
}

func newArchetype(id ArchetypeId, key archetypeKey) *Archetype {
	return &Archetype{
		id:            id,
		key:           key,
		entities:      make([]EntityId, 0, MAX_ENTITES),
		storageBuffer: make(map[ComponentId]any),
	}

}

// storage methods
func appendToStorage[T any](data T, comp ComponentId, a *Archetype) {
	str, ok := a.storageBuffer[comp]
	if !ok {
		a.storageBuffer[comp] = &Storage[T]{
			Store: make([]T, 0, MAX_COMPONENTS),
		}
		str = a.storageBuffer[comp]
	}
	store := reflect.ValueOf(str).Elem()
	if store.Len() >= MAX_COMPONENTS {
		panic("too many components created for this store!")
	}
	slc := store.Field(0)
	slc.Set(reflect.Append(slc, reflect.ValueOf(data)))
}

func removeFromStorage(row int, a *Archetype) (swappedEntity EntityId, wasSwapped bool) {
	last := len(a.entities) - 1
	if row != last {
		a.entities[row] = a.entities[last]
		for _, cid := range a.key {
			str, ok := a.storageBuffer[cid]
			if !ok {
				return
			}
			store := reflect.ValueOf(str).Elem()
			slc := store.Field(0)
			slc.Index(row).Set(slc.Index(last))
		}
		swappedEntity = a.entities[row]
		wasSwapped = true
	}
	a.entities = a.entities[:last]
	for _, cid := range a.key {
		str, ok := a.storageBuffer[cid]
		if !ok {
			return
		}
		store := reflect.ValueOf(str).Elem()
		slc := store.Field(0)
		slc.Set(slc.Slice(0, last))
	}
	return
}

func DumpStorage(a *Archetype) {
	for _, s := range a.storageBuffer {
		fmt.Println(s)
	}
}

func getStorageLen(comp ComponentId, a *Archetype) int {
	store := reflect.ValueOf(a.storageBuffer[comp]).Elem()
	slc := store.Field(0)
	return slc.Len()
}

func getFromStorage(row int, comp ComponentId, a *Archetype) Component {
	store := reflect.ValueOf(a.storageBuffer[comp]).Elem()
	slc := store.Field(0)
	if row >= slc.Len() {
		panic("trying to access beyond index for storage")
	}
	return slc.Index(row).Interface()
}

func setToStorage(row int, data any, comp ComponentId, a *Archetype) {
	str, ok := a.storageBuffer[comp]
	if !ok {
		log.Printf("no component: %d", comp)
		return
	}
	store := reflect.ValueOf(str).Elem()
	slc := store.Field(0)
	slc.Index(row).Set(reflect.ValueOf(data))
}

func gatherComponentsFromStorage(a *Archetype, row int) map[ComponentId]Component {
	comps := make(map[ComponentId]Component)
	for cid := range a.storageBuffer {
		store := reflect.ValueOf(a.storageBuffer[cid]).Elem()
		slc := store.Field(0)
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

func (a *Archetype) removeEntity(row int) (EntityId, bool) {
	return removeFromStorage(row, a)
}

func (a *Archetype) getComponent(row int, cid ComponentId) Component {
	return getFromStorage(row, cid, a)
}

func (a *Archetype) setComponent(row int, cid ComponentId, comp Component) {
	setToStorage(row, comp, cid, a)
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
}

func NewWorld() *World {
	w := &World{
		entities:   make(map[EntityId]*entityRecord),
		archetypes: make(map[string]*Archetype),
	}
	return w
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
		currentComps = map[ComponentId]Component{cid: comp}
		newKey = newArchetypeKey([]ComponentId{cid})
	} else {
		currentComps = w.gatherComponents(record)
		currentComps[cid] = comp

		ids := make([]ComponentId, 0, len(currentComps))
		for id := range currentComps {
			ids = append(ids, id)
		}
		newKey = newArchetypeKey(ids)
		swapped, wasSwapped := record.archetype.removeEntity(record.row)
		if wasSwapped {
			w.entities[swapped].row = record.row
		}
	}
	arch := w.getOrCreateArchetype(newKey)
	row := arch.addEntity(entity, currentComps)
	w.entities[entity] = &entityRecord{archetype: arch, row: row}
}

func (w *World) HasComponent(entity EntityId, cid ComponentId) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	record, ok := w.entities[entity]
	if !ok || record == nil || record.archetype != nil {
		return false
	}
	return slices.Contains(record.archetype.key, cid)
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

func (w *World) UpdateComponent(entity EntityId, component ComponentId,
	updater func(comp Component)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	record := w.entities[entity]
	if record == nil {
		return
	}
	store := reflect.ValueOf(record.archetype.storageBuffer[component]).Elem()
	slc := store.Field(0)
	updater(slc.Index(record.row).Addr())
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

// would not save the component registery because the main
func (w *World) Save(e *gob.Encoder) {
	w.mu.Lock()
	defer w.mu.Unlock()

	e.Encode(w.nextEntity)
	e.Encode(w.nextArchetype)
	e.Encode(uint32(len(w.entities)))
	for ent, record := range w.entities {
		if record != nil {
			e.Encode(ent)
			e.Encode(record)
		}
	}
	e.Encode(uint32(len(w.archetypes)))
	for archkey, arch := range w.archetypes {
		if arch != nil {
			e.Encode(archkey)
			e.Encode(arch)
		}
	}
}

func (w *World) Load(d *gob.Decoder) {
	w.mu.Lock()
	defer w.mu.Unlock()

	d.Decode(&w.nextEntity)
	d.Decode(&w.nextArchetype)

	w.entities = make(map[EntityId]*entityRecord)
	var i uint32
	d.Decode(&i)
	for range i {
		var ent EntityId
		var entite *entityRecord
		d.Decode(&ent)
		d.Decode(&entite)
		w.entities[ent] = entite
	}
	w.archetypes = make(map[string]*Archetype)
	d.Decode(&i)
	for range i {
		var key string
		var arch *Archetype
		d.Decode(&key)
		d.Decode(&arch)
		w.archetypes[key] = arch
	}
}
