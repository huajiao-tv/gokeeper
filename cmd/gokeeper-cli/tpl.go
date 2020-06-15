package main

import (
	"bytes"
	"go/format"
	"strings"
	"text/template"

	"github.com/huajiao-tv/gokeeper/model"
)

func parseTpl(structDatas []model.StructData) (map[string]string, error) {
	tpl := map[string]string{}
	m, err := parseTplStruct(structDatas)
	if err != nil {
		return nil, err
	}
	tpl = mergeMapString(tpl, m)

	m, err = parseTplMeta(structDatas)
	if err != nil {
		return nil, err
	}
	tpl = mergeMapString(tpl, m)

	return tpl, nil
}

func parseTplStruct(structDatas []model.StructData) (map[string]string, error) {
	tpl := map[string]string{}
	for _, v := range structDatas {
		tmpl, err := template.New("tplStruct").Parse(tplStruct)
		if err != nil {
			return nil, err
		}
		var b bytes.Buffer
		err = tmpl.Execute(&b, v)
		if err != nil {
			return nil, err
		}
		fb, err := format.Source(b.Bytes())
		if err != nil {
			return nil, err
		}
		tpl[strings.ToLower(v.Name)] = string(fb)
	}
	return tpl, nil
}

func parseTplMeta(structDatas []model.StructData) (map[string]string, error) {
	var b bytes.Buffer
	tmpl, err := template.New("tplMeta").Parse(tplMeta)
	if err != nil {
		return nil, err
	}
	err = tmpl.Execute(&b, structDatas)
	if err != nil {
		return nil, err
	}
	tpl := map[string]string{}
	fb, err := format.Source(b.Bytes())
	if err != nil {
		return nil, err
	}
	tpl["meta"] = string(fb)
	return tpl, nil
}

func mergeMapString(dst, src map[string]string) map[string]string {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
