package tchatroom

import "sync"

type BiMap struct {
	mu sync.RWMutex
	kv map[interface{}]interface{}
	vk map[interface{}]interface{}
}

func (bi *BiMap) AddPair(key, value interface{}) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	bi.kv[key] = value
	bi.vk[value] = key
}

func (bi *BiMap) RemoveByKey(key interface{}) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	if value, ok := bi.kv[key]; ok {
		delete(bi.kv, key)
		delete(bi.vk, value)
	}
}

func (bi *BiMap) RemoveByValue(value interface{}) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	if key, ok := bi.vk[value]; ok {
		delete(bi.vk, value)
		delete(bi.kv, key)
	}
}

func (bi *BiMap) Value(key interface{}) (value interface{}, ok bool) {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	value, ok = bi.kv[key]
	return value, ok
}

func (bi *BiMap) Values(keys []interface{}) (values []interface{}, oks []bool) {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	size := len(keys)
	values = make([]interface{}, size)
	oks = make([]bool, size)

	for i, key := range keys {
		values[i], oks[i] = bi.kv[key]
	}
	return values, oks
}

func (bi *BiMap) Key(value interface{}) (key interface{}, ok bool) {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	key, ok = bi.vk[value]
	return key, ok
}

func (bi *BiMap) Keys(values []interface{}) (keys []interface{}, oks []bool) {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	size := len(keys)
	keys = make([]interface{}, size)
	oks = make([]bool, size)

	for i, value := range values {
		keys[i], oks[i] = bi.vk[value]
	}
	return keys, oks
}

func NewBiMap() *BiMap {
	bi := &BiMap{
		kv: make(map[interface{}]interface{}),
		vk: make(map[interface{}]interface{}),
	}
	return bi
}
