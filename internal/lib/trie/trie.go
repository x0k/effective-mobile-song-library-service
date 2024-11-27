package trie

type Node[T comparable, V any] struct {
	values map[T]*Node[T, V]
	Value  V
}

func Insert[T comparable, V any](node *Node[T, V], key []T, value V) *Node[T, V] {
	if node == nil {
		node = &Node[T, V]{
			values: map[T]*Node[T, V]{},
		}
	}
	if len(key) == 0 {
		node.Value = value
		return node
	}
	next, ok := node.values[key[0]]
	if !ok {
		next = &Node[T, V]{
			values: map[T]*Node[T, V]{},
		}
	}
	node.values[key[0]] = Insert(next, key[1:], value)
	return node
}

func GetNode[T comparable, V any](node *Node[T, V], key T) *Node[T, V] {
	if node == nil {
		return nil
	}
	return node.values[key]
}
