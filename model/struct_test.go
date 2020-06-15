package model

import "testing"

func TestNewConfData(t *testing.T) {
	rawKey := "hello"
	rawValue := "world"

	cfd, err := NewConfData(rawKey, rawValue)
	if err != nil {
		t.Error(err.Error())
	}
	if cfd.Type != "string" || cfd.RawKey != rawKey || cfd.Value.(string) != rawValue {
		t.Errorf("NewConfData error: %v, %v, %v", rawKey, rawValue, cfd)
	}

}

func TestToUpperFirst(t *testing.T) {
	s := "abc"
	expect := "Abc"
	actual := ToUpperFirst(s)
	if actual != expect {
		t.Error(expect, actual)
	}
}

func TestToCamlCase(t *testing.T) {
	s := "a_bc_def"
	expect := "ABcDef"
	actual := ToCamlCase(s)
	if actual != expect {
		t.Error(expect, actual)
	}
}

func TestGetStructName(t *testing.T) {
	s := "hello.conf"
	expect := "Hello"
	actual := GetStructName(s)
	if expect != actual {
		t.Error(expect, actual)
	}
}
