package localtree

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs/colors"
)

// DirName 存放目录树的本地目录名称
const DirName = "openlist-local-tree"

// synchronizer 全局唯一的同步器实例
var synchronizer *Synchronizer

// Init 根据配置文件, 初始化本地目录树
func Init() error {
	// 判断配置是否开启
	if !config.C.Openlist.LocalTreeGen.Enable {
		return nil
	}

	if synchronizer != nil {
		return fmt.Errorf("不可重复初始化")
	}

	dirAbs := filepath.Join(config.BasePath, DirName)

	synchronizer = NewSynchronizer(dirAbs, 30)
	go startSync(synchronizer)

	return nil
}

// doSync 执行一次同步 并输出日志
func doSync(s *Synchronizer, prefix string, apiRefreshFlag bool) error {
	if prefix == "" {
		prefix = "/"
	}

	if prefix != "/" {
		logf(colors.Blue, "开始同步..., 路径前缀: [%s]", prefix)
	} else {
		logf(colors.Blue, "开始全量同步...")
	}

	start := time.Now()
	total, added, deleted, err := s.Sync(prefix, apiRefreshFlag)
	if err != nil {
		return err
	}

	logf(colors.Green, "同步完成, 总数: %d, 新增: %d, 删除: %d, 耗时: %v", total, added, deleted, time.Since(start))
	return nil
}

// startSync 立即同步一次目录树, 并开始定时扫描同步变更
func startSync(s *Synchronizer) {
	if err := doSync(s, "", false); err != nil {
		logf(colors.Red, "同步失败: %v", err)
	}

	d := time.Minute * time.Duration(config.C.Openlist.LocalTreeGen.RefreshInterval)
	timer := time.NewTicker(d)
	for range timer.C {
		if err := doSync(s, "", false); err != nil {
			logf(colors.Red, "同步失败: %v", err)
		}
	}
}

// logf 带上前缀的日志输出
func logf(c colors.C, format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	log.Println(colors.WrapColor(c, "[openlist 目录树]: "+s))
}
