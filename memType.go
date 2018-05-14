package memstorage

import (
	"sync"
	"time"
)

//memElement 内存元素
type memElement struct {
	val        interface{}
	expireTime time.Time

	timer  *time.Timer
	action func()
}

//MemStorage 内存对象暂存
type MemStorage struct {
	lock    sync.Mutex
	storage map[string]*memElement

	identity string
	//order []string
}

//New 创建新的暂存
func New() *MemStorage {
	return &MemStorage{
		storage:  map[string]*memElement{},
		identity: calcMD5(nil),
	}
}
