package model

import (
	"encoding/gob"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"encoding/json"
	"fmt"

	"github.com/huajiao-tv/gokeeper/model/parser"
)

func init() {
	gob.Register([]StructData{})
}

type StructData struct {
	Name    string              `json:"name"`
	Version int                 `json:"version"`
	Data    map[string]ConfData `json:"data"`

	//only for gokeeper-cli
	Libraries []string
}

func NewStructData(name string, version int, data map[string]ConfData) StructData {
	// do not need to initialize Libraries
	return StructData{Name: name, Version: version, Data: data}
}

func (s *StructData) SetVersion(version int) {
	s.Version = version
}

type ConfData struct {
	Type      string      `json:"type"`
	RawKey    string      `json:"raw_key"`
	RawValue  string      `json:"raw_value"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	StructKey string      `json:"struct_key"`
	IsJson    bool        `json:"is_json"`
}

func NewConfData(rawKey, rawValue string) (*ConfData, error) {
	typ, key, value, isJson, err := parser.TypeParser(rawKey, rawValue)
	if err != nil {
		return nil, err
	}
	if isJson {
		typ = ToCamlCase(typ)
	}
	return &ConfData{Type: typ, Key: key, RawKey: rawKey, RawValue: rawValue, Value: value, StructKey: ToCamlCase(key), IsJson: isJson}, nil
}

// ToCamlCase convert snake_case to CamlCase
func ToCamlCase(key string) string {
	ks := strings.Split(key, "_")
	for k, v := range ks {
		ks[k] = ToUpperFirst(v)
	}
	return strings.Join(ks, "")
}

// ToUpperFirst return first letter to upper
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

func EncodeConfData(cd ConfData) (string, error) {
	bytes, err := json.Marshal(cd)
	return string(bytes), err
}

func DecodeConfData(s string) (*ConfData, error) {
	var cd ConfData
	var err error
	if err = json.Unmarshal([]byte(s), &cd); err != nil {
		return nil, err
	}
	//@todo 有时间再优化一下该数据结构
	if cd.IsJson {
		cd.Value = cd.RawValue
	} else {
		cd.Value, err = parser.TypeParserAux(cd.Type, cd.RawValue)
	}
	return &cd, err
}

func GetRawKey(key, typ string) string {
	return strings.Trim(fmt.Sprintf("%s %s", key, typ), " ")
}
