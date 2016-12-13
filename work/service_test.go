package work

import "testing"

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
