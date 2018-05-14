package memstorage

import (
	"time"
)

//Get 获取内存暂存中的对象
func (mem *MemStorage) Get(key string) (interface{}, bool) {
	val, existed := mem.storage[key]
	if existed {
		return val.val, existed
	}
	return nil, existed
}

//GetnRenew 获取内存暂存中的对象并更新 TTL 计时
func (mem *MemStorage) GetnRenew(key string, ttl time.Duration) (interface{}, bool) {
	mem.SetTTL(key, ttl)
	return mem.Get(key)
}

//Fetch 获取内存暂存中的对象
func (mem *MemStorage) Fetch(key string) interface{} {
	val, existed := mem.storage[key]
	if existed {
		return val.val
	}
	return nil
}

//FetchnRenew 获取内存暂存中的对象并更新 TTL 计时
func (mem *MemStorage) FetchnRenew(key string, ttl time.Duration) interface{} {
	mem.SetTTL(key, ttl)
	return mem.Fetch(key)
}

//Set 设置内存暂存中的对象
func (mem *MemStorage) Set(key string, val interface{}, ttl ...time.Duration) {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	if temp := mem.get(key); temp != nil && temp.timer != nil {
		temp.timer.Stop()
	}
	span := time.Duration(0)
	if len(ttl) > 0 {
		span = ttl[0]
	}
	element := mem.set(key, val, span)
	mem.storage[key] = element
}

//SetTTL 设置内存暂存中的对象的生命周期
func (mem *MemStorage) SetTTL(key string, ttl time.Duration) bool {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	if ttl > time.Duration(0) {
		obj, existed := mem.storage[key]
		if !existed {
			return false
		}
		obj.startObserve(ttl)
	}
	return true
}

//GetTTL 获取内存暂存中的对象的生命周期
func (mem *MemStorage) GetTTL(key string) time.Duration {
	if elem := mem.get(key); elem != nil {
		if time.Since(elem.expireTime) < 0 {
			return -time.Since(elem.expireTime)
		}
	}
	return 0
}

//SetRange 设置内存暂存中的一批对象
func (mem *MemStorage) SetRange(kvs map[string]interface{}) {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	for k, v := range kvs {
		mem.set(k, v, 0)
	}
}

//Clear 清空暂存
func (mem *MemStorage) Clear() {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	mem.identity = calcMD5(nil)
	mem.storage = map[string]*memElement{}
}

//Delete 删除某组KV
func (mem *MemStorage) Delete(key string) {
	mem.lock.Lock()
	defer mem.lock.Unlock()
	delete(mem.storage, key)
}

//Keys 获取所有的键
func (mem *MemStorage) Keys() []string {
	keys := []string{}
	for k := range mem.storage {
		keys = append(keys, k)
	}
	return keys
}

//startObserve 开始观察
func (obj *memElement) startObserve(ttl time.Duration, actions ...func()) {
	if ttl > time.Duration(0) {
		if obj.timer != nil {
			obj.timer.Stop()
		}
		obj.expireTime = time.Now().Add(ttl)
		obj.timer = time.AfterFunc(ttl, func() {
			obj.timer.Stop()
			obj.action()
		})
	} else if obj.timer != nil {
		obj.timer.Stop()
	}
}
