package memstorage

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
	"time"
)

//get 通过 key 获取内存元素
func (mem *MemStorage) get(key string) *memElement {
	return mem.storage[key]
}

//set 设置 key 所映射的内存元素（非线程安全）
func (mem *MemStorage) set(key string, val interface{}, ttl time.Duration, actions ...func()) *memElement {
	if mem.identity == "" {
		mem.identity = calcMD5(nil)
		mem.storage = map[string]*memElement{}
	}
	element := &memElement{
		val:    val,
		action: combineActions(mem.delete(key), actions...),
	}
	element.startObserve(ttl)
	mem.storage[key] = element
	return element
}

//delete 删除内存中元素的别名方法（线程安全）
func (mem *MemStorage) delete(key string) func() {
	identity := mem.identity
	return func() {
		if mem.identity == identity {
			mem.Delete(key)
		}
	}
}

//combineActions 方法合并
func combineActions(defaultFunc func(), actions ...func()) func() {
	if len(actions) == 0 {
		return defaultFunc
	}
	if len(actions) == 1 {
		return actions[0]
	}
	return func() {
		for _, action := range actions {
			action()
		}
	}
}

func calcMD5(b []byte) string {
	if b == nil {
		b = make([]byte, 48)
		io.ReadFull(rand.Reader, b)
	}
	h := md5.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}
