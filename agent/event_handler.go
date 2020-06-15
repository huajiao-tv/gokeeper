package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	gokeeper "github.com/huajiao-tv/gokeeper/client"
	"github.com/huajiao-tv/gokeeper/model"
)

//
var (
	DatetimeFormat = "20060102150405"
	ErrEventData   = errors.New("event data invalid")
	ErrPidEmpty    = errors.New("pid file empty")
)

func init() {
	gokeeper.EventCallback.RegisterCallFunc(model.EventCmdStop, serviceStop)
	gokeeper.EventCallback.RegisterCallFunc(model.EventCmdStart, serviceStart)
	gokeeper.EventCallback.RegisterCallFunc(model.EventCmdRestart, serviceRestart)
}

func serviceStart(c *gokeeper.Client, evt model.Event) error {
	var err error
	node, ok := (evt.Data).(model.NodeInfo)
	if !ok {
		Logex.Error("serviceStart", "evtData invalid", fmt.Sprintf("%#v", evt))
		return ErrEventData
	}

	Logex.Trace("serviceStart", "receive start event", node.GetComponent())

	stdoutFile := fmt.Sprintf("%s/panic-%s-%s.log", filepath.Join(ComponentLogPath), node.GetComponent(), node.GetID())
	stderrFile := stdoutFile

	// backup old file
	os.Rename(stdoutFile, fmt.Sprintf("%s.%s", stdoutFile, time.Now().Format(DatetimeFormat)))
	os.Rename(stderrFile, fmt.Sprintf("%s.%s", stderrFile, time.Now().Format(DatetimeFormat)))

	stdout, err := os.OpenFile(stdoutFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		Logex.Error("serviceStart", "OpenFile", node.GetComponent(), stdoutFile, err.Error())
		return err
	}
	stderr, err := os.OpenFile(stderrFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		Logex.Error("serviceStart", "OpenFile", node.GetComponent(), stdoutFile, err.Error())
		return err
	}

	command := fmt.Sprintf("%s/%s -d=%s -k=%s -n=%s", filepath.Join(ComponentBinPath), node.GetComponent(), node.GetDomain(), node.GetKeeperAddr(), node.GetID())
	cmd := exec.Command(os.Getenv("SHELL"), "-c", command)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	Logex.Trace("serviceStart", "command", node.GetComponent(), command, "stdout>"+stdoutFile, "stderr>"+stderrFile)

	if err := cmd.Start(); err != nil {
		Logex.Error("serviceStart", "cmd.Start", node.GetComponent(), command, err.Error())
		return fmt.Errorf("%s: %s", err.Error(), command)
	}

	go func(node model.NodeInfo) {
		cmd.Wait()
		stdout.Close()
		stderr.Close()

		Logex.Trace("serviceStart", "component has exited", node.GetComponent())
	}(node)

	err = saveComponentPid(node.GetComponent(), node.GetID(), strconv.Itoa(cmd.Process.Pid))
	if err != nil {
		Logex.Error("serviceStart", "saveComponentPid", node.GetComponent(), err.Error())
		return err
	}

	Logex.Trace("serviceStart", "component has been started", "pid="+strconv.Itoa(cmd.Process.Pid))

	return nil
}

func serviceRestart(c *gokeeper.Client, evt model.Event) error {
	Logex.Trace("serviceRestart", "receive restart event")

	serviceStop(c, evt)
	time.Sleep(1 * time.Second)
	if err := serviceStart(c, evt); err != nil {
		return err
	}
	return nil
}

func serviceStop(c *gokeeper.Client, evt model.Event) error {
	node, ok := (evt.Data).(model.NodeInfo)
	if !ok {
		return ErrEventData
	}

	Logex.Trace("serviceStop", "receive stop event", node.GetComponent())

	pidStr, err := getComponentPid(node.GetComponent(), node.GetID())
	if err != nil {
		Logex.Error("serviceStop", "getComponentPid", node.GetComponent(), err.Error())
		return err
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		Logex.Error("serviceStop", "pid invalid", node.GetComponent(), "pid="+pidStr, err.Error())
		return err
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		Logex.Error("serviceStop", "FindProcess", node.GetComponent(), pid, err.Error())
		return err
	}

	process.Signal(syscall.SIGTERM)
	if err := process.Signal(syscall.SIGUSR1); err != nil {
		Logex.Error("serviceStop", "Signal", node.GetComponent(), err.Error())
		return err
	}

	Logex.Trace("serviceStop", "component should have stopped", node.GetComponent(), pid)

	return nil
}

func getComponentPid(component, nodeID string) (string, error) {
	pidFile := getComponentPidFile(component, nodeID)
	pid, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return "", err
	}
	if len(pid) == 0 {
		return "", ErrPidEmpty
	}
	return string(pid), nil
}

func saveComponentPid(component, nodeID, pid string) error {
	pidFile := getComponentPidFile(component, nodeID)
	err := ioutil.WriteFile(pidFile, []byte(pid), 0744)
	return err
}

func getComponentPidFile(component, nodeID string) string {
	filename := fmt.Sprintf("%s-%s.pid", component, nodeID)
	pidFile := filepath.Join(ComponentPidPath, filename)
	return pidFile
}
