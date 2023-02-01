package consulconfig

import "strings"

type configTree struct {
	separator string
	nodes     map[string]any
}

func newConfigTree() *configTree {
	return &configTree{separator: "/", nodes: map[string]any{}}
}

func newConfigTreeFromMap(m map[string]string) *configTree {
	t := newConfigTree()
	for k, v := range m {
		t.addNode(k, v)
	}
	return t
}

func newSubTree(nodes map[string]any) *configTree {
	return &configTree{separator: "/", nodes: nodes}
}

func (ct *configTree) getSubTree(k string) (*configTree, bool) {
	v, exists := ct.nodes[k]
	if exists {
		if subtree, ok := v.(map[string]any); ok {
			return newSubTree(subtree), ok
		} else {
			return nil, false
		}
	}
	return nil, exists
}

func (ct *configTree) getString(k string) (string, bool) {
	if v, exists := ct.nodes[k]; !exists {
		return "", false
	} else if value, castOk := v.(string); castOk {
		return value, true
	}
	return "", false
}

func (ct *configTree) getInString(ks ...string) (string, bool) {

	var tmpNode = ct.nodes

	for len(ks) > 0 {
		k := ks[0]
		ksNum := len(ks)
		ks = ks[1:]
		v, exists := tmpNode[k]
		if vmap, ok := v.(map[string]any); exists && ok {
			tmpNode = vmap
		} else if exists && ksNum == 1 {
			if value, ok := v.(string); ok {
				return value, ok
			}
		} else {
			break
		}
	}
	return "", false
}

func (ct *configTree) addNode(key, value string) {
	var tmpNode = ct.nodes
	ks := strings.Split(key, ct.separator)
	for len(ks) > 0 {
		k := ks[0]
		ksNum := len(ks)
		ks = ks[1:]

		if v, exists := tmpNode[k]; exists {
			if vmap, ok := v.(map[string]any); ok {
				tmpNode = vmap
			}
		} else if ksNum > 1 {
			m := map[string]any{}
			tmpNode[k] = m
			tmpNode = m
		} else if ksNum == 1 {
			tmpNode[k] = value
		}
	}
}
