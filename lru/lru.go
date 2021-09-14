package lru

import "container/list"

type Cache struct { // lru cache
	maxBytes  int64
	nBytes    int64 // 已使用内存
	list      *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) // 移除记录的回调函数
}

type entry struct {
	key   string
	value Value
}

type Value interface { // len: how many bytes it takes
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		list:      list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.list.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	ele := c.list.Back()
	if ele != nil {
		c.list.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // update exist value
		c.list.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len()) // update nBytes
		kv.value = value
	} else { // add
		ele := c.list.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nBytes { // 超过内存限制淘汰缓存
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.list.Len()
}
