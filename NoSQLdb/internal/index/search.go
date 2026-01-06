package index

import "bytes"

// Search выполняет точечный поиск по ключу ($eq)
func (tree *BTree) Search(key Key) []Value {
	if tree.root == nil {
		return nil
	}

	leaf := tree.findLeaf(tree.root, key)

	// ищем ключ в листе
	for i, k := range leaf.keys {
		if bytes.Equal(k, key) {
			return leaf.values[i]
		}
	}

	return nil
}

// RangeSearch выполняет диапазонный поиск ($gt, $lt, $gte, $lte)
func (tree *BTree) RangeSearch(start, end Key, includeStart, includeEnd bool) []Value {
	if tree.root == nil {
		return nil
	}

	var result []Value

	var startLeaf *Node
	if start == nil {
		startLeaf = tree.findLeftmostLeaf(tree.root)
	} else {
		startLeaf = tree.findLeaf(tree.root, start)
	}

	// проходим по всем листьям через связанный список
	for leaf := startLeaf; leaf != nil; leaf = leaf.next {
		for i, k := range leaf.keys {
			if start != nil {
				cmp := bytes.Compare(k, start)
				if cmp < 0 || (cmp == 0 && !includeStart) {
					continue
				}
			}

			if end != nil {
				cmp := bytes.Compare(k, end)
				if cmp > 0 || (cmp == 0 && !includeEnd) {
					return result
				}
			}

			// добавляем все значения для этого ключа
			result = append(result, leaf.values[i]...)
		}

		// если достигли конца диапазона, выходим
		if end != nil && len(leaf.keys) > 0 {
			lastKey := leaf.keys[len(leaf.keys)-1]
			if bytes.Compare(lastKey, end) >= 0 {
				break
			}
		}
	}

	return result
}

// SearchGreaterThan ищет все значения где ключ > key ($gt)
func (tree *BTree) SearchGreaterThan(key Key) []Value {
	return tree.RangeSearch(key, nil, false, false)
}

// SearchLessThan ищет все значения где ключ < key ($lt)
func (tree *BTree) SearchLessThan(key Key) []Value {
	return tree.RangeSearch(nil, key, false, false)
}

// SearchGreaterThanOrEqual ищет все значения где ключ >= key ($gte)
func (tree *BTree) SearchGreaterThanOrEqual(key Key) []Value {
	return tree.RangeSearch(key, nil, true, false)
}

// SearchLessThanOrEqual ищет все значения где ключ <= key ($lte)
func (tree *BTree) SearchLessThanOrEqual(key Key) []Value {
	return tree.RangeSearch(nil, key, false, true)
}

// SearchIn выполняет множественный точечный поиск ($in)
// возвращает все значения для списка ключей
func (tree *BTree) SearchIn(keys []Key) []Value {
	var result []Value

	for _, key := range keys {
		values := tree.Search(key)
		if values != nil {
			result = append(result, values...)
		}
	}

	return result
}

// findLeftmostLeaf находит самый левый лист дерева
func (tree *BTree) findLeftmostLeaf(node *Node) *Node {
	if node == nil {
		return nil
	}

	for !node.isLeaf {
		if len(node.children) > 0 {
			node = node.children[0]
		} else {
			break
		}
	}

	return node
}

// GetAllValues возвращает все значения из дерева (для full scan)
func (tree *BTree) GetAllValues() []Value {
	if tree.root == nil {
		return nil
	}

	var result []Value
	leaf := tree.findLeftmostLeaf(tree.root)

	for leaf != nil {
		for _, values := range leaf.values {
			result = append(result, values...)
		}
		leaf = leaf.next
	}

	return result
}
