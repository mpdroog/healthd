package config

import (
	"os"
	"path/filepath"
)

type File struct {
	Cmd string
}
type Config struct {
	Files []File
}

var (
	Verbose   bool
	Scriptdir string
	C         Config
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
	if len(base) > 0 {
		return filepath.Walk(base, func(path string, f os.FileInfo, err error) error {
			if path == base {
				// ignore root
				return nil
			}

			C.Files = append(C.Files, File{
				Cmd: path,
			})
			return nil
		})
	}
	return nil
}
