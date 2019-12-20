package etcd

import "encoding/json"

type domainVersion struct {
	Version     int64 `json:"version"`
	EtcdVersion int64 `json:"etcd_version"`
}

func encodeDomainVersion(dv domainVersion) (string, error) {
	bytes, err := json.Marshal(dv)
	return string(bytes), err
}

func decodeDomainVersion(data string) (*domainVersion, error) {
	var dv domainVersion
	err := json.Unmarshal([]byte(data), &dv)
	return &dv, err
}

type Recode struct {
	ID             int64  `json:"id"` //用于兼容老版本gokeeper
	Domain         string `json:"domain"`
	Version        int64  `json:"version"`
	PackageVersion int64  `json:"package_version"` // etcdVersion
	Note           string `json:"note"`            // 备注
	Timestamp      int64  `json:"timestamp"`       // 时间
}

func encodeRecode(r Recode) (string, error) {
	bytes, err := json.Marshal(r)
	return string(bytes), err
}

func decodeRecode(data string) (*Recode, error) {
	var r Recode
	err := json.Unmarshal([]byte(data), &r)
	return &r, err
}

func mapDiff(lm, rm map[string]map[string]map[string]string, needCorrection bool) map[string]map[string]map[string]string {
	dm := map[string]map[string]map[string]string{}
	for lk1, lm1 := range lm {
		rm1, exist := rm[lk1]
		if !exist {
			dm[lk1] = lm1
			continue
		}
		dm1 := map[string]map[string]string{}
		for lk2, lm2 := range lm1 {
			rm2, exist := rm1[lk2]
			if !exist {
				dm1[lk2] = lm2
				continue
			}
			dm2 := map[string]string{}
			for lk3, lm3 := range lm2 {
				rm3, exist := rm2[lk3]
				if !exist {
					dm2[lk3] = lm3
					continue
				}
				if needCorrection {
					if rm3 != lm3 {
						dm2[lk3] = lm3
					}
				}
			}
			if len(dm2) > 0 {
				dm1[lk2] = dm2
			}
		}
		if len(dm1) > 0 {
			dm[lk1] = dm1
		}
	}
	return dm
}
