package service

import (
	"bytes"
	"encoding/json"
	"github.com/blablacar/ggn/utils"
	"github.com/peterbourgon/mergemap"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func (u Unit) Generate(tmpl *utils.Templating) {
	u.Log.Debug("Generate")
	data := u.GenerateAttributes()
	out, err := json.Marshal(data)
	if err != nil {
		u.Log.WithError(err).Panic("Cannot marshall attributes")
	}
	data["attributes"] = strings.Replace(string(out), "\\\"", "\\\\\\\"", -1)

	u.prepareEnvironmentAttributes(data)

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
	data := utils.CopyMap(u.Service.GetAttributes())
	data["acis"] = u.Service.PrepareAciList(nil)
	data = mergemap.Merge(data, u.Service.NodeAttributes(u.Name))
	if data["attributes"] != nil {
		data = mergemap.Merge(data, data["attributes"].(map[string]interface{})) // TODO this should be remove in the future
	}
	return data
}

func (u Unit) prepareEnvironmentAttributes(data map[string]interface{}) {
	var envAttr bytes.Buffer
	var envAttrVars bytes.Buffer
	num := 0
	for i, c := range data["attributes"].(string) {
		if i%1950 == 0 {
			if num != 0 {
				envAttr.WriteString("'\n")
			}
			envAttr.WriteString("Environment='ATTR_")
			envAttr.WriteString(strconv.Itoa(num))
			envAttr.WriteString("=")
			envAttrVars.WriteString("${ATTR_")
			envAttrVars.WriteString(strconv.Itoa(num))
			envAttrVars.WriteString("}")
			num++
		}
		envAttr.WriteRune(c)
	}
	envAttr.WriteString("'\n")

	data["environmentAttributes"] = envAttr.String()
	data["environmentAttributesVars"] = envAttrVars.String()
}
