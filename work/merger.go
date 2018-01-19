package work

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/peterbourgon/mergemap"
	"gopkg.in/yaml.v2"
)

func MergeAttributesFilesForMap(omap map[string]interface{}, files []string) (map[string]interface{}, error) {

	newMap := make(map[string]interface{})
	newMap["default"] = omap

	// loop over attributes files
	// merge override files to default files
	for _, file := range files {
		var contentData interface{}
		yml, err := ioutil.ReadFile(file)
		if err != nil {
			return newMap, errs.WithEF(err, data.WithField("file", file), "Failed to read file")
		}
		// yaml to contentData
		err = yaml.Unmarshal(yml, &contentData)
		if err != nil {
			return newMap, errs.WithEF(err, data.WithField("file", file), "Failed to unmarshal file")
		}
		contentData, err = transformYampToJson(contentData)
		if err != nil {
			return newMap, errs.WithEF(err, data.WithField("file", file), "Failed to transform file to json")
		}
		// contentData to map
		json := contentData.(map[string]interface{})
		omap = mergemap.Merge(newMap, json)
	}
	return ProcessOverride(newMap), nil
}

func ProcessOverride(omap map[string]interface{}) map[string]interface{} {
	// merge override to default inside the file
	_, okd := omap["default"]
	if okd == false {
		omap["default"] = make(map[string]interface{}) //init if default doesn't exist
	}
	_, oko := omap["override"]
	if oko == true {
		omap = mergemap.Merge(omap["default"].(map[string]interface{}), omap["override"].(map[string]interface{}))
	} else {
		omap = omap["default"].(map[string]interface{})
	}
	return omap
}

// transformYampToJson YAML to JSON
func transformYampToJson(in interface{}) (_ interface{}, err error) {
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
			v, err = transformYampToJson(v)
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
			o[i], err = transformYampToJson(in1[i])
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
