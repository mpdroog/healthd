package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	return ReloadConf()
}
func Close() error {
	return nil
}

func ReloadConf() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	C.Files = make(map[string]File)
	return loadConfDir(ctx, Scriptdir)
}

func loadConfDir(ctx context.Context, base string) error {
	if len(base) == 0 {
		return fmt.Errorf("no path given")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		return fmt.Errorf("Context has no deadline")
	}
	C.Files = make(map[string]File)
	Departments = make(map[string]struct{})
	Departments["default"] = struct{}{}

	return filepath.Walk(base, func(path string, f os.FileInfo, err error) error {
		if time.Now().After(deadline) {
			return fmt.Errorf("Context deadline reached, deadline=%s", deadline)
		}
		if path == base {
			// ignore root
			return nil
		}
		if strings.HasSuffix(path, ".json") {
			r, e := os.OpenFile(path, os.O_RDONLY, 0)
			if e != nil {
				return e
			}
			if e := r.SetDeadline(deadline); e != nil {
				return e
			}
			// DevNote: be careful to close r!
			obj := Meta{}
			if e := json.NewDecoder(r).Decode(&obj); e != nil {
				r.Close() // ignore
				return e
			}
			if e := r.Close(); e != nil {
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
