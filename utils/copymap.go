package utils

func CopyMap(from map[string]interface{}) map[string]interface{} {
	node := make(map[string]interface{})
	for k, v := range from {
		node[k] = v
	}
	return node
}

func CopyMapInterface(from interface{}) interface{} {
	switch x := from.(type) {
	case map[interface{}]interface{}:
		node := make(map[string]interface{})
		for k, v := range from.(map[interface{}]interface{}) {
			node[k.(string)] = CopyMapInterface(v)
		}
		return node
	default:
		_ = x
		return from
	}
}
