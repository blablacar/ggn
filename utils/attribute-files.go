package utils

import (
	"github.com/blablacar/attributes-merger/attributes"
	"k8s.io/kubernetes/Godeps/_workspace/src/github.com/google/cadvisor/utils"
)

func AttributeFiles(path string) ([]string, error) {
	res := []string{}
	if !utils.FileExists(path) {
		return res, nil
	}

	in := attributes.NewInputs(path)
	// initialize input files list
	err := in.ListFiles()
	if err != nil {
		return nil, err
	}

	for _, file := range in.Files {
		res = append(res, in.Directory+file)
	}
	return res, nil
}
