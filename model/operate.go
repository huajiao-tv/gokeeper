package model

import (
	"os"
	"path/filepath"
)

type Opcode string

const (
	OpcodeGet    Opcode = "get"
	OpcodeAdd    Opcode = "add"
	OpcodeUpdate Opcode = "update"
	OpcodeDelete Opcode = "delete"

	DefaultSection = "DEFAULT"
	PathSeparator  = string(os.PathSeparator)
)

type Operate struct {
	Opcode  Opcode
	Domain  string
	File    string
	Section string
	Key     string
	Type    string
	Value   string
	Note    string
	ID      int
	Version int64 //@todo 后台传递回滚版本号
}

func (op *Operate) Format() {
	op.File = filepath.Join(PathSeparator, op.File)
	if op.Section == "" {
		op.Section = DefaultSection
	}
}
