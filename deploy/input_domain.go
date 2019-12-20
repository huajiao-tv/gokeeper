package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/huajiao-tv/gokeeper/utility/go-ini/ini"
)

var (
	typMap = map[string]interface{}{
		"bool":                true,
		"int":                 int(0),
		"int64":               int64(0),
		"float64":             float64(0),
		"string":              "",
		"[]string":            []string{},
		"[]int":               []int{},
		"[]int64":             []int64{},
		"[]float64":           []float64{},
		"[]bool":              []bool{},
		"map[string]string":   map[string]string{},
		"map[string]int":      map[string]int{},
		"map[string]bool":     map[string]bool{},
		"map[string]struct{}": map[string]struct{}{},
		"map[int]string":      map[int]string{},
		"map[int]int":         map[int]int{},
		"map[int]bool":        map[int]bool{},
		"map[int]struct{}":    map[int]struct{}{},
	}
)
var domainPath = flag.String("c", "", "domain config path")
var keeperAdminUrl = flag.String("k", "", "keeper admin url")
var ConfSuffix = []string{".conf"}

func NewConfManager(root string, backupPath string) (*ConfManager, error) {
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

func main() {
	flag.Parse()
	infos, err := ioutil.ReadDir(*domainPath)
	if err != nil {
		fmt.Println("read file error:", err)
		return
	}
	for _, info := range infos {
		if !info.IsDir() {
		}
		if []byte(info.Name())[0] == '.' {
		}
		domain := info.Name()

		configs := getDomain(domain)
		cf, err := NewConfManager(filepath.Join(*domainPath, info.Name()), filepath.Join("~", info.Name()))
		if err != nil {
		}
		for _, v := range cf.keyData {
			for _, key := range v {
				for _, f := range configs {
					for _, v := range f.Sections {
						if v.Name == "DEFAULT" {
							for _, v := range v.Keys {
								if v.Key == key.Key && v.Type == key.Type && v.RawValue == key.RawValue {
									goto END
								}
							}
						}
					}
				}
				if !addConfig(domain, strings.Trim(key.file, "/"), "update", key.Key, key.Value, "DEFAULT", key.Type) {
					fmt.Println("key init failed:", key)
				}
				fmt.Println("key change:", key.Key)
			END:
			}
		}
	}
}

type Resp struct {
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

func addConfig(domain string, file string, operate string, key string, value interface{}, section string, typ string) bool {
	resp, err := http.Get(fmt.Sprintf("http://%v/conf/manage?domain=%v&note=init&operates=[{\"opcode\":\"%v\",\"file\":\"%v\",\"key\":\"%v\",\"value\":\"%v\",\"section\":\"%s\",\"type\":\"%v\"}]", *keeperAdminUrl, domain, operate, file, key, value, section, typ))
	if err != nil {
		fmt.Println("manage domain config error:", err)
		return false
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var res Resp
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("json unmarsha1 error :", err)
		return false
	}
	if res.ErrorCode != 0 {
		fmt.Println("error data:", res.Error)
		return false
	}
	return true
}

type DomainListResp struct {
	Data []struct {
		Name    string `json:"name"`
		Version int    `json:"version"`
	} `json:"data"`
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

func getDomainList() *DomainListResp {
	resp, err := http.Get("http://" + *keeperAdminUrl + "/domain/list")
	if err != nil {
		fmt.Println("get domain error:", err)
		return nil
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var res DomainListResp
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("json unmarsha1 error :", err)
		return nil
	}
	if res.ErrorCode != 0 {
		fmt.Println("resp error :", res.Error, res.ErrorCode)
		return nil
	}
	return &res
}

type DomainConfResp struct {
	Data      []File `json:"data"`
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

func getDomain(domainName string) []File {
	resp, err := http.Get("http://" + *keeperAdminUrl + "/get/domain/from/conf?domain=" + domainName)
	if err != nil {
		fmt.Println("get domain error:", err)
		return nil
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var res DomainConfResp
	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("json unmarsha1 error :", err)
		return nil
	}
	return res.Data
}

func OpenPath(root string) (*Files, error) {
	root = filepath.Join(root)
	files, err := openPath(root)
	if err != nil {
		return nil, err
	}
	return files, nil
}

type Files struct {
	root     string
	child    []*File
	childMap map[string]*File
	mu       sync.RWMutex
}

type File struct {
	Name     string    `json:"name"`
	Sections []Section `json:"sections"`

	path   string
	info   os.FileInfo
	config *ini.File
	mu     sync.RWMutex
}

type Section struct {
	Name string              `json:"name"`
	Keys map[string]ConfData `json:"keys"`
}

type ConfData struct {
	Type      string      `json:"type"`
	RawKey    string      `json:"raw_key"`
	RawValue  string      `json:"raw_value"`
	Key       string      `json:"key"`
	Value     interface{} `json:"-"`
	StructKey string      `json:"-"`
}

func NewConfData(rawKey, rawValue string) (*ConfData, error) {
	typ, key, value, err := TypeParser(rawKey, rawValue)
	if err != nil {
		return nil, err
	}
	return &ConfData{Type: typ, Key: key, RawKey: rawKey, RawValue: rawValue, Value: value, StructKey: ToCamlCase(key)}, nil
}
func ToCamlCase(key string) string {
	ks := strings.Split(key, "_")
	for k, v := range ks {
		ks[k] = ToUpperFirst(v)
	}
	return strings.Join(ks, "")
}
func TypeParser(rawKey string, rawValue string) (typ string, key string, value interface{}, err error) {
	ks := strings.Fields(strings.Trim(rawKey, " "))
	lks := len(ks)
	if lks == 1 {
		return "string", rawKey, rawValue, nil
	}
	if lks == 0 || lks > 2 {
		return "string", rawKey, rawValue, fmt.Errorf("key invalid, key=%s value=%s", rawKey, rawValue)
	}

	key, typ = ks[0], strings.Replace(ks[1], " ", "", -1)
	value, err = typeParser(typ, rawValue)
	if err != nil {
		return "string", rawKey, rawValue, fmt.Errorf("parse key error:%s, key=%s value=%s", err.Error(), rawKey, rawValue)
	}
	return typ, key, value, nil
}
func typeParser(typ string, value string) (interface{}, error) {
	_, ok := typMap[typ]
	if !ok {
		return value, errors.New("type unsupport:" + typ)
	}

	var v interface{}
	var err error

	switch typ {
	case "string":
		return value, nil
	case "bool":
		v, err = strconv.ParseBool(value)
	case "int":
		v, err = strconv.Atoi(value)
	case "int64":
		v, err = strconv.ParseInt(value, 10, 64)
	case "float64":
		v, err = strconv.ParseFloat(value, 64)
	case "[]string":
		v, err = strings.Split(value, ","), nil
	case "[]int":
		v, err = parserSi(value)
	case "[]int64":
		v, err = parserSi64(value)
	case "[]float64":
		v, err = parserSf64(value)
	case "[]bool":
		v, err = parserSb(value)
	case "map[string]string":
		v, err = parserMapss(value)
	case "map[string]int":
		v, err = parserMapsi(value)
	case "map[int]string":
		v, err = parserMapis(value)
	case "map[int]int":
		v, err = parserMapii(value)
	case "map[string]bool":
		v, err = parserMapsb(value)
	case "map[int]bool":
		v, err = parserMapib(value)
	case "map[string]struct{}":
		v, err = parserMapsst(value)
	case "map[int]struct{}":
		v, err = parserMapist(value)
	default:
		v, err = nil, errors.New("switch type unsupport:"+typ)
	}

	return v, err
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
func (f *File) initSections() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.info.IsDir() {
		return nil
	}

	sections := []Section{}
	for _, section := range f.config.Sections() {
		keys := map[string]ConfData{}
		for rawKey, rawValue := range section.KeysHash() {
			confData, err := NewConfData(rawKey, rawValue)
			if err != nil {
				return err
			}
			keys[confData.Key] = *confData
		}
		s := Section{Name: section.Name(), Keys: keys}
		sections = append(sections, s)
	}
	f.Sections = sections

	return nil
}

type ConfManager struct {
	// root path
	root       string
	backupPath string

	// see the documentation for Files
	files *Files

	// keyData lists all struct keyData, mainly for the detection of the type of conflict
	keyData map[string]map[string]KeyData

	sync.RWMutex
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
func (f *File) GetInfo() os.FileInfo {
	return f.info
}

type KeyData struct {
	ConfData
	file string
}

func (f *File) KeyList() []ConfData {
	f.mu.RLock()
	defer f.mu.RUnlock()
	keys := []ConfData{}
	for _, section := range f.GetSections() {
		for _, cfd := range section.GetKeys() {
			keys = append(keys, cfd)
		}
	}
	return keys
}
func (f *File) GetSections() []Section {
	f.mu.RLock()
	sections := f.Sections
	f.mu.RUnlock()
	return sections
}
func (s Section) GetKeys() map[string]ConfData {
	return s.Keys
}

func NewKeyData(file string, cfd ConfData) KeyData {
	return KeyData{file: file, ConfData: cfd}
}
func (c *ConfManager) initKeyData() error {
	c.Lock()
	defer c.Unlock()

	c.keyData = map[string]map[string]KeyData{}
	err := c.files.Walk(func(file *File) error {
		if file.GetInfo().IsDir() {
			return nil
		}
		structName := GetStructName(file.GetInfo().Name())
		structKeyData, ok := c.keyData[structName]
		if !ok {
			structKeyData = map[string]KeyData{}
			c.keyData[structName] = structKeyData
		}

		for _, cfd := range file.KeyList() {
			keyData, ok := structKeyData[cfd.Key]
			if !ok {
				keyData = NewKeyData(filepath.Join(file.path, file.GetInfo().Name()), cfd)
				structKeyData[cfd.Key] = keyData
			}
			if keyData.Type != cfd.Type {
				return fmt.Errorf("key %s type conflict: %s, %s", cfd.Key, keyData.file, filepath.Join(file.path, file.GetInfo().Name()))
			}
		}

		return nil
	})

	return err
}

func ToUpperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

func GetStructName(fname string) string {
	fname = filepath.Base(fname)
	f := strings.Split(fname, ".")
	return ToUpperFirst(f[0])
}

func parserSf64(value string) ([]float64, error) {
	if value == "" {
		return []float64{}, nil
	}
	s := strings.Split(value, ",")
	si := []float64{}
	for _, v := range s {
		i, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		si = append(si, i)
	}
	return si, nil
}

func parserSi64(value string) ([]int64, error) {
	if value == "" {
		return []int64{}, nil
	}
	s := strings.Split(value, ",")
	si := []int64{}
	for _, v := range s {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		si = append(si, i)
	}
	return si, nil
}

func parserSi(value string) ([]int, error) {
	if value == "" {
		return []int{}, nil
	}

	s := strings.Split(value, ",")
	si := []int{}
	for _, v := range s {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		si = append(si, i)
	}
	return si, nil
}

func parserSb(value string) ([]bool, error) {
	if value == "" {
		return []bool{}, nil
	}

	s := strings.Split(value, ",")
	si := []bool{}
	for _, v := range s {
		i, err := strconv.ParseBool(v)
		if err != nil {
			return si, err
		}
		si = append(si, i)
	}
	return si, nil
}

func parserMapss(value string) (map[string]string, error) {
	if value == "" {
		return map[string]string{}, nil
	}
	s := strings.Split(value, ",")
	si := map[string]string{}
	for _, v := range s {
		ar := strings.SplitN(v, ":", 2)
		if len(ar) != 2 {
			return nil, errors.New("key type map[string]string format invalid:" + v)
		}
		si[ar[0]] = ar[1]
	}
	return si, nil
}

func parserMapsi(value string) (map[string]int, error) {
	if value == "" {
		return map[string]int{}, nil
	}
	var err error
	s := strings.Split(value, ",")
	si := map[string]int{}
	for _, v := range s {
		ar := strings.SplitN(v, ":", 2)
		if len(ar) != 2 {
			return nil, errors.New("key type map[string]int format invalid:" + v)
		}
		if si[ar[0]], err = strconv.Atoi(ar[1]); err != nil {
			return nil, err
		}
	}
	return si, nil
}

func parserMapis(value string) (map[int]string, error) {
	if value == "" {
		return map[int]string{}, nil
	}
	s := strings.Split(value, ",")
	si := map[int]string{}
	for _, v := range s {
		ar := strings.SplitN(v, ":", 2)
		if len(ar) != 2 {
			return nil, errors.New("key type map[int]string format invalid:" + v)
		}
		sk, err := strconv.Atoi(ar[0])
		if err != nil {
			return nil, err
		}
		si[sk] = ar[1]
	}
	return si, nil
}

func parserMapii(value string) (map[int]int, error) {
	if value == "" {
		return map[int]int{}, nil
	}
	s := strings.Split(value, ",")
	si := map[int]int{}
	for _, v := range s {
		ar := strings.SplitN(v, ":", 2)
		if len(ar) != 2 {
			return nil, errors.New("key type map[int]int format invalid:" + v)
		}
		sk, err := strconv.Atoi(ar[0])
		if err != nil {
			return nil, err
		}
		sv, err := strconv.Atoi(ar[1])
		if err != nil {
			return nil, err
		}
		si[sk] = sv
	}
	return si, nil
}

func parserMapsst(value string) (map[string]struct{}, error) {
	if value == "" {
		return map[string]struct{}{}, nil
	}
	s := strings.Split(value, ",")
	si := map[string]struct{}{}
	for _, v := range s {
		si[v] = struct{}{}
	}
	return si, nil
}

func parserMapist(value string) (map[int]struct{}, error) {
	if value == "" {
		return map[int]struct{}{}, nil
	}
	s := strings.Split(value, ",")
	si := map[int]struct{}{}
	for _, v := range s {
		sk, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		si[sk] = struct{}{}
	}
	return si, nil
}

func parserMapsb(value string) (map[string]bool, error) {
	if value == "" {
		return map[string]bool{}, nil
	}
	s := strings.Split(value, ",")
	si := map[string]bool{}
	for _, v := range s {
		key := v
		value := true
		sk := strings.Split(v, ":")
		if len(sk) == 2 {
			if b, err := strconv.ParseBool(sk[1]); err == nil {
				key = sk[0]
				value = b
			}
		}
		si[key] = value
	}
	return si, nil
}

func parserMapib(value string) (map[int]bool, error) {
	if value == "" {
		return map[int]bool{}, nil
	}
	s := strings.Split(value, ",")
	si := map[int]bool{}
	for _, v := range s {
		key := v
		value := true
		sk := strings.Split(v, ":")
		if len(sk) == 2 {
			if b, err := strconv.ParseBool(sk[1]); err == nil {
				key = sk[0]
				value = b
			}
		}

		i, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		si[i] = value
	}
	return si, nil
}
