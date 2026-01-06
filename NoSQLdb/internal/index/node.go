package index

type Key []byte
type Value []byte

// Node представляет узел B+ дерева
type Node struct {
	isLeaf   bool
	keys     []Key
	values   [][]Value
	children []*Node
	next     *Node
	parent   *Node
}

// NewNode создаёт новый узел
func NewNode(isLeaf bool) *Node {
	return &Node{
		isLeaf:   isLeaf,
		keys:     []Key{},
		values:   [][]Value{},
		children: []*Node{},
	}
}

// GetIsLeaf возвращает флаг isLeaf
func (n *Node) GetIsLeaf() bool {
	return n.isLeaf
}

// GetKeys возвращает ключи узла
func (n *Node) GetKeys() []Key {
	return n.keys
}

// GetValues возвращает значения узла
func (n *Node) GetValues() [][]Value {
	return n.values
}

// GetChildren возвращает дочерние узлы
func (n *Node) GetChildren() []*Node {
	return n.children
}

// GetNext возвращает указатель на следующий лист
func (n *Node) GetNext() *Node {
	return n.next
}

// GetParent возвращает родителя узла
func (n *Node) GetParent() *Node {
	return n.parent
}

// AddKey добавляет ключ в узел
func (n *Node) AddKey(key Key) {
	n.keys = append(n.keys, key)
}

// AddValues добавляет значения в узел
func (n *Node) AddValues(values []Value) {
	n.values = append(n.values, values)
}

// AddChild добавляет дочерний узел
func (n *Node) AddChild(child *Node) {
	n.children = append(n.children, child)
}

// SetNext устанавливает next для узла
func (n *Node) SetNext(next *Node) {
	n.next = next
}

// SetParent устанавливает parent для узла
func (n *Node) SetParent(parent *Node) {
	n.parent = parent
}
