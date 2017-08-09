package work

import (
	"os"
	"regexp"
	"testing"
)

func TestAddServiceIncludeFiles(t *testing.T) {
	service := Service{}
	service.env.path = "../examples/env/prod-dc1"
	files := []string{"../examples/env/prod-dc1/services/test-service/attributes/test.yml"}
	files, err := service.addIncludeFiles(files)
	if err != nil {
		t.Logf("addIncludeFiles returned an error : %v", err)
		t.Fail()
	}
	for _, includeFile := range []string{
		"../examples/common-attributes/test/includeFile.yml",
		"../examples/env/prod-dc2/common-attributes/test/includeFile.yml",
		"../examples/env/prod-dc1/common-attributes/test/includeFile.yml",
	} {
		if !itemInSlice(includeFile, files) {
			t.Logf("File not included : %v", includeFile)
			for _, f := range files {
				t.Log(f)
			}
			t.Fail()
		}
	}
}

func TestRenderManifest(t *testing.T) {
	attr := make(map[string]interface{})
	attr["version"] = 123
	service := Service{
		path:               "./testdata/cassandra-tmpl",
		manifestAttributes: attr,
	}

	template, err := service.renderManifest()
	if err != nil {
		t.Fatalf("Unexpected error : %q", err)
	}

	foundVersion, _ := regexp.Match("pod-cassandra:123", template)
	if !foundVersion {
		t.Errorf("Unexpected template rendering : %q", template)
	}
}

func TestReadManifest(t *testing.T) {
	attr := make(map[string]interface{})
	attr["version"] = 123
	service := Service{
		path:               "./testdata/cassandra-tmpl",
		manifestAttributes: attr,
	}

	nowhere := Service{path: "/nowhere"}
	_, err := nowhere.readManifest(false)
	if !os.IsNotExist(err) {
		t.Errorf("Unexpected error on non-existant manifest : %q", err)
	}

	manifest, _ := service.readManifest(false)
	foundTag, _ := regexp.Match("pod-cassandra:{{.version}}", manifest)
	if !foundTag {
		t.Errorf("Unexpected manifest content : %q", manifest)
	}
}

func TestReloadService(t *testing.T) {
	attr := make(map[string]interface{})
	attr["version"] = 123
	service := Service{
		path:               "./testdata/cassandra-tmpl",
		manifestAttributes: attr,
	}
	service.reloadService()
	if service.manifest.Containers[0] != "aci.blbl.cr/pod-cassandra:123" {
		t.Errorf("Unexpected manifest content : %q", service.manifest)
	}

}
