package linkedlists

// doubly linked list
type Node struct {
	key  int
	val  int
	prev *Node
	next *Node
}

type LRUCache struct {
	cache    map[int]*Node
	capacity int
	head     *Node
	tail     *Node
}

func Constructor(capacity int) LRUCache {
	head := &Node{}
	tail := &Node{}
	head.next = tail
	tail.prev = head
	return LRUCache{
		cache:    make(map[int]*Node),
		capacity: capacity,
		head:     head,
		tail:     tail,
	}
}

func (this *LRUCache) insert(node *Node) {
	prev := this.tail.prev
	prev.next = node
	node.prev = prev
	node.next = this.tail
	this.tail.prev = node
}

func (this *LRUCache) remove(node *Node) {
	prev := node.prev
	next := node.next
	prev.next = next
	next.prev = prev
}

func (this *LRUCache) Get(key int) int {
	if node, found := this.cache[key]; found {
		// move to MRU position -> tail position
		this.remove(node)
		this.insert(node)
		return node.val
	}
	return -1
}

func (this *LRUCache) Put(key int, value int) {
	if node, found := this.cache[key]; found {
		node.val = value
		// move to MRU position -> tail position
		this.remove(node)
		this.insert(node)
		return
	}
	newNode := &Node{
		key: key,
		val: value,
	}
	this.cache[key] = newNode
	this.insert(newNode)

	// capacity exceeded
	if len(this.cache) > this.capacity {
		lru := this.head.next
		this.remove(lru)
		delete(this.cache, lru.key)
	}
}

/**
 * Your LRUCache object will be instantiated and called as such:
 * obj := Constructor(capacity);
 * param_1 := obj.Get(key);
 * obj.Put(key,value);
 */
