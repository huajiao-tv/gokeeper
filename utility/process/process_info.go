package process

import (
	"strings"
)

// 启动文件格式 etc/cluster/processName-farm-num.conf
func GetProcessInfo(f string) (processName, farm, num string) {
	infoFile := strings.Split(f, "/")
	if len(infoFile) < 3 {
		return
	}
	info := strings.Split(infoFile[len(infoFile)-1], "-")
	if len(info) != 3 {
		return
	}
	processName = info[0]
	farm = info[1]
	farmInfo := strings.Split(info[2], ".")
	if len(farmInfo) != 2 {
		return
	}
	num = farmInfo[0]
	return
}
