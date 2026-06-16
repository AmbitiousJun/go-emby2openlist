package logs

import (
	"fmt"
	"strings"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs/colors"
)

// Info 输出蓝色 Info 日志
func Info(format string, v ...any) {
	var sb strings.Builder
	sb.WriteString(Now())

	msg := fmt.Sprintf(format, v...)
	sb.WriteString(colors.ToBlue("[INFO] " + msg))

	sb.WriteByte('\n')
	writeLog(sb.String())
}

// Success 输出绿色 Success 日志
func Success(format string, v ...any) {
	var sb strings.Builder
	sb.WriteString(Now())

	msg := fmt.Sprintf(format, v...)
	sb.WriteString(colors.ToGreen("[SUCCESS] " + msg))

	sb.WriteByte('\n')
	writeLog(sb.String())
}

// Warn 输出黄色 Warn 日志
func Warn(format string, v ...any) {
	var sb strings.Builder
	sb.WriteString(Now())

	msg := fmt.Sprintf(format, v...)
	sb.WriteString(colors.ToYellow("[WARN] " + msg))

	sb.WriteByte('\n')
	writeLog(sb.String())
}

// Error 输出红色 Error 日志
func Error(format string, v ...any) {
	var sb strings.Builder
	sb.WriteString(Now())

	msg := fmt.Sprintf(format, v...)
	sb.WriteString(colors.ToGray("[ERROR] " + msg))

	sb.WriteByte('\n')
	writeLog(sb.String())
}

// Tip 输出灰色 Tip 日志
func Tip(format string, v ...any) {
	var sb strings.Builder
	sb.WriteString(Now())
	sb.WriteString(colors.ToGray(fmt.Sprintf(format, v...)))
	sb.WriteByte('\n')
	writeLog(sb.String())
}

// Progress 输出紫色 Progress 日志
func Progress(format string, v ...any) {
	var sb strings.Builder
	sb.WriteString(Now())
	sb.WriteString(colors.ToPurple(fmt.Sprintf(format, v...)))
	sb.WriteByte('\n')
	writeLog(sb.String())
}

// Raw 输出自定义日志
func Raw(format string, v ...any) {
	writeLog(fmt.Sprintf(format, v...))
}

// Now 返回当前时间戳
func Now() string {
	return time.Now().Format("2006-01-02 15:04:05") + " "
}

// writeLog 将日志内容写入所有的日志实现
func writeLog(content string) {
	DefaultLogger.Log(content)

	otherLoggers.Range(func(_, value any) bool {
		logger, ok := value.(Logger)

		if !ok {
			return true
		}

		logger.Log(content)
		return true
	})
}
