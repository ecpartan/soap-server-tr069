package repository

import "fmt"

var globalcache *cache

type cache struct {
	nodeMap  map[string]*node
	listMap  map[int]*list
	capacity int
	min      int
}

type node struct {
	prev, next *node
	key        string
	val        map[string]any
	freq       int
}

type list struct {
	head *node
	tail *node
	size int
}

func makeList() *list {
	head := &node{}
	tail := &node{}

	head.next = tail
	tail.prev = head

	return &list{
		head: head,
		tail: tail,
	}
}

func (l *list) pushNode(node *node) {
	node.prev = l.tail.prev
	node.next = l.tail

	l.tail.prev.next = node
	l.tail.prev = node

	l.size++
}

func (l *list) deleteNode(node *node) {
	node.prev.next = node.next
	node.next.prev = node.prev
	l.size--
}

func InitCache(capacity int) {
	if capacity <= 0 {
		return
	}

	if globalcache != nil {
		return
	}

	globalcache = &cache{
		nodeMap:  make(map[string]*node),
		listMap:  make(map[int]*list),
		capacity: capacity,
		min:      0,
	}

}

func GetCache() cache {
	return *globalcache
}

func (c *cache) Get(key string) map[string]any {
	cur_node, ok := c.nodeMap[key]
	if !ok {
		return nil
	}

	list, ok := c.listMap[cur_node.freq]
	if ok {
		list.deleteNode(cur_node)
	}

	cur_node.freq++

	nextList, nextOk := c.listMap[cur_node.freq]
	if !nextOk {
		nextList = makeList()
	}

	nextList.pushNode(cur_node)
	c.listMap[cur_node.freq] = nextList

	if list.size == 0 && c.min == cur_node.freq-1 {
		c.min++
	}

	return cur_node.val
}

func (c *cache) Put(key string, value map[string]any) {
	if c.capacity == 0 {
		return
	}

	curr_node, ok := c.nodeMap[key]

	if ok {
		fmt.Println("update")
		curr_node.val = value
		c.Get(key)
		return
	}

	if len(c.nodeMap) == c.capacity {
		minList := c.listMap[c.min]
		leastFrequencyNode := minList.head.next
		minList.deleteNode(leastFrequencyNode)
		delete(c.nodeMap, leastFrequencyNode.key)
	}

	curr_node = &node{key: key, val: value, freq: 1}
	c.min = 1
	list, ok := c.listMap[curr_node.freq]
	if !ok {
		list = makeList()
	}
	list.pushNode(curr_node)
	c.listMap[curr_node.freq] = list
	c.nodeMap[key] = curr_node
}
