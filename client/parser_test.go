package client

import (
	"testing"

	"github.com/huajiao-tv/gokeeper/model"
)

type ttt struct {
	Kint    int
	Kstring string
	Kbool   bool
	Kmapss  map[string]string
	Kmapsi  map[string]int
	Kmapis  map[int]string
	Kmapii  map[int]int
	Kss     []string
	Ksi     []int
}

var (
	td = map[string]model.ConfData{
		"Int":    model.ConfData{Type: "int", Key: "Kint", Value: 10, StructKey: "Kint"},
		"String": model.ConfData{Type: "string", Key: "Kstring", Value: "hello world", StructKey: "Kstring"},
		"Bool":   model.ConfData{Type: "bool", Key: "Kbool", Value: false, StructKey: "Kbool"},
		"MapSS":  model.ConfData{Type: "map[string]string", Key: "Kmapss", Value: map[string]string{"a": "1"}, StructKey: "Kmapss"},
	}
)

func TestSetStructField(t *testing.T) {
	var itr interface{}
	itr = &ttt{}

	setStructField(itr, td)

	v := itr.(*ttt)

	if v.Kint != 10 {
		t.Error("setStructField error:int ", v.Kint)
	}

}
