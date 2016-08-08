package config

import (
	"os"
	"strings"
	"path/filepath"
	"github.com/BurntSushi/toml"
	"fmt"
)

type File struct {
	Cmd string
	Args string
	Errprefix string
}
type Config struct {
	Files map[string]File
}

var (
	Verbose bool
	Confdir string
	C Config
)

func Init() error {
	C.Files = make(map[string]File)
	return loadConfDir(Confdir)
}
func Close() error {
	return nil
}

func loadConfDir(base string) error {
	if len(base) > 0 {
		return filepath.Walk(base, func(path string, f os.FileInfo, err error) error {
			if path == base {
				// ignore root
				return nil
			}

			if strings.HasSuffix(path, ".toml") {
				r, e := os.Open(path)
				if e != nil {
					return e
				}
				var f File
				if _, e := toml.DecodeReader(r, &f); e != nil {
					r.Close()
					return fmt.Errorf("TOML(%s): %s", path, e)
				}
				r.Close()
				C.Files[path] = f
			}

			return nil
		})
	}
	return nil
}