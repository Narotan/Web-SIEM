package storage

import (
	"nosql_db/internal/index"
)

// IndexFile структура для сохранения индекса
type IndexFile struct {
	Field string           `json:"field"`
	Order int              `json:"order"`
	Nodes []SerializedNode `json:"nodes"`
}

// SerializedNode представляет сериализованный узел b-tree
type SerializedNode struct {
	IsLeaf   bool       `json:"is_leaf"`
	Keys     [][]byte   `json:"keys"`
	Values   [][][]byte `json:"values,omitempty"`
	Children []int      `json:"children,omitempty"`
}

// serializeBTree сериализует b-tree в структуру для json
func serializeBTree(tree *index.BTree, fieldName string, order int) *IndexFile {
	if tree == nil || tree.GetRoot() == nil {
		return &IndexFile{
			Field: fieldName,
			Order: order,
			Nodes: []SerializedNode{},
		}
	}
	var nodes []SerializedNode
	nodeIndexMap := make(map[*index.Node]int)
	root := tree.GetRoot()
	queue := []*index.Node{root}
	idx := 0
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		nodeIndexMap[node] = idx
		idx++
		if !node.GetIsLeaf() {
			children := node.GetChildren()
			for _, child := range children {
				if child != nil {
					queue = append(queue, child)
				}
			}
		}
	}
	queue = []*index.Node{root}
	visited := make(map[*index.Node]bool)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if visited[node] {
			continue
		}
		visited[node] = true
		serialized := SerializedNode{
			IsLeaf: node.GetIsLeaf(),
			Keys:   make([][]byte, 0),
		}
		keys := node.GetKeys()
		for _, key := range keys {
			serialized.Keys = append(serialized.Keys, []byte(key))
		}
		if node.GetIsLeaf() {
			values := node.GetValues()
			serialized.Values = make([][][]byte, len(values))
			for i, vals := range values {
				serialized.Values[i] = make([][]byte, len(vals))
				for j, val := range vals {
					serialized.Values[i][j] = []byte(val)
				}
			}
		} else {
			children := node.GetChildren()
			serialized.Children = make([]int, 0)
			for _, child := range children {
				if child != nil {
					serialized.Children = append(serialized.Children, nodeIndexMap[child])
					queue = append(queue, child)
				}
			}
		}
		nodes = append(nodes, serialized)
	}
	return &IndexFile{
		Field: fieldName,
		Order: order,
		Nodes: nodes,
	}
}

// deserializeBTree восстанавливает b-tree из сериализованных данных
func deserializeBTree(data *IndexFile) *index.BTree {
	if len(data.Nodes) == 0 {
		return index.NewBPlusTree(data.Order)
	}
	nodes := make([]*index.Node, len(data.Nodes))
	for i, sn := range data.Nodes {
		node := index.NewNode(sn.IsLeaf)
		for _, key := range sn.Keys {
			node.AddKey(index.Key(key))
		}
		if sn.IsLeaf && len(sn.Values) > 0 {
			for _, vals := range sn.Values {
				var values []index.Value
				for _, val := range vals {
					values = append(values, index.Value(val))
				}
				node.AddValues(values)
			}
		}
		nodes[i] = node
	}
	for i, sn := range data.Nodes {
		if !sn.IsLeaf && len(sn.Children) > 0 {
			for _, childIdx := range sn.Children {
				if childIdx < len(nodes) {
					nodes[i].AddChild(nodes[childIdx])
					nodes[childIdx].SetParent(nodes[i])
				}
			}
		}
	}
	var prevLeaf *index.Node
	for _, node := range nodes {
		if node.GetIsLeaf() {
			if prevLeaf != nil {
				prevLeaf.SetNext(node)
			}
			prevLeaf = node
		}
	}
	tree := index.NewBPlusTree(data.Order)
	tree.SetRoot(nodes[0])
	return tree
}
