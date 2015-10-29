package utils

func CopyMap(from map[string]interface{}) map[string]interface{} {
	node := make(map[string]interface{})
	for k, v := range from {
		node[k] = v
	}
	return node
}
