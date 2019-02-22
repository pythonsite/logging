package logging

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type FileLogger struct {
	level         int
	logPath       string
	logName       string
	file          *os.File
	warnFile      *os.File		// 针对Error,Warn,Fatal的日志单独写一个文件中
	LogDataChan   chan *LogData
	logSplitType  int		// 日志切割的方式，有以小时和以大小两种方式
	logSplitSize  int64
	lastSplitHour int
}

func NewFileLogger(config map[string]string) (log LogInterface, err error) {
	logPath, ok := config["logPath"]
	if !ok {
		err = fmt.Errorf("not found logPath config")
		return
	}
	logName, ok := config["logName"]
	if !ok {
		err = fmt.Errorf("not found logName config")
		return
	}
	logLevel, ok := config["logLevel"]
	if !ok {
		err = fmt.Errorf("not found logLevel config")
		return
	}
	level := getLevel(logLevel)

	logChanSize, ok := config["logChanSize"]
	if !ok {
		logChanSize = "50000"
	}
	chanSize, err := strconv.Atoi(logChanSize)
	if err != nil {
		chanSize = 50000
	}

	var logSplitType int = LogSplitTypeHour
	var logSplitSize int64
	logSplitStr, ok := config["logSplitType"]
	if !ok {
		logSplitStr = "hour"
	} else {
		if logSplitStr == "size" {
			logSplitSizeStr, ok := config["logSplitSize"]
			if !ok {
				logSplitSizeStr = "104857600"
			}
			logSplitSize, err = strconv.ParseInt(logSplitSizeStr, 10, 64)
			if err != nil {
				logSplitSize = 104857600
			}
			logSplitType = LogSplitTypeSize
		} else {
			logSplitType = LogSplitTypeHour
		}
	}
	log = &FileLogger{
		level:         level,
		logPath:       logPath,
		logName:       logName,
		LogDataChan:   make(chan *LogData, chanSize),
		logSplitSize:  logSplitSize,
		logSplitType:  logSplitType,
		lastSplitHour: time.Now().Hour(),
	}
	log.Init()
	return
}

func (f *FileLogger) Init() {
	filename := fmt.Sprintf("%s/%s.log", f.logPath, f.logName)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Sprintf("open file %s failed, err:%v", filename, err))
	}
	f.file = file
	// 写错误日志和Fatal日志的文件
	filename = fmt.Sprintf("%s/%s.wf.log", f.logPath, f.logName)
	file, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		panic(fmt.Sprintf("open file %s failed, err:%v", filename, err))
	}
	f.warnFile = file
	// 在初始化的时候单独用一个线程从channel中获取日志内容并写入到文件中
	go f.writeLogBackground()
}

func (f *FileLogger) splitFileHour(warnFile bool) {
	now := time.Now()
	hour := now.Hour()
	if f.lastSplitHour == hour {
		return
	}

	var backupFileName string
	var oldFileName string
	if warnFile {
		backupFileName = fmt.Sprintf("%s/%s.wf.log_%04d%02d%02d%02d",
			f.logPath, f.logName, now.Year(), now.Month(), now.Day(), f.lastSplitHour)
		oldFileName = fmt.Sprintf("%s/%s.wf.log", f.logPath, f.logName)

	} else {
		backupFileName = fmt.Sprintf("%s/%s.log_%04d%02d%02d%02d",
			f.logPath, f.logName, now.Year(), now.Month(), now.Day(), f.lastSplitHour)
		oldFileName = fmt.Sprintf("%s/%s.log", f.logPath, f.logName)
	}
	file := f.file
	if warnFile {
		file = f.warnFile
	}
	f.Close()

	os.Rename(oldFileName, backupFileName)
	file, err := os.OpenFile(oldFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}
	if warnFile {
		f.warnFile = file
	} else {
		f.file = file
	}
	f.lastSplitHour = hour
}

func (f *FileLogger) splitFileSize(warnFile bool) {
	file := f.file
	if warnFile {
		file = f.warnFile
	}
	statInfo, err := file.Stat()
	if err != nil {
		return
	}
	fileSize := statInfo.Size()
	if fileSize <= f.logSplitSize {
		return
	}

	var backupFileName string
	var oldFileName string

	now := time.Now()
	if warnFile {
		backupFileName = fmt.Sprintf("%s/%s.wf.log_%04d%02d%02d%02d%02d%02d",
			f.logPath, f.logName, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
		oldFileName = fmt.Sprintf("%s/%s.wf.log", f.logPath, f.logName)

	} else {
		backupFileName = fmt.Sprintf("%s/%s.log_%04d%02d%02d%02d%02d%02d",
			f.logPath, f.logName, now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
		oldFileName = fmt.Sprintf("%s/%s.log", f.logPath, f.logName)
	}
	f.Close()

	os.Rename(oldFileName, backupFileName)
	file, err = os.OpenFile(oldFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return
	}
	if warnFile {
		f.warnFile = file
	} else {
		f.file = file
	}
}

func (f *FileLogger) checkSplitFile(warnFile bool) {
	if f.logSplitType == LogSplitTypeHour {
		f.splitFileHour(warnFile)
		return
	}
	f.splitFileSize(warnFile)
}

// 单独的线程将日志从channel中写入到文件
func (f *FileLogger) writeLogBackground() {
	for data := range f.LogDataChan {
		var file *os.File = f.file
		if data.WarnAndFatal {
			file = f.warnFile
		}
		f.checkSplitFile(data.WarnAndFatal)
		fmt.Fprintf(file, "[%s] [%s] [%s:%s:%d] %s\n",
			data.TimeStr, data.LevelStr, data.FileName, data.FuncName, data.LineNo, data.Message)
	}
}

func (f *FileLogger) SetLevel(level int) {
	if level < LogLevelDebug || level > LogLevelFatal {
		level = LogLevelDebug
	}
	f.level = level
}

func (f *FileLogger) Debug(format string, args ...interface{}) {
	if f.level > LogLevelDebug {
		return
	}
	logData := writeLog(LogLevelDebug, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Trace(format string, args ...interface{}) {
	if f.level > LogLevelTrace {
		return
	}
	logData := writeLog(LogLevelTrace, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Info(format string, args ...interface{}) {
	if f.level > LogLevelInfo {
		return
	}
	logData := writeLog(LogLevelInfo, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Fatal(format string, args ...interface{}) {
	if f.level > LogLevelFatal {
		return
	}
	logData := writeLog(LogLevelFatal, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Error(format string, args ...interface{}) {
	if f.level > LogLevelError {
		return
	}
	logData := writeLog(LogLevelError, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Warn(format string, args ...interface{}) {
	if f.level > LogLevelWarn {
		return
	}
	logData := writeLog(LogLevelWarn, format, args...)
	select {
	case f.LogDataChan <- logData:
	default:
	}
}

func (f *FileLogger) Close() {
	f.file.Close()
	f.warnFile.Close()
}
