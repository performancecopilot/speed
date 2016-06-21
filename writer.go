package speed

import (
	"errors"
	"io"
	"os"
	"path"
	"strings"
)

// Writer defines the interface of a MMV file writer's properties
type Writer interface {
	io.Writer
	Registry() Registry // a writer must contain a registry of metrics and instance domains
	Start() error       // writes an mmv file
}

func mmvFileLocation(name string) (string, error) {
	if strings.ContainsRune(name, os.PathSeparator) {
		return "", errors.New("name cannot have path separator")
	}

	tdir, present := Config["PCP_TMP_DIR"]
	var loc string
	if present {
		loc = path.Join(RootPath, tdir)
	} else {
		loc = os.TempDir()
	}

	return path.Join(loc, "mmv", name), nil
}
