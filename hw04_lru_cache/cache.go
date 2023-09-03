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

func (l *lruCache) Set(key Key, value interface{}) bool {
	if item, ok := l.items[key]; ok {
		item.Value = value
		l.queue.MoveToFront(item)
		return true
	}
	item := l.addToQueue(value)
	l.items[key] = item

	return false
}

// Add new Item to Queue and delete last if necessary.
func (l *lruCache) addToQueue(value interface{}) *ListItem {
	item := l.queue.PushFront(value)

	if l.queue.Len() > l.capacity {
		l.queue.Remove(l.queue.Back())
		// TODO: how to delete it without key?
		// delete(l.items, l.queue.Back())
	}

	return item
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := l.items[key]; ok {
		l.queue.MoveToFront(item)
		return item.Value, true
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
