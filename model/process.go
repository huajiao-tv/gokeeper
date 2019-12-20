package model

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type ProcInfo struct {
	Base *ProcBase
	Cpu  *ProcCpu
	Mem  *ProcMem
}

func init() {
	gob.Register(&ProcInfo{})
}

func NewProcInfo() *ProcInfo {
	return &ProcInfo{}
}

func (this *ProcInfo) Init(pid string) {
	this.Base = &ProcBase{}
	this.Cpu = &ProcCpu{}
	this.Mem = &ProcMem{}
	this.Base.Pid = pid
	this.Base.GetProcInfo()
	this.Mem.Pid = pid
	this.Cpu.Pid = pid
}

func (this *ProcInfo) StartDate() string {
	machine := NewMachineInfo()
	machine.Cpu.Refresh()
	startSeconds := machine.Cpu.Uptime + this.Cpu.StartTime/100
	this.Base.StartTime = time.Unix(int64(startSeconds), 0).Format("2006-01-02 15:04:05")
	return this.Base.StartTime
}

type ProcMem struct {
	VmSize int //Virtual memory size
	VmRss  int //Resident set size
	VmData int //Size of data
	VmStk  int //Size of Stack
	VmExe  int //Size of text segments
	VmLib  int //Shared library code size
	ProcBase
}

func (this *ProcMem) Refresh() error {
	file, err := os.Open("/proc/" + this.Pid + "/status")
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		if strings.Trim(fields[1], " ") == "0" {
			continue
		}
		switch strings.Trim(fields[0], ":") {
		case "VmSize":
			this.VmSize, _ = strconv.Atoi(fields[1])
		case "VmRSS":
			this.VmRss, _ = strconv.Atoi(fields[1])
		case "VmData":
			this.VmData, _ = strconv.Atoi(fields[1])
		case "VmStk":
			this.VmStk, _ = strconv.Atoi(fields[1])
		case "VmExe":
			this.VmExe, _ = strconv.Atoi(fields[1])
		case "VmLib":
			this.VmLib, _ = strconv.Atoi(fields[1])
		}
	}
	return nil
}

func (this *ProcMem) ReSet() {
	this.VmData = 0
	this.VmSize = 0
	this.VmRss = 0
	this.VmLib = 0
	this.VmExe = 0
	this.VmStk = 0
}

func (this *ProcMem) String() string {
	return fmt.Sprintf("VIRT:%d KB, RES:%d KB, Data:%d KB, Stack:%d KB, Text Segment:%d KB, Lib:%d KB",
		this.VmSize, this.VmRss, this.VmData, this.VmStk, this.VmExe, this.VmLib)
}

type ProcCpu struct {
	Utime     uint64
	Stime     uint64
	Cutime    uint64
	Cstime    uint64
	StartTime uint64
	LastUS    uint64    //Utime+Stime
	LastTimer time.Time //time.Now()
	CpuUsage  string
	ProcBase
}

func (this *ProcCpu) Refresh() error {
	file, err := os.Open("/proc/" + this.Pid + "/stat")
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil
	}
	fields := strings.Fields(line)
	if utime, err := strconv.ParseUint(fields[13], 10, 64); err == nil {
		this.Utime = utime
	}
	if stime, err := strconv.ParseUint(fields[14], 10, 64); err == nil {
		this.Stime = stime
	}
	if cutime, err := strconv.ParseUint(fields[15], 10, 64); err == nil {
		this.Cutime = cutime
	}
	if cstime, err := strconv.ParseUint(fields[16], 10, 64); err == nil {
		this.Cstime = cstime
	}
	if starttime, err := strconv.ParseUint(fields[21], 10, 64); err == nil {
		this.StartTime = starttime
	}
	return nil
}

func (this *ProcCpu) ReSet() {
	this.Utime = 0
	this.Stime = 0
	this.Cutime = 0
	this.Cstime = 0
	this.StartTime = 0
}

/**
 * 采样时间段内的cpu使用率,算法与top命令一致。
 * top计算进程cpu使用率源码参见:procs的top.c:prochlp()
 */
func (this *ProcCpu) CurrentUsage() float64 {
	machine := NewMachineInfo()
	nowTime := time.Now()
	totalTime := this.Utime + this.Stime
	sub := nowTime.Sub(this.LastTimer).Seconds()
	sec := 100 / (float64(machine.Hertz) * sub)
	pcpu := float64(totalTime) - float64(this.LastUS)
	this.LastUS = totalTime
	this.LastTimer = nowTime
	this.CpuUsage = fmt.Sprintf("%0.2f", pcpu*sec)
	return pcpu * sec
}

func (this *ProcCpu) String() string {
	return fmt.Sprintf("Cpu:%0.2f%%", this.CurrentUsage())
}

type ProcBase struct {
	Pid       string
	PPid      string
	Command   string
	State     string
	StartTime string
}

func (this *ProcBase) GetProcInfo() error {
	file, err := os.Open("/proc/" + this.Pid + "/stat")
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil
	}
	fields := strings.Fields(line)
	this.PPid = fields[3]
	this.Command = this.GetCommand()
	this.State = fields[2]
	return nil
}

func (this *ProcBase) GetCommand() string {
	command, _ := ioutil.ReadFile("/proc/" + this.Pid + "/cmdline")
	return string(command)
}

type MachineCpu struct {
	User        uint64
	Nice        uint64
	System      uint64
	Idle        uint64
	Iowait      uint64
	Irq         uint64
	SoftIrq     uint64
	Stealstolen uint64
	Guest       uint64
	Uptime      uint64
}

func (this *MachineCpu) Refresh() error {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err == io.EOF {
			return nil
		}
		fields := strings.Fields(line)
		switch fields[0] {
		case "cpu":
			lens := len(fields)
			for i := 1; i < lens; i++ {
				if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					switch i {
					case 1:
						this.User = val
					case 2:
						this.Nice = val
					case 3:
						this.System = val
					case 4:
						this.Idle = val
					case 5:
						this.Iowait = val
					case 6:
						this.Irq = val
					case 7:
						this.SoftIrq = val
					case 8:
						this.Stealstolen = val
					case 9:
						this.Guest = val
					}
				}
			}
		case "btime":
			if uptime, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				this.Uptime = uptime
			}
		}
	}

	return nil
}

func (this *MachineCpu) ReSet() {
	this.User = 0
	this.Nice = 0
	this.System = 0
	this.Idle = 0
	this.Iowait = 0
	this.Irq = 0
	this.SoftIrq = 0
	this.Stealstolen = 0
	this.Guest = 0
}

type MachineInfo struct {
	Uptime float64
	Hertz  int
	Host   string
	Cpu    *MachineCpu
}

/**
 * 获取Hertz值:C.sysconf(C._SC_CLK_TCK),sysconf具体实现参见glibc源码:
 * x86_64平台:glibc/sysdeps/posix/sysconf.c
 * glibc/sysdeps/unix/sysv/linux/getclktck.c
 */
func NewMachineInfo() *MachineInfo {
	return &MachineInfo{Hertz: 100, Cpu: &MachineCpu{}}
}

func (this *MachineInfo) GetUptime() float64 {
	if uptime, err := ioutil.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(uptime))
		this.Uptime, _ = strconv.ParseFloat(fields[0], 64)
	}
	return this.Uptime
}
