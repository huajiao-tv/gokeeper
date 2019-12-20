package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/storage"
	"github.com/huajiao-tv/gokeeper/server/storage/operate"
	"github.com/huajiao-tv/gokeeper/utility/go-ini/ini"
)

type Section struct {
	Name string                    `json:"name"`
	Keys map[string]model.ConfData `json:"keys"`
}

func (s *Section) GetName() string {
	return s.Name
}

func (s *Section) GetKeys() map[string]model.ConfData {
	return s.Keys
}

func (s *Section) Set(key string, cfd model.ConfData) {
	s.Keys[key] = cfd
}

func (s *Section) DeleteKey(key string) {
	delete(s.Keys, key)
}

type File struct {
	Name     string     `json:"name"`
	Sections []*Section `json:"sections"`

	// only for gokeeper-cli
	path   string
	info   os.FileInfo
	config *ini.File
	mu     sync.RWMutex
}

func (f *File) GetPath() string {
	return f.path
}

func (f *File) GetInfo() os.FileInfo {
	return f.info
}

func (f *File) GetSections() []*Section {
	f.mu.RLock()
	sections := f.Sections
	f.mu.RUnlock()
	return sections
}

func (f *File) GetKeyData(section, key string) (model.ConfData, error) {
	var cfd model.ConfData
	s, err := f.GetSection(section)
	if err != nil {
		return cfd, err
	}
	for _, cfd := range s.GetKeys() {
		if cfd.Key == key {
			return cfd, nil
		}
	}
	return cfd, fmt.Errorf("key %s not found", key)
}

func (f *File) GetSection(section string) (Section, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, s := range f.Sections {
		if s.GetName() == section {
			return *s, nil
		}
	}
	return Section{}, fmt.Errorf("section %s not found", section)
}

func (f *File) KeyList() []model.ConfData {
	f.mu.RLock()
	defer f.mu.RUnlock()
	keys := []model.ConfData{}
	for _, section := range f.GetSections() {
		for _, cfd := range section.GetKeys() {
			keys = append(keys, cfd)
		}
	}
	return keys
}

func (f *File) initSections() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.info != nil && f.info.IsDir() {
		return nil
	}

	sections := []*Section{}
	for _, section := range f.config.Sections() {
		keys := map[string]model.ConfData{}
		for rawKey, rawValue := range section.KeysHash() {
			confData, err := model.NewConfData(rawKey, rawValue)
			if err != nil {
				return err
			}
			keys[confData.Key] = *confData
		}
		s := &Section{Name: section.Name(), Keys: keys}
		sections = append(sections, s)
	}
	f.Sections = sections

	return nil
}

func (f *File) convertToMap() (map[string]map[string]string, error) {
	var err error
	data := map[string]map[string]string{}
	for _, section := range f.Sections {
		sectionData := map[string]string{}
		for key, confData := range section.Keys {
			s, e := model.EncodeConfData(confData)
			if e != nil {
				err = e
				continue
			}
			sectionData[key] = s
		}
		data[section.Name] = sectionData
	}
	return data, err
}

func (f *File) Debug() {
	for _, section := range f.Sections {
		fmt.Println("sectionName:", section.Name)
		for key, cd := range section.Keys {
			fmt.Printf("%v : %v\n", key, cd)
		}
	}
}

type Files struct {
	root     string
	child    []*File
	childMap map[string]*File
	mu       sync.RWMutex
}

func OpenPath(root string) (*Files, error) {
	root = filepath.Join(root)
	files, err := openPath(root)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func InitFiles(domainName string, withLock bool) (*Files, error) {
	data, err := storage.KStorage.GetDomain(domainName, withLock)
	if err != nil {
		return nil, err
	}
	return newFilesAux(data)
}

func NewFiles(event operate.Event) (*Files, error) {
	sectionData := map[string]string{event.Key: event.Data.(string)}
	fileData := map[string]map[string]string{event.Section: sectionData}
	data := map[string]map[string]map[string]string{event.File: fileData}
	return newFilesAux(data)
}

func newFilesAux(data map[string]map[string]map[string]string) (*Files, error) {
	var err error
	files := &Files{childMap: map[string]*File{}}
	for fileName, fileData := range data {
		file := &File{Name: fileName}
		for sectionName, sectionData := range fileData {
			keys := map[string]model.ConfData{}
			for key, rawValue := range sectionData {
				confData, e := model.DecodeConfData(rawValue)
				if e != nil {
					err = e
					continue
				}
				keys[key] = *confData
			}
			s := &Section{Name: sectionName, Keys: keys}
			file.Sections = append(file.Sections, s)
		}
		files.child = append(files.child, file)
		files.childMap[fileName] = file
	}
	return files, err
}

func WrapFiles(fs []File) *Files {
	files := &Files{childMap: map[string]*File{}}
	for _, f := range fs {
		newFile := &File{Name: f.Name, Sections: f.Sections}
		files.child = append(files.child, newFile)
		files.childMap[f.Name] = newFile
	}
	return files
}

func AddFile(domainName, fileName, conf, note string) error {
	file, err := ParseFileConf(conf)
	//file.Debug()
	if err != nil {
		return err
	}
	data, err := file.convertToMap()
	if err != nil {
		return err
	}
	return storage.KStorage.AddFile(domainName, fileName, data, note)
}

func (f *Files) GetFile(fname string) (*File, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if file, ok := f.childMap[fname]; ok {
		return file, nil
	}
	return nil, os.ErrNotExist
}

func (f *Files) SetKey(file, section, key, value string) error {
	fn, err := f.GetFile(file)
	if err != nil {
		return err
	}

	confData, err := model.DecodeConfData(value)
	if err != nil {
		return err
	}

	found := false
	for _, s := range fn.Sections {
		if s.Name == section {
			s.Set(key, *confData)
			found = true
		}
	}
	if !found {
		s := &Section{Name: section, Keys: map[string]model.ConfData{}}
		s.Set(key, *confData)
		fn.Sections = append(fn.Sections, s)
	}
	return nil
}

func (f *Files) DelKey(file, section, key string) error {
	fn, err := f.GetFile(file)
	if err != nil {
		return err
	}

	for _, s := range fn.Sections {
		if s.Name == section {
			s.DeleteKey(key)
		}
	}
	return nil
}

func (f *Files) SectionKeyList(fname string, section string) map[string]model.ConfData {
	f.mu.RLock()
	defer f.mu.RUnlock()

	file, err := f.GetFile(fname)
	if err != nil {
		return map[string]model.ConfData{}
	}
	//f.Debug()
	s, err := file.GetSection(section)
	if err != nil {
		return map[string]model.ConfData{}
	}
	return s.GetKeys()
}

func (f *Files) FileList() []File {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var files []File
	for _, file := range f.child {
		if file.info != nil && file.info.IsDir() {
			continue
		}
		files = append(files, *file)
	}
	return files
}

// return file paths which exist in Files
func (f *Files) GetExistPaths(fpaths []string) (files []string) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, fpath := range fpaths {
		if _, err := f.GetFile(fpath); err == nil {
			files = append(files, fpath)
		}
	}
	return files
}

func (f *Files) Walk(fn func(file *File) error) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, file := range f.child {
		if err := fn(file); err != nil {
			return err
		}
	}
	return nil
}

func (f *Files) Debug() {
	for _, file := range f.child {
		fmt.Println("fileName:", file.Name)
		file.Debug()
	}
}

func openPath(root string) (*Files, error) {
	files := &Files{root: root, child: []*File{}, childMap: map[string]*File{}}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if root == path {
			return nil
		}
		if Ignore(info.Name(), info.IsDir()) {
			return nil
		}

		var cfg *ini.File
		if !info.IsDir() {
			if cfg, err = ini.Load(path); err != nil {
				return fmt.Errorf("load file %s error: %s", path, err)
			}
		}

		p := strings.Replace(path, root, "", 1)
		file := &File{path: filepath.Dir(p), Name: p, info: info, config: cfg}
		if err = file.initSections(); err != nil {
			return err
		}

		files.child = append(files.child, file)
		files.childMap[p] = file

		return nil
	})

	return files, err
}

func ParseFileConf(conf string) (*File, error) {
	cfg, err := ini.Load([]byte(conf))
	if err != nil {
		return nil, err
	}
	file := &File{config: cfg}
	if err = file.initSections(); err != nil {
		return nil, err
	}
	return file, nil
}

// IgnoreFile .
// 忽略以.开头的文件和目录
// 忽略后缀不为 ConfSuffix 的文件
func Ignore(name string, isDir bool) bool {
	b := []byte(name)
	if len(b) == 0 || b[0] == '.' {
		return true
	}

	hasSuffix := false
	for _, suffix := range ConfSuffix {
		if strings.HasSuffix(strings.ToLower(name), suffix) {
			hasSuffix = true
			break
		}
	}
	if !isDir && !hasSuffix {
		return true
	}

	return false
}
