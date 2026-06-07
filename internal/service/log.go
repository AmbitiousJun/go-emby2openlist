package service

import (
	"net/http"
	"sync/atomic"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/model"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// wsLogger 实现将日志输出到 websocket
type wsLogger struct {
	conn    *websocket.Conn // 连接实例
	channel chan string     // 日志接收通道
	done    chan struct{}   // 用于通知消费协程
	closed  atomic.Bool     // 标记连接是否已经关闭
}

// newWsLogger 初始化一个 ws 日志输出器
func newWsLogger(conn *websocket.Conn) *wsLogger {
	return &wsLogger{
		conn:    conn,
		channel: make(chan string, 100),
		done:    make(chan struct{}),
	}
}

// close 关闭连接
func (l *wsLogger) close() {
	if ok := l.closed.CompareAndSwap(false, true); !ok {
		return
	}

	close(l.done)
	_ = l.conn.Close()
}

// consume 循环阻塞写出日志
func (l *wsLogger) consume() {
	for {
		select {
		case <-l.done:
			return
		case content, ok := <-l.channel:
			if !ok {
				return
			}

			err := l.conn.WriteMessage(websocket.TextMessage, []byte(content))
			if err != nil {
				l.close()
				return
			}
		}
	}
}

// Log 实现 Logger 接口
func (l *wsLogger) Log(content string) {
	select {
	case <-l.done:
	case l.channel <- content:
	default:
	}
}

// logUpgrader 日志输出专用 WebSocket 升级器
var logUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SyncServerLog 建立 WebSocket 连接, 同步服务端日志
func SyncServerLog(c *gin.Context) {
	// 1 校验程序密钥
	secret := c.Query("secret")
	if secret != config.C.Ge2o.ApiSecret {
		c.JSON(http.StatusOK, model.Response{Message: "密钥错误"})
		return
	}

	// 2 升级连接
	conn, err := logUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logs.Error("SyncServerLog 升级 WebSocket 连接失败: %v", err)
		return
	}
	wsLogger := newWsLogger(conn)
	defer wsLogger.close()

	// 3 注册 Logger
	id, ok := logs.RegisterLogger(wsLogger)
	if !ok {
		logs.Error("SyncServerLog 注册 Logger 失败")
		return
	}
	defer logs.RemoveLogger(id)

	// 4 循环消费
	go wsLogger.consume()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
