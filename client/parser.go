package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/huajiao-tv/gokeeper/model"
)

func setStructField(itr interface{}, data map[string]model.ConfData) error {
	rfv := reflect.ValueOf(itr).Elem()
	for _, v := range data {
		field := rfv.FieldByName(v.StructKey)
		if !field.IsValid() {
			Stderr.Write([]byte(fmt.Sprintf("%s|gokeeper|setStructField|field invalid|%s \n", time.Now().String(), v.StructKey)))
			continue
		}
		var fieldValue reflect.Value
		typ := field.Type().String()
		if v.IsJson {
			st := reflect.New(field.Type())
			if !strings.Contains(strings.Replace(typ, " ", "", -1), v.Type) {
				Stderr.Write([]byte(fmt.Sprintf("%s|gokeeper|setStructField|field type invalid|%s|%s|%s \n", time.Now().String(), v.StructKey, typ, v.Type)))
				continue
			}
			jsonStr, ok := v.Value.(string)
			if !ok {
				Stderr.Write([]byte(fmt.Sprintf("%s|gokeeper|setStructField|field type invalid|json value is not string|%s|%s|%s \n", time.Now().String(), v.StructKey, field.Type().String(), v.Type)))
				continue
			}
			err := json.Unmarshal([]byte(jsonStr), st.Interface())
			if err != nil {
				Stderr.Write([]byte(fmt.Sprintf("%s|gokeeper|setStructField|field type invalid|json unmarshal error|%s|%s|%s|%s|%s \n", time.Now().String(), err, v.StructKey, field.Type().String(), v.Type, jsonStr)))
				continue
			}
			fieldValue = st.Elem()
		} else {
			if strings.Replace(typ, " ", "", -1) != v.Type {
				Stderr.Write([]byte(fmt.Sprintf("%s|gokeeper|setStructField|field type invalid|%s|%s|%s \n", time.Now().String(), v.StructKey, typ, v.Type)))
				continue
			}
			fieldValue = reflect.ValueOf(v.Value)
		}
		field.Set(fieldValue)
	}
	return nil
}

func fill(rdata map[string]interface{}, sd model.StructData) (interface{}, error) {
	structInterface, ok := rdata[sd.Name]
	if !ok {
		return nil, errors.New("struct not load:" + sd.Name)
	}
	itr := reflect.New(reflect.TypeOf(structInterface)).Interface()
	if err := setStructField(itr, sd.Data); err != nil {
		return nil, err
	}

	return itr, nil
}
