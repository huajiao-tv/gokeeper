package process

import (
	"os"
	"path/filepath"
)

func GetProcessBinaryDir() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir, err
}
