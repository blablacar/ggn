package work

import (
	"testing"

	"github.com/blablacar/ggn/ggn"
)

func itemInSlice(a string, s []string) bool {
	for _, x := range s {
		if a == x {
			return true
		}
	}
	return false
}

func TestAddEnvIncludeFiles(t *testing.T) {
	ggn.Home.Config.WorkPath = "../examples"
	env := Env{}
	files := []string{"../examples/env/prod-dc1/attributes/test.yml"}
	files, err := env.addIncludeFiles(files)
	if err != nil {
		t.Logf("addIncludeFiles returned an error : %v", err)
		t.Fail()
	}
	for _, includeFile := range []string{
		"../examples/common-attributes/test/includeFile.yml",
		"../examples/env/prod-dc2/common-attributes/test/includeFile.yml",
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
