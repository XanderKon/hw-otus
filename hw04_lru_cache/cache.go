package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type cacheItem struct {
	key   Key
	value interface{}
}

func (l *lruCache) Set(key Key, value interface{}) bool {
	if item, ok := l.items[key]; ok {
		item.Value = cacheItem{key: key, value: value}
		l.queue.MoveToFront(item)

		return true
	}
	item := l.addToQueue(cacheItem{key: key, value: value})
	l.items[key] = item

	return false
}

// Add new Item to Queue and delete last if necessary.
func (l *lruCache) addToQueue(value interface{}) *ListItem {
	item := l.queue.PushFront(value)

	if l.queue.Len() > l.capacity {
		delete(l.items, l.queue.Back().Value.(cacheItem).key)
		l.queue.Remove(l.queue.Back())
	}

	return item
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := l.items[key]; ok {
		l.queue.MoveToFront(item)

		return item.Value.(cacheItem).value, true
	}
	return nil, false
}

func (l *lruCache) Clear() {
	l.items = make(map[Key]*ListItem)
	l.queue = NewList()
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
