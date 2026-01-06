package index

import "bytes"

type BTree struct {
	root  *Node
	order int
}

// NewBPlusTree создаёт новый b+ tree с указанным order
func NewBPlusTree(order int) *BTree {
	return &BTree{
		root: &Node{
			isLeaf: true,
			keys:   []Key{},
			values: [][]Value{},
		},
		order: order,
	}
}

// Insert вставляет ключ и значение в дерево
func (tree *BTree) Insert(key Key, value Value) {

	if tree.root == nil {
		NewBPlusTree(tree.order).Insert(key, value)
	}

	// поиск листа в дереве для вставки ключа
	leaf := tree.findLeaf(tree.root, key)

	// вставляем ключ и значение в лист
	tree.insertInLeaf(leaf, key, value)

	// если лист переполнен, разделим его
	if len(leaf.keys) > tree.order*2-1 {
		tree.splitLeaf(leaf)
	}
}

// findLeaf возвращает лист для заданного ключа
func (tree *BTree) findLeaf(node *Node, key Key) *Node {
	if node.isLeaf {
		return node
	}

	// смотрим куда идем: вправо или влево по разделению
	for i, k := range node.keys {
		if bytes.Compare(key, k) < 0 {
			return tree.findLeaf(node.children[i], key)
		}
	}

	return tree.findLeaf(node.children[len(node.children)-1], key)
}

// insertInLeaf вставляет ключ и значение в лист
func (tree *BTree) insertInLeaf(leaf *Node, key Key, value Value) {
	pos := 0
	for pos < len(leaf.keys) && bytes.Compare(leaf.keys[pos], key) < 0 {
		pos++
	}

	// если ключ уже есть, добавляем значение в массив
	if pos < len(leaf.keys) && bytes.Equal(leaf.keys[pos], key) {
		leaf.values[pos] = append(leaf.values[pos], value)
		return
	}

	leaf.keys = append(leaf.keys, nil)
	copy(leaf.keys[pos+1:], leaf.keys[pos:])
	leaf.keys[pos] = key

	leaf.values = append(leaf.values, nil)
	copy(leaf.values[pos+1:], leaf.values[pos:])
	leaf.values[pos] = []Value{value}
}

// splitLeaf разделяет лист и добавляет новый лист в связный список
func (tree *BTree) splitLeaf(leaf *Node) {
	mid := len(leaf.keys) / 2

	// создаем новый лист где пойдет вторая половина ключей и значений
	newLeaf := &Node{
		isLeaf: true,
		keys:   append([]Key{}, leaf.keys[mid:]...),
		values: append([][]Value{}, leaf.values[mid:]...),
		next:   leaf.next,
		parent: leaf.parent,
	}

	leaf.keys = leaf.keys[:mid]
	leaf.values = leaf.values[:mid]
	leaf.next = newLeaf

	// после того как мы сплитнули лист, нужно поднимать ключ в родителя
	// потом он должен знать о новом листе
	tree.insertInParent(leaf, newLeaf.keys[0], newLeaf)
}

// insertInParent поднимает ключ в родителя после сплита
// надо чтобы родитель знал когда идти в правое, а когда в левое поддерево
func (tree *BTree) insertInParent(left *Node, key Key, right *Node) {
	if left.parent == nil {
		newRoot := &Node{
			isLeaf:   false,
			keys:     []Key{key},
			children: []*Node{left, right},
		}
		left.parent = newRoot
		right.parent = newRoot
		tree.root = newRoot
		return
	}

	parent := left.parent

	pos := 0
	for pos < len(parent.keys) && bytes.Compare(parent.keys[pos], key) < 0 {
		pos++
	}

	parent.keys = append(parent.keys, nil)
	copy(parent.keys[pos+1:], parent.keys[pos:])
	parent.keys[pos] = key

	parent.children = append(parent.children, nil)
	copy(parent.children[pos+2:], parent.children[pos+1:])
	parent.children[pos+1] = right
	right.parent = parent

	// если родитель переполнен, то сплитим родителя
	if len(parent.keys) > tree.order*2-1 {
		tree.splitInternal(parent)
	}
}

// splitInternal разделяет внутренний узел и поднимает ключ в родителя
func (tree *BTree) splitInternal(node *Node) {
	mid := len(node.keys) / 2
	keyToPushUp := node.keys[mid]

	newNode := &Node{
		isLeaf:   false,
		keys:     append([]Key{}, node.keys[mid+1:]...),
		children: append([]*Node{}, node.children[mid+1:]...),
		parent:   node.parent,
	}

	for _, child := range newNode.children {
		child.parent = newNode
	}

	node.keys = node.keys[:mid]
	node.children = node.children[:mid+1]

	// поднимаем ключ в родителя
	tree.insertInParent(node, keyToPushUp, newNode)
}

// Delete удаляет значение из дерева по ключу
func (tree *BTree) Delete(key Key, value Value) bool {
	if tree.root == nil {
		return false
	}

	leaf := tree.findLeaf(tree.root, key)
	return tree.deleteFromLeaf(leaf, key, value)
}

// deleteFromLeaf удаляет конкретное значение из листа
func (tree *BTree) deleteFromLeaf(leaf *Node, key Key, value Value) bool {
	// Находим позицию ключа
	pos := -1
	for i, k := range leaf.keys {
		if bytes.Equal(k, key) {
			pos = i
			break
		}
	}

	if pos == -1 {
		return false // ключ не найден
	}

	// Удаляем конкретное значение из массива значений
	valuePos := -1
	for i, v := range leaf.values[pos] {
		if bytes.Equal(v, value) {
			valuePos = i
			break
		}
	}

	if valuePos == -1 {
		return false // значение не найдено
	}

	// Удаляем значение
	leaf.values[pos] = append(leaf.values[pos][:valuePos], leaf.values[pos][valuePos+1:]...)

	// Если значений больше нет для этого ключа, удаляем ключ
	if len(leaf.values[pos]) == 0 {
		leaf.keys = append(leaf.keys[:pos], leaf.keys[pos+1:]...)
		leaf.values = append(leaf.values[:pos], leaf.values[pos+1:]...)
	}

	return true
}

// GetRoot возвращает корень дерева
func (tree *BTree) GetRoot() *Node {
	return tree.root
}

// SetRoot устанавливает корень дерева
func (tree *BTree) SetRoot(node *Node) {
	tree.root = node
}

// GetOrder возвращает порядок дерева
func (tree *BTree) GetOrder() int {
	return tree.order
}
