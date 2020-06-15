package conf

import (
	"path/filepath"
	"sync"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/storage/operate"
	"github.com/pkg/errors"
)

// ConfManager for managing configuration files and subscriptions
type ConfManager struct {
	// root path, for gokeeper-cli
	root       string
	backupPath string

	// see the documentation for Files
	files *Files

	// keyData lists all struct keyData, mainly for gokeeper-cli
	keyData map[string]map[string]KeyData

	sync.RWMutex
}

func InitConfManager(domainName string, withLock bool) (*ConfManager, error) {
	files, err := InitFiles(domainName, withLock)
	if err != nil {
		return nil, err
	}
	cm := &ConfManager{files: files}
	return cm, err
}

func NewConfManager(event operate.Event) (*ConfManager, error) {
	files, err := NewFiles(event)
	if err != nil {
		return nil, err
	}
	cm := &ConfManager{files: files}
	return cm, err
}

// NewConfManager return a new ConfManager given a path
func NewConfManagerFromLocal(root string, backupPath string) (*ConfManager, error) {
	files, err := OpenPath(root)
	if err != nil {
		return nil, err
	}

	cm := &ConfManager{root: root, backupPath: backupPath, files: files}
	if err := cm.initKeyData(); err != nil {
		return nil, err
	}
	return cm, nil
}

func WrapConfManager(fs []File) (*ConfManager, error) {
	files := WrapFiles(fs)
	cm := &ConfManager{files: files}
	if err := cm.initKeyData(); err != nil {
		return nil, err
	}
	return cm, nil
}

func (c *ConfManager) NewFile(event operate.Event) (*File, error) {
	files, err := NewFiles(event)
	if err != nil {
		return nil, err
	}
	if len(files.child) == 0 {
		return nil, errors.New("get files error")
	}

	file := files.child[0]
	c.files.child = append(c.files.child, file)
	c.files.childMap[event.File] = file

	return file, nil
}

func (c *ConfManager) initKeyData() error {
	c.Lock()
	defer c.Unlock()

	c.keyData = map[string]map[string]KeyData{}
	err := c.files.Walk(func(file *File) error {
		structName := model.GetStructName(filepath.Base(file.Name))
		structKeyData, ok := c.keyData[structName]
		if !ok {
			structKeyData = map[string]KeyData{}
			c.keyData[structName] = structKeyData
		}

		for _, cfd := range file.KeyList() {
			keyData, ok := structKeyData[cfd.Key]
			if !ok {
				keyData = NewKeyData(filepath.Join(file.path, filepath.Base(file.Name)), cfd)
				structKeyData[cfd.Key] = keyData
			}
			if keyData.Type != cfd.Type {
				return errors.Errorf("key %s type conflict: %s, %s", cfd.Key, keyData.file, file.Name)
			}
		}
		return nil
	})

	return err
}

func (c *ConfManager) GetRoot() string {
	return c.root
}

func (c *ConfManager) GetKeyData() map[string]map[string]KeyData {
	c.RLock()
	k := c.keyData
	c.RUnlock()
	return k
}

func (c *ConfManager) Update(event operate.Event) error {
	c.Lock()
	defer c.Unlock()
	return c.files.SetKey(event.File, event.Section, event.Key, event.Data.(string))
}

func (c *ConfManager) Delete(event operate.Event) error {
	c.Lock()
	defer c.Unlock()
	return c.files.DelKey(event.File, event.Section, event.Key)
}

func (c *ConfManager) GetFiles() *Files {
	c.Lock()
	defer c.Unlock()
	return c.files
}

func (c *ConfManager) FileList() []File {
	c.RLock()
	fs := c.files.FileList()
	c.RUnlock()
	return fs
}

// ParseSubscribe handle subscription relationship and return real subscription file
//
// Subscription file format
//  - subscribe section   : /subscribe/file.conf/section
//  - subscribe file	  : /subscribe/file.conf
//  - subscribe directory : /subscribe/dir/ -> deprecated?
func (c *ConfManager) ParseSubscribe(rawSubscriptions []string) []string {
	var subscribed []string
	for _, s := range rawSubscriptions {
		s = filepath.Join(PathSeparator, s)
		subscribed = append(subscribed, c.parseSubscribe(s)...)
	}
	return subscribed
}

// parse rawSubscription and its inherited default section
func (c *ConfManager) parseSubscribe(rawSubscription string) []string {
	c.RLock()
	files := c.files
	c.RUnlock()

	var fileList []string
	paths := files.GetExistPaths(Subscription(rawSubscription).InvolvedFilesPath())
	for _, path := range paths {
		fileList = append(fileList, filepath.Join(path, DefaultSection))
	}

	switch Subscription(rawSubscription).Type() {
	case SubscribeFile:
	case SubscribeSection:
		//@todo bugfix  when subscribe /sr-bjcc/global.conf/127.0.0.1:80, now include [/global.conf/DEFAULT /sr-bjcc/global.conf/127.0.0.1:80], not include /global.conf/127.0.0.1:80
		//section := Subscription(rawSubscription).Section()
		//for _, path := range paths {
		//	fileList = append(fileList, filepath.Join(path, section))
		//}
		fileList = append(fileList, rawSubscription)
	default:
		return []string{}
	}
	return fileList
}

// Subscribe return struct data
func (c *ConfManager) Subscribe(subscriptions []string) []model.StructData {
	c.RLock()
	files := c.files
	c.RUnlock()

	cfdMaps := map[string]map[string]model.ConfData{}
	for _, subscription := range subscriptions {
		s := Subscription(subscription)
		if s.Type() != SubscribeSection {
			continue
		}
		structName := model.GetStructName(s.File())
		cfdMap, ok := cfdMaps[structName]
		if !ok {
			cfdMap = map[string]model.ConfData{}
			cfdMaps[structName] = cfdMap
		}
		keys := files.SectionKeyList(s.File(), s.Section())
		for _, cfd := range keys {
			cfdMap[cfd.Key] = cfd
		}
	}

	var structDatas []model.StructData
	for structName, cfdMap := range cfdMaps {
		structDatas = append(structDatas, model.NewStructData(structName, 0, cfdMap))
	}
	return structDatas
}

func (c *ConfManager) Debug() {
	c.files.Debug()
}
