package logging

import (
	"testing"
)

func TestFileLogger(t *testing.T) {
	m := make(map[string]string, 8)
	m["log_path"] = "./logs/"
	m["log_name"] = "server"
	m["log_level"] = "debug"
	logger ,err := NewFileLogger(m)
	if err != nil {
		return
	}
	logger.Debug("user debug")
	logger.Info("user info")
	logger.Warn("user warn")
	logger.Error("user error")
	logger.Fatal("user fatal")
}
