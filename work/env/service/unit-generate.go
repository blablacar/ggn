package service

import (
	"bytes"
	"encoding/json"
	"github.com/blablacar/ggn/utils"
	"github.com/peterbourgon/mergemap"
	"io/ioutil"
	"os"
	"strings"
)

func (u Unit) Generate(tmpl *utils.Templating) {
	u.Log.Debug("Generate")
	data := u.GenerateAttributes()

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

func (u Unit) GenerateAttributes() map[string]interface{} {
	data := make(map[string]interface{})
	data["node"] = u.Service.NodeAttributes(u.Name)
	data["node"].(map[string]interface{})["acis"] = u.Service.PrepareAciList(nil)
	data["attribute"] = utils.CopyMap(u.Service.GetAttributes())
	data["attribute"].(map[string]interface{})["hostname"] = data["node"].(map[string]interface{})["hostname"]
	if data["node"].(map[string]interface{})["attributes"] != nil {
		source := utils.CopyMapInterface(data["node"].(map[string]interface{})["attributes"].(map[interface{}]interface{}))
		data["attribute"] = mergemap.Merge(data["attribute"].(map[string]interface{}), source.(map[string]interface{}))
	}

	return data
}
