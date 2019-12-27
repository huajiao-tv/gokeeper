package models

import (
	"errors"
	"fmt"
	"strings"

	km "github.com/huajiao-tv/gokeeper/model"
	"github.com/qmessenger/utility/go-ini/ini"
)

func ParseConfigFile(domain, file, content string) ([]*km.Operate, error) {
	cfg, err := ini.Load([]byte(content))
	if err != nil {
		return nil, err
	}

	var cfgOps []*km.Operate
	for _, section := range cfg.Sections() {
		for _, key := range section.Keys() {
			parts := strings.SplitN(key.Name(), " ", 2)
			var name, typ string
			switch len(parts) {
			case 1:
				name = strings.TrimSpace(parts[0])
				typ = "string"
			case 2:
				name = strings.TrimSpace(parts[0])
				typ = strings.TrimSpace(parts[1])
			default:
				msg := fmt.Sprintf("invalid config, section: %s, key: %s", section.Name(), key.Name())
				return nil, errors.New(msg)

			}

			cfgOps = append(cfgOps, &km.Operate{
				Opcode:  AddConfig,
				Domain:  domain,
				File:    file,
				Section: section.Name(),
				Key:     name,
				Type:    typ,
				Value:   key.Value(),
				Note:    "batch add",
			})
		}
	}

	return cfgOps, nil
}
