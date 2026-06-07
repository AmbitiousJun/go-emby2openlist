package logs

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Logger interface {
	Log(content string)
}

type defaultLogger struct{}

func (dl *defaultLogger) Log(content string) {
	fmt.Print(content)
}

// DefaultLogger 默认的日志实例
var DefaultLogger Logger = &defaultLogger{}

// otherLoggers 存放其他地方注册进来的日志输出实例
var otherLoggers = sync.Map{}

// RegisterLogger 注册自定义日志输出实例
func RegisterLogger(logger Logger) (id string, ok bool) {
	if logger == nil {
		return "", false
	}

	id = uuid.NewString()
	otherLoggers.Store(id, logger)

	return id, true
}

// RemoveLogger 移除自定义的日志输出实例
func RemoveLogger(id string) {
	otherLoggers.Delete(id)
}
