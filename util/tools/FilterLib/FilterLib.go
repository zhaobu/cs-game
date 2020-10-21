//敏感字过滤
package FilterLib

import "unicode/utf8"

type Trie struct {
	Root *TrieNode
}

type TrieNode struct {
	Children map[rune]*TrieNode
	End      bool
}

func NewFilterLib(_flib []string)*Trie   {
	t := NewTrie()
	for _,v:= range _flib {
		t.Inster(v)
	}
	return t
}

func NewTrie() *Trie {
	var r Trie
	r.Root = NewTrieNode()
	return &r
}

func NewTrieNode() *TrieNode {
	n := new(TrieNode)
	n.Children = make(map[rune]*TrieNode)
	return n
}

func (this *Trie) Inster(txt string) {
	if len(txt) < 1 {
		return
	}
	node := this.Root
	key := []rune(txt)
	for i := 0; i < len(key); i++ {
		if _, exists := node.Children[key[i]]; !exists {
			node.Children[key[i]] = NewTrieNode()
		}
		node = node.Children[key[i]]
	}

	node.End = true
}

func (this *Trie) Replace(txt string) string {
	if len(txt) < 1 {
		return txt
	}
	node := this.Root
	key := []rune(txt)
	var chars []rune = nil
	slen := len(key)
	for i := 0; i < slen; i++ {
		if _, exists := node.Children[key[i]]; exists {
			node = node.Children[key[i]]
			for j := i + 1; j < slen; j++ {
				if _, exists := node.Children[key[j]]; exists {
					node = node.Children[key[j]]
					if node.End == true {
						if chars == nil {
							chars = key
						}
						for t := i; t <= j; t++ {
							c, _ := utf8.DecodeRuneInString("*")
							chars[t] = c
						}
						i = j
						node = this.Root
						break
					}
				}
			}
			node = this.Root
		}
	}
	if chars == nil {
		return txt
	} else {
		return string(chars)
	}
}

//校验是否包含有敏感字符 true 有 false 没有
func (this *Trie) CheckFilter(txt string) bool {
	if len(txt) < 1 {
		return false
	}
	node := this.Root
	key := []rune(txt)
	slen := len(key)
	for i := 0; i < slen; i++ {
		if _, exists := node.Children[key[i]]; exists {
			node = node.Children[key[i]]
			for j := i + 1; j < slen; j++ {
				if _, exists := node.Children[key[j]]; exists {
					node = node.Children[key[j]]
					if node.End == true {
						return true
					}
				}
			}
			node = this.Root
		}
	}
	return false
}