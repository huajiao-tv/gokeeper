package parser

import (
	"encoding/gob"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
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
		"map[string][]string": map[string][]string{},
		"map[string]int":      map[string]int{},
		"map[string]bool":     map[string]bool{},
		"map[string]struct{}": map[string]struct{}{},
		"map[int]string":      map[int]string{},
		"map[int]int":         map[int]int{},
		"map[int]bool":        map[int]bool{},
		"map[int]struct{}":    map[int]struct{}{},
		"time.Duration":       time.Duration(0),
	}
)

func init() {
	gob.Register(map[string]string{})
	gob.Register(map[string]int{})
	gob.Register(map[string]bool{})
	gob.Register(map[string]struct{}{})
	gob.Register(map[int]string{})
	gob.Register(map[int]int{})
	gob.Register(map[int]bool{})
	gob.Register(map[int]struct{}{})
	gob.Register(map[string][]string{})
	gob.Register(time.Duration(0))
}

// TypeParser convert raw key and raw value to type, key, value
func TypeParser(rawKey string, rawValue string) (typ string, key string, value interface{}, isJson bool, err error) {
	ks := strings.Fields(strings.Trim(rawKey, " "))
	lks := len(ks)
	if lks == 1 {
		return "string", rawKey, rawValue, false, nil
	}

	if lks == 0 || lks > 2 {
		return "string", rawKey, rawValue, false, fmt.Errorf("key invalid, key=%s value=%s", rawKey, rawValue)
	}

	if ks[1] == "json" {
		return ks[0], ks[0], rawValue, true, nil
	}

	key, typ = ks[0], strings.Replace(ks[1], " ", "", -1)
	value, err = TypeParserAux(typ, rawValue)
	if err != nil {
		return "string", rawKey, rawValue, false, fmt.Errorf("parse key error:%s, key=%s value=%s", err.Error(), rawKey, rawValue)
	}
	return typ, key, value, false, nil
}

func TypeParserAux(typ string, value string) (interface{}, error) {
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
	case "map[string][]string":
		v, err = parserMapsss(value)
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
	case "time.Duration":
		v, err = time.ParseDuration(value)
	default:
		v, err = nil, errors.New("switch type unsupport:"+typ)
	}

	return v, err
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

func parserMapsss(value string) (map[string][]string, error) {
	if value == "" {
		return map[string][]string{}, nil
	}
	s := strings.Split(value, ";")
	si := map[string][]string{}
	for _, v := range s {
		ar := strings.SplitN(v, ":", 2)
		if len(ar) != 2 {
			return nil, errors.New("key type map[string][]string format invalid:" + v)
		}
		si[ar[0]] = strings.Split(ar[1], ",")
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

func interfaceToString(itr interface{}) string {
	switch itr.(type) {
	case int, int64, float64, bool, string:
		return fmt.Sprintf("%v", itr)
	case []string, []int, []int64, []float64, []bool:
		return sliceToString(itr)
	case map[string]string, map[string]int, map[int]string, map[int]int, map[string]struct{}, map[int]struct{}, map[string]bool, map[int]bool:
		return mapToString(itr)
	}
	return ""
}

func sliceToString(itr interface{}) string {
	s := []string{}
	switch itr.(type) {
	case []string:
		s = itr.([]string)
	case []int:
		for _, v := range itr.([]int) {
			s = append(s, fmt.Sprintf("%v", v))
		}
	case []int64:
		for _, v := range itr.([]int64) {
			s = append(s, fmt.Sprintf("%v", v))
		}
	case []float64:
		for _, v := range itr.([]float64) {
			s = append(s, fmt.Sprintf("%v", v))
		}
	case []bool:
		for _, v := range itr.([]bool) {
			s = append(s, fmt.Sprintf("%v", v))
		}
	}
	return strings.Join(s, ",")
}

func mapToString(itr interface{}) string {
	s := []string{}
	switch itr.(type) {
	case map[string]string:
		for k, v := range itr.(map[string]string) {
			s = append(s, fmt.Sprintf("%v:%v", k, v))
		}
	case map[string]int:
		for k, v := range itr.(map[string]int) {
			s = append(s, fmt.Sprintf("%v:%v", k, v))
		}
	case map[int]string:
		for k, v := range itr.(map[int]string) {
			s = append(s, fmt.Sprintf("%v:%v", k, v))
		}
	case map[int]int:
		for k, v := range itr.(map[int]int) {
			s = append(s, fmt.Sprintf("%v:%v", k, v))
		}
	case map[string]bool:
		for k, v := range itr.(map[string]bool) {
			key := k
			if v == false {
				key = fmt.Sprintf("%v:%v", key, v)
			}
			s = append(s, key)
		}
	case map[int]bool:
		for k, v := range itr.(map[int]bool) {
			key := strconv.Itoa(k)
			if v == false {
				key = fmt.Sprintf("%v:%v", key, v)
			}
			s = append(s, key)
		}
	case map[string]struct{}:
		for k, _ := range itr.(map[string]struct{}) {
			s = append(s, k)
		}
	case map[int]struct{}:
		for k, _ := range itr.(map[int]struct{}) {
			s = append(s, strconv.Itoa(k))
		}
	}
	return strings.Join(s, ",")
}
