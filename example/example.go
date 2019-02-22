package main

import (
	"fmt"
	"logging"
	// "time"
)

func initLogger(logPath, logName, level string) (err error) {
	m := make(map[string]string, 8)
	m["logPath"] = logPath
	m["logName"] = logName
	m["logLevel"] = level
	m["logSplitType"] = "size"
	err = logging.InitLogger("file", m)
	if err != nil {
		return
	}
	logging.Info("init logger success")
	return
}

func run() {
	for {
		logging.Info("user server is running")
		// time.Sleep(time.Second)
	}
}

func main() {
	err := initLogger("./", "user_server", "debug")
	if err != nil {
		fmt.Printf("init logger error:%v\n", err)
	}
	run()
	return
}
