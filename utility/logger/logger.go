package logger

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DATEFORMAT = "2006-01-02"
	HOURFORMAT = "2006010215"
)

var logLevel = [4]string{"debug", "trace", "warn", "error"}

/*
 * 默认文件包括debug/error/trace/warn
 */
type Logger struct {
	logMap     map[string]*LoggerInfo
	suffixInfo string
	logLevel   int // 需要记录的日志级别
	sync.RWMutex
}

type LoggerInfo struct {
	filename       string
	bufferInfoLock sync.RWMutex
	buffer         *LoggerBuffer
	bufferQueue    chan LoggerBuffer
	fsyncInterval  time.Duration
	hour           time.Time
	fileOrder      int
	logFile        *os.File
	backupDir      string
}

const (
	_        = iota
	KB int64 = 1 << (iota * 10)
	MB
	GB
	TB
	maxFileSize       = 2 * GB
	maxFileCount      = 10
	defaultBufferSize = 2 * KB
)

type LoggerBuffer struct {
	bufferLock    sync.RWMutex
	bufferContent *bytes.Buffer
}

/* logger 新版本 */
func NewLogger(filename, suffix, backupDir string) (*Logger, error) {
	var err error
	var loggerInfo *LoggerInfo
	logMap := make(map[string]*LoggerInfo)
	for _, level := range logLevel {
		if loggerInfo, err = NewLoggerInfo(filename, level); err != nil {
			return nil, err
		}
		loggerInfo.backupDir, _ = filepath.Abs(backupDir)
		go loggerInfo.WriteBufferToQueue()
		go loggerInfo.FlushBufferQueue()
		logMap[level] = loggerInfo
	}

	logger := &Logger{logMap: logMap, suffixInfo: suffix}
	return logger, nil
}

/*
 * 写日志类，根据filename重新创建一个LoggerInfo，主要是针对自定义文件
 * 参数 filename-文件名；suffix-是否需要后缀信息；args-写入的内容
 */
func (this *Logger) Write(filename string, suffix bool, args ...interface{}) {

	var loggerInfo *LoggerInfo
	var err error
	var Ok bool
	/*不存在需要重新初始化一下*/
	this.Lock()
	defer this.Unlock()
	if loggerInfo, Ok = this.logMap[filename]; !Ok {
		if loggerInfo, err = NewLoggerInfo(filename, ""); err != nil {
			println("[NewLoggerInfo] Write : " + err.Error())
			return
		}
		go loggerInfo.WriteBufferToQueue()
		go loggerInfo.FlushBufferQueue()
		this.logMap[filename] = loggerInfo
	}
	loggerInfo.Write(Format(suffix, this.suffixInfo, args...))
}

// 记录级别，0最低，所有日志都记录，3表示只记录error日志
func (this *Logger) SetLevel(l int) {
	this.Lock()
	defer this.Unlock()
	if l > len(logLevel) {
		this.logLevel = len(logLevel)
	} else {
		this.logLevel = l
	}
}

func (this *Logger) CheckLevel(logType string) bool {
	if this.logLevel <= 0 {
		return true
	}
	logSet := logLevel[this.logLevel:]
	for _, v := range logSet {
		if logType == v {
			return true
		}
	}
	return false
}

/*
 * 以下四个函数主要是写入不同的日志类型
 * 参数，写入的具体内容数组
 */
func (this *Logger) Debug(args ...interface{}) {
	this.RLock()
	loggerInfo := this.logMap["debug"]
	d := this.CheckLevel("debug")
	this.RUnlock()
	if !d {
		return
	}
	loggerInfo.Write(Format(true, this.suffixInfo, args...))
}

func (this *Logger) Trace(args ...interface{}) {
	this.RLock()
	loggerInfo := this.logMap["trace"]
	d := this.CheckLevel("trace")
	this.RUnlock()
	if !d {
		return
	}
	loggerInfo.Write(Format(true, this.suffixInfo, args...))
}

func (this *Logger) Warn(args ...interface{}) {
	this.RLock()
	loggerInfo := this.logMap["warn"]
	d := this.CheckLevel("warn")
	this.RUnlock()
	if !d {
		return
	}
	loggerInfo.Write(Format(true, this.suffixInfo, args...))
}

func (this *Logger) Error(args ...interface{}) {
	this.RLock()
	loggerInfo := this.logMap["error"]
	d := this.CheckLevel("error")
	this.RUnlock()
	if !d {
		return
	}
	loggerInfo.Write(Format(true, this.suffixInfo, args...))
}

func NewLoggerInfo(filename, level string) (*LoggerInfo, error) {
	var err error
	loggerInfo := &LoggerInfo{
		bufferQueue:   make(chan LoggerBuffer, 50000),
		fsyncInterval: time.Second,
		buffer:        NewLoggerBuffer(),
		fileOrder:     0,
		backupDir:     "",
	}

	t, _ := time.Parse(HOURFORMAT, time.Now().Format(HOURFORMAT))
	loggerInfo.hour = t

	// 直接调用write写日志的文件名，用原始的文件名
	if len(level) == 0 {
		loggerInfo.filename, _ = filepath.Abs(filename)
	} else {
		loggerInfo.filename, _ = filepath.Abs(filename + "-" + level + ".log")
	}

	err = loggerInfo.CreateFile()
	if err != nil {
		println("[NewLogger] openfile error : " + err.Error())
		return nil, err
	}
	return loggerInfo, nil
}

/*
 * 获取文件大小，如果文件不存在则重新创建文件
 * 则文件指针指向错误，重新open一下文件
 * 如果有其他的错误，此处无法处理，只能是丢掉部分日志内容
 */
func (this *LoggerInfo) FileSize() (int64, error) {
	if f, err := os.Stat(this.filename); err != nil {
		return 0, err
	} else {
		return f.Size(), nil
	}
}

/*
 * 创建文件
 */
func (this *LoggerInfo) CreateFile() error {
	var err error
	hourStr := time.Now().Format(HOURFORMAT)
	trueFilename := fmt.Sprintf("%v.%v", this.filename, hourStr)
	this.logFile, err = os.OpenFile(trueFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)

	os.Remove(this.filename)
	os.Symlink(trueFilename, this.filename)
	return err
}

/*
 * 判断文件是否需要切分
 */
func (this *LoggerInfo) NeedSplit() (split bool, backup bool) {
	t, _ := time.Parse(HOURFORMAT, time.Now().Format(HOURFORMAT))
	if t.After(this.hour) {
		return false, true
	} else {
		/*
		 * 判断文件大小错误，当做文件不存在，
		 * 重新创建一次文件，只重建一次，如果还有错误，
		 * 只做记录
		 */
		if size, err := this.FileSize(); err != nil {
			if os.IsNotExist(err) {
				/* 文件不存在，重新创建文件 */
				println("[NeedSplit] FileSize: " + err.Error())
				if err = this.CreateFile(); err != nil {
					println("[NeedSplit] CreateFile : " + err.Error())
				}
				return false, false
			} else {
				/* 如果不是文件不存在错误，不做处理*/
				println("[NeedSplit] FileSize: " + err.Error())
				return false, false
			}
		} else {
			if size > maxFileSize {
				return true, false
			}
		}
		return false, false
	}
	return false, false
}

func (this *LoggerInfo) Write(content string) {
	this.bufferInfoLock.Lock()
	this.buffer.WriteString(content)
	this.bufferInfoLock.Unlock()
}

// 只有该函数goroutine对map操作，map没有加锁
func (this *LoggerInfo) WriteBufferToQueue() {
	ticker := time.NewTicker(this.fsyncInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		this.bufferInfoLock.RLock()
		this.buffer.WriteBuffer(this.bufferQueue)
		this.bufferInfoLock.RUnlock()
	}
}

/*
 * buffer中的数据flush到硬盘
 */
func (this *LoggerInfo) FlushBufferQueue() {
	for {
		select {
		case buffer := <-this.bufferQueue:
			/* 需要做文件切分 */
			isSplit, isBackup := this.NeedSplit()
			if isSplit {
				this.logFile.Close()
				if trueFilename, err := os.Readlink(this.filename); err == nil {
					newFilename := this.filename + "." + this.hour.Format(HOURFORMAT) + "." + strconv.Itoa(this.fileOrder%maxFileCount)
					_, fileErr := os.Stat(newFilename)
					if fileErr == nil {
						os.Remove(newFilename)
					}
					err := os.Rename(trueFilename, newFilename)
					if err != nil {
						println("[FlushBufferQueue] Rename : " + err.Error())
					}
				}
				if err := this.CreateFile(); err != nil {
					println("[FlushBufferQueue] CreateFile : " + err.Error())
				}

				this.fileOrder++
				if isBackup {
					this.fileOrder = 0
					go this.LoggerBackup(this.hour)
					this.hour, _ = time.Parse(HOURFORMAT, time.Now().Format(HOURFORMAT))
				}
			} else {
				if isBackup {
					this.logFile.Close()

					if this.fileOrder != 0 {
						if trueFilename, err := os.Readlink(this.filename); err == nil {
							newFilename := this.filename + "." + this.hour.Format(HOURFORMAT) + "." + strconv.Itoa(this.fileOrder%maxFileCount)
							_, fileErr := os.Stat(newFilename)
							if fileErr == nil {
								os.Remove(newFilename)
							}
							err := os.Rename(trueFilename, newFilename)
							if err != nil {
								println("[FlushBufferQueue] Rename : " + err.Error())
							}
						}
					}

					if err := this.CreateFile(); err != nil {
						println("[FlushBufferQueue] CreateFile : " + err.Error())
					}

					this.fileOrder = 0
					go this.LoggerBackup(this.hour)
					this.hour, _ = time.Parse(HOURFORMAT, time.Now().Format(HOURFORMAT))
				}
			}

			/* 写失败的话尝试再写一次 */
			if _, err := this.logFile.Write(buffer.bufferContent.Bytes()); err != nil {
				println("[FlushBufferQueue] File.Write : " + err.Error())
				this.logFile.Write(buffer.bufferContent.Bytes())
			}
			this.logFile.Sync()

		}
	}
}

/*
 * 错误日志备份
 * backupDir 待备份的目录
 * os中没有mv的函数，只能先rename，后remove
 * backupDir -> /data/messenger/servers/log/saver/trace/2014-09-10/*.log
 */
func (this *LoggerInfo) LoggerBackup(hour time.Time) {
	var oldFile string   //待备份文件
	var newFile string   //需要备份的新文件
	var backupDir string //备份的路径

	if this.backupDir == "" {
		return
	}
	backupDir = filepath.Join(this.backupDir, hour.Format(DATEFORMAT))
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		os.MkdirAll(backupDir, 0777)
	}

	/* backup filename like saver-zwt-01-error.log.2014-09-10*/
	oldFile = this.filename + "." + hour.Format(HOURFORMAT)
	if stat, err := os.Stat(oldFile); err == nil {
		newFile = filepath.Join(backupDir, stat.Name())
		if err := os.Rename(oldFile, newFile); err != nil {
			println("[LoggerBackup] os.Rename:" + err.Error())
		}
	}

	/* backup filename like saver-zwt-01-error.log.2014-09-10.{0/1...} */
	for i := 0; i < maxFileCount; i++ {
		oldFile = this.filename + "." + hour.Format(HOURFORMAT) + "." + strconv.Itoa(i)
		if stat, err := os.Stat(oldFile); err == nil {
			newFile = filepath.Join(backupDir, stat.Name())
			if err := os.Rename(oldFile, newFile); err != nil {
				println("[LoggerBackup] os.Rename:" + err.Error())
			}
		}
	}
}

func NewLoggerBuffer() *LoggerBuffer {
	return &LoggerBuffer{
		bufferContent: bytes.NewBuffer(make([]byte, 0, defaultBufferSize)),
	}
}

func (this *LoggerBuffer) WriteString(str string) {
	this.bufferContent.WriteString(str)
}

func (this *LoggerBuffer) WriteBuffer(bufferQueue chan LoggerBuffer) {
	this.bufferLock.Lock()
	if this.bufferContent.Len() > 0 {
		bufferQueue <- *this
		this.bufferContent = bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	}
	this.bufferLock.Unlock()
}

func GetDatetime() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

func Format(suffix bool, suffixInfo string, args ...interface{}) string {
	var content string
	for _, arg := range args {
		switch arg.(type) {
		case int:
			content = content + "|" + strconv.Itoa(arg.(int))
			break
		case string:
			content = content + "|" + strings.TrimRight(arg.(string), "\n")
			break
		case int64:
			str := strconv.FormatInt(arg.(int64), 10)
			content = content + "|" + str
			break
		default:
			content = content + "|" + fmt.Sprintf("%v", arg)
			break
		}
	}
	if suffix {
		content = GetDatetime() + content + "|" + suffixInfo + "\n"
	} else {
		content = GetDatetime() + content + "\n"
	}
	return content
}

func GetInnerIp() string {
	info, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range info {
		ipMask := strings.Split(addr.String(), "/")
		if ipMask[0] != "127.0.0.1" && ipMask[0] != "24" {
			return ipMask[0]
		}
	}
	return ""
}
