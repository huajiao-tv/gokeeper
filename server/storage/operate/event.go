package operate

type Opcode string

const (
	OpcodeUpdateKey    Opcode = "updateKey"
	OpcodeDeleteKey    Opcode = "deleteKey"
	OpcodeRollback     Opcode = "rollback"
	OpcodeUpdateDomain Opcode = "updateDomain"
)

type EventModeType string

const (
	EventModeConf    EventModeType = "conf"
	EventModeVersion EventModeType = "version"
)

type Event struct {
	Opcode  Opcode
	Domain  string
	File    string
	Section string
	Key     string
	Data    interface{}
}

func IsValidEventMode(mode string) bool {
	modes := []string{string(EventModeConf), string(EventModeVersion)}
	for _, m := range modes {
		if mode == m {
			return true
		}
	}
	return false
}
