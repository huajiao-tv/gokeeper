package apihttp

import (
	"crypto/md5"
	"fmt"
	"strings"
)

const (
	InnerSecretKey = "D2BsaMhN4IPnsFJ8NPIki83n2f6xdC0s"
)

type GuidParams struct {
	Partner string
	Rand    string
	Time    string
}

func GetServerGUID(params *GuidParams) string {
	verifyParams := []string{params.Partner, params.Rand, params.Time}
	verifySign := fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(verifyParams, "_")+InnerSecretKey)))
	return verifySign
}

// 校验服务端的请求里的guid是否合法
func CheckServerGUID(params *GuidParams, guid string) bool {
	verifySign := GetServerGUID(params)
	return strings.ToLower(verifySign) == guid
}
