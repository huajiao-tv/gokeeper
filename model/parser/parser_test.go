package parser

import "testing"

func TestTypeParser(t *testing.T) {
	key, value := "kint int", "10"
	typ, k, v, isJson, err := TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "int" || k != "kint" || v.(int) != 10 || isJson != false {
		t.Error("TypeParser error:", key, value)
	}

	key, value = "kint64 int64", "10"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "int64" || k != "kint64" || v.(int64) != int64(10) || isJson != false {
		t.Error("TypeParser error:", key, value)
	}

	key, value = "kstring string", "haha"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "string" || k != "kstring" || v.(string) != value || isJson != false {
		t.Error("TypeParser error:", key, value)
	}

	key, value = "kbool bool", "true"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "bool" || k != "kbool" || v.(bool) != true || isJson != false {
		t.Error("TypeParser error:", key, value)
	}

	key, value = "kfloat64 float64", "3.14"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "float64" || k != "kfloat64" || v.(float64) != float64(3.14) || isJson != false {
		t.Error("TypeParser error:", key, value)
	}

	key, value = "kss []string", "a,b,c"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "[]string" || k != "kss" || len(v.([]string)) != 3 || isJson != false {
		t.Error("TypeParser error:", key, value)
	}

	key, value = "ksi []int", "1,2,3"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "[]int" || k != "ksi" || len(v.([]int)) != 3 || isJson != false {
		t.Error("TypeParser error:", key, value)
	}

	key, value = "ksi64 []int64", "1,2,3"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "[]int64" || k != "ksi64" || len(v.([]int64)) != 3 || isJson != false {
		t.Error("TypeParser error:", key, value)
	}

	key, value = "ksf64 []float64", "1,2,3"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "[]float64" || k != "ksf64" || len(v.([]float64)) != 3 || isJson != false {
		t.Error("TypeParser error:", key, value)
	}

	key, value = "kmss map[string]string", "a:1,b:2,c:3"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "map[string]string" || k != "kmss" || len(v.(map[string]string)) != 3 || isJson != false {
		t.Fatal("TypeParser error:", key, value)
	}
	ss := v.(map[string]string)
	if ss["a"] != "1" || ss["b"] != "2" || ss["c"] != "3" {
		t.Fatal("TypeParser error:", key, value, ss)
	}

	key, value = "kmsi map[string]int", "a:1,b:2,c:3"
	typ, k, v, isJson, err = TypeParser(key, value)
	if err != nil {
		t.Error(err)
	}
	if typ != "map[string]int" || k != "kmsi" || len(v.(map[string]int)) != 3 || isJson != false {
		t.Fatal("TypeParser error:", key, value)
	}
	si := v.(map[string]int)
	if si["a"] != 1 || si["b"] != 2 || si["c"] != 3 {
		t.Fatal("TypeParser error:", key, value, si)
	}

	key, value = "kmis map[int]string", "1:a,2:b,3:c"
	typ, k, v, isJson, err = TypeParser(key, value)
	if typ != "map[int]string" || k != "kmis" || len(v.(map[int]string)) != 3 || isJson != false {
		t.Fatal("TypeParser error:", key, value)
	}
	is := v.(map[int]string)
	if is[1] != "a" || is[2] != "b" || is[3] != "c" {
		t.Fatal("TypeParser error:", key, value, is)
	}

	key, value = "kmii map[int]int", "1:1,2:2,3:3"
	typ, k, v, isJson, err = TypeParser(key, value)
	if typ != "map[int]int" || k != "kmii" || len(v.(map[int]int)) != 3 || isJson != false {
		t.Fatal("TypeParser error:", key, value)
	}
	ii := v.(map[int]int)
	if ii[1] != 1 || ii[2] != 2 || ii[3] != 3 {
		t.Fatal("TypeParser error:", key, value, ii)
	}

	key, value = "kmsst map[string]struct{}", "1,2,3"
	typ, k, v, isJson, err = TypeParser(key, value)
	if typ != "map[string]struct{}" || k != "kmsst" || len(v.(map[string]struct{})) != 3 || isJson != false {
		t.Fatal("TypeParser error:", key, value, typ, k, v, err)
	}

	key, value = "kmist map[int]struct{}", "1,2,3"
	typ, k, v, isJson, err = TypeParser(key, value)
	if typ != "map[int]struct{}" || k != "kmist" || len(v.(map[int]struct{})) != 3 || isJson != false {
		t.Fatal("TypeParser error:", key, value)
	}

	key, value = "kmsb map[string]bool", "1,2:false,3"
	typ, k, v, isJson, err = TypeParser(key, value)
	msb, ok := v.(map[string]bool)
	if typ != "map[string]bool" || k != "kmsb" || !ok || len(msb) != 3 || isJson != false {
		t.Fatal("TypeParser error:", key, value)
	}
	if msb["1"] != true || msb["2"] != false {
		t.Fatal("TypeParser error", key, value, msb)
	}

	key, value = "kmib map[int]bool", "1,2:false,3"
	typ, k, v, isJson, err = TypeParser(key, value)
	mib, ok := v.(map[int]bool)
	if typ != "map[int]bool" || k != "kmib" || !ok || len(mib) != 3 || isJson != false {
		t.Fatal("TypeParser error:", key, value)
	}
	if mib[1] != true || mib[2] != false {
		t.Fatal("TypeParser error", key, value, msb)
	}

	key, value = "person json", "{\"name\":\"zhangsan\",\"age\":20}"
	typ, k, v, isJson, err = TypeParser(key, value)
	if typ != "person" || k != "person" || !ok || isJson != true {
		t.Fatal("TypeParser error:", key, value)
	}

}

func TestInterfaceToString(t *testing.T) {
	var itr interface{}
	s := "abc"
	i := int(10)
	ir := "10"
	i64 := int64(64)
	i64r := "64"
	f64 := float64(3.14)
	f64r := "3.14"
	ss := []string{"a", "b", "c"}
	ssr := "a,b,c"
	si := []int{1, 2, 3}
	sir := "1,2,3"
	mss := map[string]string{"a": "1", "b": "2", "c": "3"}
	msr := "a:1,b:2,c:3"
	msi := map[string]int{"a": 1, "b": 2, "c": 3}
	mis := map[int]string{1: "10", 2: "20"}
	mir := "1:10,2:20"
	mii := map[int]int{1: 10, 2: 20}

	itr = s
	if v := interfaceToString(itr); v != s {
		t.Error("interfaceToString error", v, s)
	}

	itr = i
	if v := interfaceToString(itr); v != ir {
		t.Error("interfaceToString error", v, ir)
	}

	itr = i64
	if v := interfaceToString(itr); v != i64r {
		t.Error("interfaceToString error", v, i64r)
	}

	itr = f64
	if v := interfaceToString(itr); v != f64r {
		t.Error("interfaceToString error", v, f64r)
	}

	itr = ss
	if v := interfaceToString(itr); v != ssr {
		t.Error("interfaceToString error", v, ssr)
	}

	itr = si
	if v := interfaceToString(itr); v != sir {
		t.Error("interfaceToString error", v, sir)
	}

	return
	itr = mss
	if v := interfaceToString(itr); v != msr {
		t.Error("interfaceToString error", v, msr)
	}

	itr = msi
	if v := interfaceToString(itr); v != msr {
		t.Error("interfaceToString error", v, msr)
	}

	itr = mis
	if v := interfaceToString(itr); v != mir {
		t.Error("interfaceToString error", v, mir)
	}

	itr = mii
	if v := interfaceToString(itr); v != mir {
		t.Error("interfaceToString error", v, mir)
	}

}
