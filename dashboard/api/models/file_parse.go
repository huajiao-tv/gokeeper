package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qmessenger/utility/go-ini/ini"
)

func ParseConfigFile(domain, file, content string) ([]*ConfigOperate, error) {
	cfg, err := ini.Load([]byte(content))
	if err != nil {
		return nil, err
	}

	var cfgOps []*ConfigOperate
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

			cfgOps = append(cfgOps, &ConfigOperate{
				Action:  AddConfig,
				Cluster: domain,
				File:    file,
				Section: section.Name(),
				Key:     name,
				Type:    typ,
				Value:   key.Value(),
				Comment: "batch add",
			})
		}
	}

	return cfgOps, nil
}
