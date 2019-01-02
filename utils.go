package logging

import (
	"runtime"
	"time"
	"path"
	"fmt"
)

type LogData struct {
	Message string
	TimeStr string
	LevelStr string
	FileName string
	FuncName string
	LineNo int
	WarnAndFatal bool	// 是否是Warn和Fatal类型的日志
}

func GetLineInfo()(fileName string, funcName string, lineNo int) {
	pc, file, line, ok := runtime.Caller(4)
	if ok {
		fileName = file
		funcName = runtime.FuncForPC(pc).Name()
		lineNo = line
	}
	return
}

func writeLog(level int, format string, args...interface{})*LogData{
	now := time.Now()
	nowStr := now.Format("2006-01-02 15:04:05.999")
	levelStr := getLevelText(level)
	fileName , funcName, lineNo := GetLineInfo()
	fileName = path.Base(fileName)
	funcName = path.Base(funcName)
	msg := fmt.Sprintf(format, args...)
	LogData := &LogData{
		Message: msg,
		TimeStr: nowStr,
		LevelStr: levelStr,
		FileName: fileName,
		FuncName: funcName,
		LineNo: lineNo,
		WarnAndFatal: false,
	}
	if level == LogLevelError || level == LogLevelFatal || level == LogLevelWarn {
		LogData.WarnAndFatal = true
	}
	return LogData
	// fmt.Fprintf(file, "[%s] [%s] [%s:%s:%d] %s\n", nowStr, levelStr, fileName, funcName, lineNo, msg)

}