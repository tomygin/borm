package cache

import (
	"container/list"
	"sync"

	"github.com/tomygin/box/session"
)

type Cache struct {
	mu        sync.RWMutex
	maxBytes  int64
	usedBytes int64

	ll    *list.List
	cache map[string]*list.Element

	// 通常用于写入数据到数据库
	// 后期可能会用接口型函数修改
	Ondelete func(Key string, Value string, s *session.Session)
	writer   *session.Session
}

type Item struct {
	Key   string
	Value string
}

// New maxBytes设置为零就代表不限制缓存大小
func New(maxBytes int64, ondelete func(key, value string, s *session.Session), writer *session.Session) *Cache {
	if ondelete == nil {
		ondelete = saveToDatabase
	}

	writer.Model(&Item{})
	if !writer.IsExistTable() {
		writer.CreateTable()
	}

	return &Cache{
		maxBytes: maxBytes,
		Ondelete: ondelete,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
		writer:   writer,
	}
}

func (c *Cache) Get(key string) (value string, isok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if ele, ok := c.cache[key]; ok {

		//双向队列的尾部是热数据
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*Item)
		return kv.Value, true

	}

	//如果有数据库，从数据库里面找
	if c.writer != nil {
		item := Item{}
		if err := c.writer.Where("Key = ?", key).First(&item); err == nil {
			return item.Value, true
		}
	}
	return
}

// Move 将队列头的一条冷数据清除，至于清除的数据怎么处理交给Ondelete回调函数
func (c *Cache) move() {

	//双向队列头是冷数据
	ele := c.ll.Front()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*Item)
		delete(c.cache, kv.Key)

		c.usedBytes -= int64(len(kv.Key)) + int64(len(kv.Value))
		if c.Ondelete != nil {
			c.Ondelete(kv.Key, kv.Value, c.writer)
		}
	}
}

// Add向cache中添加一条热数据，或者修改一条数据
// 如果超过最大缓存值将进行写入数据库操作
func (c *Cache) Add(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	//修改数据
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*Item)

		c.usedBytes += int64(len(value)) - int64(len(kv.Value))
		kv.Value = value
	} else {
		// 添加数据
		ele := c.ll.PushBack(&Item{key, value})
		c.cache[key] = ele
		c.usedBytes += int64(len(key)) + int64(len(value))
	}

	//清除冷数据
	for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
		c.move()
	}

}

// Flush 清除缓存区的所有缓存
func (c *Cache) Flush() {
	for c.usedBytes > 0 {
		c.move()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}

// saveToDatabase MOVE后的默认函数
func saveToDatabase(key, value string, s *session.Session) {
	s.Insert(&Item{Key: key, Value: value})
}

