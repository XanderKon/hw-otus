package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	first *ListItem
	last  *ListItem
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.first
}

func (l *list) Back() *ListItem {
	return l.last
}

func (l *list) PushFront(v interface{}) *ListItem {
	newItem := ListItem{Value: v}
	oldFirst := l.first
	l.first = &newItem

	if oldFirst != nil {
		l.first.Next = oldFirst
		oldFirst.Prev = &newItem
	} else {
		l.last = &newItem
	}

	l.len++

	return &newItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	newItem := ListItem{Value: v}
	oldLast := l.last
	l.last = &newItem

	if oldLast != nil {
		l.last.Prev = oldLast
		oldLast.Next = &newItem
	} else {
		l.first = &newItem
	}

	l.len++

	return &newItem
}

func (l *list) Remove(i *ListItem) {
	if i == nil {
		return
	}

	l.connectChainAfterRemoveItem(i)
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil {
		return
	}

	l.connectChainAfterRemoveItem(i)

	oldFirst := l.first
	l.first = i

	if oldFirst != nil {
		l.first.Next = oldFirst
	} else {
		l.last = l.first
	}
}

// Connect Prev and Next item in list.
func (l *list) connectChainAfterRemoveItem(i *ListItem) {
	prev := i.Prev
	next := i.Next

	if prev != nil {
		prev.Next = next
	} else {
		l.first = next
	}

	if next != nil {
		next.Prev = prev
	} else {
		l.last = prev
	}
}

func NewList() List {
	return new(list)
}
