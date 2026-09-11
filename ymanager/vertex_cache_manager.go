package ymanager

type VertexCacheManager struct {
	ActiveCache int
	ActiveSkin  int
	Caches      map[string]VertexCache
}

func (vcm *VertexCacheManager) GetActiveCache() int {
	return vcm.ActiveCache
}
