package conf

import (
	"path/filepath"
	"strings"

	"github.com/huajiao-tv/gokeeper/model"
)

var (
	ConfSuffix     = []string{".conf"}
	DefaultSection = model.DefaultSection
	PathSeparator  = model.PathSeparator
)

type KeyData struct {
	model.ConfData
	file string
}

func NewKeyData(file string, cfd model.ConfData) KeyData {
	return KeyData{file: file, ConfData: cfd}
}

type SubscribeType int

const (
	SubscribeSection SubscribeType = 1
	SubscribeFile    SubscribeType = 2
	SubscribeDir     SubscribeType = 3
)

type Subscription string

func (s Subscription) Type() SubscribeType {
	if HasConfSuffix(string(s)) {
		return SubscribeFile
	}

	p := strings.Split(string(s), PathSeparator)
	if len(p) >= 2 && HasConfSuffix(p[len(p)-2]) {
		return SubscribeSection
	}

	return SubscribeDir
}

func (s Subscription) File() string {
	switch s.Type() {
	case SubscribeSection:
		p := strings.Split(string(s), PathSeparator)
		return strings.Join(p[:len(p)-1], PathSeparator)
	default:
		return string(s)
	}
}

// input = "/test.conf/section" output = "/test.conf"
// input = "/zy/test.conf/section"  output =  "/test.conf" "/zy/test.conf"
// input - "/center.conf" output = "/center.conf"
// input = "/zy/center.conf" output = "/center.conf" "/zy/center.conf"
// inherited path include current file and parents files
func (s Subscription) InvolvedFilesPath() (files []string) {
	file := s.File()
	if strings.Count(file, PathSeparator) == 0 {
		return
	}
	fs := strings.Split(file, PathSeparator)
	lfs := len(fs)
	fname := fs[lfs-1]
	for i := 0; i < lfs-1; i++ {
		fpath := filepath.Join(strings.Join(fs[:i+1], PathSeparator), fname)
		fpath = PathSeparator + strings.TrimLeft(fpath, PathSeparator)
		files = append(files, fpath)
	}
	return files
}

func (s Subscription) Section() string {
	switch s.Type() {
	case SubscribeSection:
		p := strings.Split(string(s), PathSeparator)
		return p[len(p)-1]
	default:
		return ""
	}
	return ""
}

func HasConfSuffix(file string) bool {
	for _, suffix := range ConfSuffix {
		if strings.HasSuffix(strings.ToLower(file), suffix) {
			return true
		}
	}
	return false
}
