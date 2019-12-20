package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/lvsz1/gojson"
)

// json类型的field
type JF struct {
	Type  string //json struct类型
	Value string //json数据
}

// 根据json生成结构体，并写道.go文件中
func json2Struct(packageName, outputFile string, jsonFields map[string]JF) error {
	tagList := []string{"json"}

	output := fmt.Sprintf("package %s\n\n", packageName)
	for _, v := range jsonFields {
		r := strings.NewReader(v.Value)
		b, err := gojson.Generate(r, gojson.ParseJson, v.Type, tagList, false, true)
		if err != nil {
			return err
		}
		output += string(b) + "\n\n"
	}

	err := ioutil.WriteFile(outputFile, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil
}
