package env

import (
	"bytes"
	"github.com/blablacar/attributes-merger/attributes"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/green-garden/spec"
	"github.com/blablacar/green-garden/work/env/service"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Service struct {
	path     string
	name     string
	manifest spec.ServiceManifest
}

func NewService(path string, name string) *Service {
	service := new(Service)
	service.path = path + "/" + name
	service.name = name
	service.loadManifest()
	return service
}

//////////////////////////////////////

func (s Service) LoadUnit(name string) *service.Unit {
	unit := service.NewUnit(s.path+"/units", name)
	return unit
}

func (s Service) GenerateUnits(envAttributePath string, envName string) {

	tmpl := s.loadUnitTemplate()
	for i, node := range s.manifest.Nodes {
		if node[spec.NODE_HOSTNAME].(string) == "" {
			log.Get().Panic("hostname is mandatory in node informations : " + s.manifestPath() + "node[" + string(i) + "]")
		}
		log.Get().Debug("Processing node : " + s.name + ":" + node[spec.NODE_HOSTNAME].(string))

		//		log.Get().Warn()
		//		for k := range node {
		//			log.Get().Error(k)
		//		}

		unitName := envName + "_" + s.name + "_" + node[spec.NODE_HOSTNAME].(string) + ".service"
		log.Get().Trace("Unit name is :" + unitName)
		template := NewTemplating(node[spec.NODE_HOSTNAME].(string), string(tmpl))
		//		data := string(attributes.Merge("CONFD_DATA", ))
		//		log.Get().Debug("env data :" + data)
		//		template.AddVar("/data", data)
		template.Parse()

		var b bytes.Buffer

		attributes := attributes.MergeAttributes(mergeAttributesDirectories(envAttributePath, s.path+spec.PATH_ATTRIBUTES))
		//		hits := make(map[string]interface{})
		attributes["node"] = node

		//		type yopla struct {
		//			Node interface{}
		//		}
		//
		//		ss := yopla{Node: node}

		err := template.Execute(&b, attributes)
		if err != nil {
			log.Get().Panic("Failed to run templating for unit "+unitName, err)
		}
		ioutil.WriteFile(s.path+"/units"+"/"+unitName, b.Bytes(), 0644)

		//		export CONFD_DATA=$(cat attributes.json)
		//		fi
		//		${BASEDIR}/confd -onetime -config-file=${CNT_PATH}/prestart/confd.toml
		//		m := node.(map[string]interface{})
		//		for k, v := range m {
		//			if k == "Hostname" {
		//				switch vv := v.(type) {
		//				case string:
		//					log.Get().Error("HOSTNAME FOUND" + vv)
		//					//								fmt.Println(k, "is string", vv)
		//					//				case int:
		//					//				fmt.Println(k, "is int", vv)
		//					//				case []interface{}:
		//					//				fmt.Println(k, "is an array:")
		//					//				for i, u := range vv {
		//					//					fmt.Println(i, u)
		//					//				}
		//					//				default:
		//					//				fmt.Println(k, "is of a type I don't know how to handle")
		//				}
		//			}
		//		}

		//		envs := attributes.Merge("CONFD_DATA", mergeAttributesDirectories(envAttributePath, s.path+spec.PATH_ATTRIBUTES))
		//		log.Get().Info(string(envs))
		//		os.Setenv("CONFD_DATA", string(envs))
		//
		//		templates := make([]*confd.TemplateResource, 1)
		//		templates[0].Dest
		//		templates[0].Src
		//
		//		confd.ProcessTemplates()

		//		confd.Template(tmpl, envs, s.path)

		//		type Inventory struct {
		//			Material string
		//			Count    uint
		//		}
		//		sweaters := Inventory{"wool", 17}
		//		tmpl, err := template.New(path.Base(t.Src)).Funcs(t.funcMap).ParseFiles(t.Src)
		//
		//		err := tmpl.Execute(os.Stdout, sweaters)
		//		if err != nil {
		//			panic(err)
		//		}
	}
}

///////////////////////////////////////

func mergeAttributesDirectories(envAttributesPath string, serviceAttributesPath string) []string {
	res := []string{}

	{
		in := attributes.NewInputs(envAttributesPath)
		// initialize input files list
		err := in.ListFiles()
		if err != nil {
			panic(err)
		}

		for _, file := range in.Files {
			res = append(res, in.Directory+file)
		}
	}
	{
		in := attributes.NewInputs(serviceAttributesPath)
		// initialize input files list
		err := in.ListFiles()
		if err != nil {
			panic(err)
		}

		for _, file := range in.Files {
			res = append(res, in.Directory+file)
		}
	}
	return res
}

func (s *Service) loadUnitTemplate() []byte /* *template.Template */ {
	path := s.path + spec.PATH_UNIT_TEMPLATE
	source, err := ioutil.ReadFile(path)
	if err != nil {
		log.Get().Warn("Cannot read unit template file '" + s.name + "' : " + path)
	}
	//	tmpl, err := template.New("unit").Parse(string(source))
	//	if err != nil {
	//		log.Get().Panic("Cannot parse unit template", err)
	//	}
	return source
}

func (s Service) manifestPath() string {
	return s.path + spec.PATH_SERVICE_MANIFEST
}

func (s *Service) loadManifest() {
	manifest := spec.ServiceManifest{}
	path := s.manifestPath()
	log.Get().Trace("Service manifest is at : " + path)
	source, err := ioutil.ReadFile(path)
	if err != nil {
		log.Get().Warn("Cannot find manifest for service '" + s.name + "' : " + path)
	}
	err = yaml.Unmarshal([]byte(source), &manifest)
	if err != nil {
		log.Get().Panic(err)
	}
	s.manifest = manifest
}
