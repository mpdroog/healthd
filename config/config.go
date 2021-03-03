package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Department string
}
type Meta struct {
	Department string
}
type Config struct {
	Files map[string]File
}

var (
	Verbose     bool
	Scriptdir   string
	C           Config
	Departments map[string]struct{}
)

func Init() error {
	return loadConfDir(Scriptdir)
}
func Close() error {
	return nil
}

func ReloadConf() error {
	C.Files = []File{}
	return loadConfDir(Scriptdir)
}

func loadConfDir(base string) error {
	if len(base) == 0 {
		return fmt.Errorf("no path given")
	}

	C.Files = make(map[string]File)
	Departments = make(map[string]struct{})
	Departments["default"] = struct{}{}

	return filepath.Walk(base, func(path string, f os.FileInfo, err error) error {
		if path == base {
			// ignore root
			return nil
		}
		if strings.HasSuffix(path, ".json") {
			txt, e := ioutil.ReadFile(path)
			if e != nil {
				return e
			}
			obj := Meta{}
			if e := json.Unmarshal(txt, &obj); e != nil {
				return e
			}
			f := C.Files[path]
			f.Department = obj.Department
			C.Files[path[0:len(path)-len(".json")]] = f
			Departments[obj.Department] = struct{}{}
			return nil
		}

		C.Files[path] = File{Department: "default"}
		return nil
	})
}
