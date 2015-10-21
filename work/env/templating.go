package env

import (
	"bufio"
	"encoding/json"
	"github.com/kelseyhightower/memkv"
	"io"
	"os"
	"path"
	"strings"
	"text/template"
	"time"
)

type Templating struct {
	template  *template.Template
	name      string
	content   string
	functions map[string]interface{}
	vars      memkv.Store
}

func NewTemplating(name, content string) *Templating {
	t := new(Templating)
	t.name = name
	t.content = cleanupOfTemplate(content)
	t.functions = newFuncMap()
	t.vars = memkv.New()
	addFuncs(t.functions, t.vars.FuncMap)
	return t
}

func cleanupOfTemplate(content string) string {
	var lines []string
	var currentLine string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		part := strings.TrimRight(scanner.Text(), " ")
		leftTrim := strings.TrimLeft(part, " ")
		if strings.HasPrefix(leftTrim, "{{-") {
			part = "{{" + leftTrim[3:]
		}
		currentLine += part
		if strings.HasSuffix(currentLine, "-}}") {
			currentLine = currentLine[0:len(currentLine)-3] + "}}"
		} else {
			lines = append(lines, currentLine)
			currentLine = ""
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return strings.Join(lines, "\n")
}

func (t *Templating) Parse() error {
	tmpl, err := template.New(t.name).Funcs(t.functions).Parse(t.content)
	t.template = tmpl
	return err
}

func (t *Templating) Execute(wr io.Writer, data interface{}) error {
	return t.template.Execute(wr, data)
}

func (t *Templating) AddFunctions(fs map[string]interface{}) {
	addFuncs(t.functions, fs)
}

func (t *Templating) AddVar(key string, value string) {
	t.vars.Set(key, value)
}

///////////////////////////////////

func newFuncMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["base"] = path.Base
	m["split"] = strings.Split
	m["json"] = UnmarshalJsonObject
	m["jsonArray"] = UnmarshalJsonArray
	m["dir"] = path.Dir
	m["getenv"] = os.Getenv
	m["join"] = strings.Join
	m["datetime"] = time.Now
	m["toUpper"] = strings.ToUpper
	m["toLower"] = strings.ToLower
	m["contains"] = strings.Contains
	m["replace"] = strings.Replace
	return m
}

func addFuncs(out, in map[string]interface{}) {
	for name, fn := range in {
		out[name] = fn
	}
}

func UnmarshalJsonObject(data string) (map[string]interface{}, error) {
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func UnmarshalJsonArray(data string) ([]interface{}, error) {
	var ret []interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}
