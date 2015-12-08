package utils

import (
	"errors"
	"fmt"
	"strconv"
)

func CopyMap(from map[string]interface{}) map[string]interface{} {
	node := make(map[string]interface{})
	for k, v := range from {
		node[k] = v
	}
	return node
}

//func CopyMapInterface(from interface{}) interface{} {
//	switch x := from.(type) {
//	case map[interface{}]interface{}:
//		node := make(map[string]interface{})
//		for k, v := range from.(map[interface{}]interface{}) {
//			node[k.(string)] = CopyMapInterface(v)
//		}
//		return node
//	default:
//		_ = x
//		return from
//	}
//}

// transform YAML to JSON
func TransformYamlToJson(in interface{}) (_ interface{}, err error) {
	switch in.(type) {
	case map[interface{}]interface{}:
		o := make(map[string]interface{})
		for k, v := range in.(map[interface{}]interface{}) {
			sk := ""
			switch k.(type) {
			case string:
				sk = k.(string)
			case int:
				sk = strconv.Itoa(k.(int))
			default:
				return nil, errors.New(
					fmt.Sprintf("type not match: expect map key string or int get: %T", k))
			}
			v, err = TransformYamlToJson(v)
			if err != nil {
				return nil, err
			}
			o[sk] = v
		}
		return o, nil
	case []interface{}:
		in1 := in.([]interface{})
		len1 := len(in1)
		o := make([]interface{}, len1)
		for i := 0; i < len1; i++ {
			o[i], err = TransformYamlToJson(in1[i])
			if err != nil {
				return nil, err
			}
		}
		return o, nil
	default:
		return in, nil
	}
	return in, nil
}
