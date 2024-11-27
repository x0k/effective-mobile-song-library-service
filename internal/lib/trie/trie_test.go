package trie_test

import (
	"testing"

	"github.com/x0k/effective-mobile-song-library-service/internal/lib/trie"
)

func TestTrie(t *testing.T) {
	var n *trie.Node[rune, int]
	n = trie.Insert(n, []rune("abc"), 1)
	a := trie.GetNode(n, 'a')
	if a == nil {
		t.Fatal("a should not be nil")
	}

	b := trie.GetNode(a, 'b')
	if b == nil {
		t.Fatal("b should not be nil")
	}

	c := trie.GetNode(b, 'c')
	if c == nil {
		t.Fatal("c should not be nil")
	}
}
