package apihttp

import (
	"crypto/md5"
	"fmt"
	"strings"
)

const (
	InnerSecretKey = "5ff10ecc78ada17c37b96fdf1ecb0c9e"
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
