package work

import (
	"io/ioutil"
	"os"

	"github.com/blablacar/ggn/ggn"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/logs"
)

const PATH_ENV = "/env"

type Work struct {
	path   string
	fields data.Fields
}

func NewWork(path string) *Work {
	return &Work{
		path:   path,
		fields: data.WithField("path", path),
	}
}

func (w Work) LoadEnv(name string) *Env {
	return NewEnvironment(w.path+PATH_ENV, name)
}

func (w Work) ListEnvs() []string {
	path := ggn.Home.Config.WorkPath + PATH_ENV
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logs.WithEF(err, w.fields).WithField("path", path).Fatal("env directory not found")
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		logs.WithEF(err, w.fields).WithField("path", path).Fatal("Cannot read env directory")
	}

	var envs []string
	for _, file := range files {
		if file.Mode()&os.ModeSymlink == os.ModeSymlink {
			followed_file, err := os.Readlink(path + "/" + file.Name())
			if err != nil {
				continue
			}
			if followed_file[0] != '/' {
				followed_file = path + "/" + followed_file
			}
			file, err = os.Lstat(followed_file)
			if err != nil {
				continue
			}
			logs.WithField("followed_link", file.Name()).Trace("Followed Link")
		}

		if !file.IsDir() {
			continue
		}
		envs = append(envs, file.Name())
	}
	return envs
}
