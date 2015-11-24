package service

import (
	"bytes"
	"encoding/json"
	"github.com/blablacar/green-garden/utils"
	"github.com/peterbourgon/mergemap"
	"io/ioutil"
	"os"
	"strings"
)

func (u Unit) Generate(node map[string]interface{}, tmpl *utils.Templating, acis string, attributes map[string]interface{}) {
	u.Log.Debug("Generate")

	data := make(map[string]interface{})

	data["node"] = node
	data["node"].(map[string]interface{})["acis"] = acis

	data["attribute"] = utils.CopyMap(attributes)
	if data["node"].(map[string]interface{})["attributes"] != nil {
		source := utils.CopyMapInterface(data["node"].(map[string]interface{})["attributes"].(map[interface{}]interface{}))
		data["attribute"] = mergemap.Merge(data["attribute"].(map[string]interface{}), source.(map[string]interface{}))
	}

	out, err := json.Marshal(data["attribute"])
	if err != nil {
		u.Log.WithError(err).Panic("Cannot marshall attributes")
	}
	data["attributes"] = strings.Replace(string(out), "\\\"", "\\\\\\\"", -1)

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		u.Log.Error("Failed to run templating", err)
	}
	ok, err := utils.Exists(u.path)
	if !ok || err != nil {
		os.Mkdir(u.path, 0755)
	}
	err = ioutil.WriteFile(u.path+"/"+u.Filename, b.Bytes(), 0644)
	if err != nil {
		u.Log.WithError(err).WithField("path", u.path+"/"+u.Filename).Error("Cannot writer unit")
	}
}
