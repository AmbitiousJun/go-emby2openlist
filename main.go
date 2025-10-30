package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/constant"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/lib/ffmpeg"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist/localtree"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs/colors"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/web"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/web/webport"
	"github.com/gin-gonic/gin"
)

var ginMode = gin.DebugMode

func main() {
	go func() { http.ListenAndServe(":60360", nil) }()

	dataRoot := parseFlag()

	if err := config.ReadFromFile(filepath.Join(dataRoot, "config.yml")); err != nil {
		log.Fatal(err)
	}
	// 初始化阶段检查并自动下载 ffmpeg（只做一次）
	if config.C.Openlist != nil && config.C.Openlist.LocalTreeGen != nil && config.C.Openlist.LocalTreeGen.FFmpegEnable {
		if err := ffmpeg.AutoDownloadExec(dataRoot); err != nil {
			logs.Error("ffmpeg 初始化失败: %v", err)
		} else {
			logs.Info("ffmpeg 环境检测通过 ✓")
		}
	}

	printBanner()

	logs.Info("正在初始化本地目录树模块...")
	if err := localtree.Init(); err != nil {
		log.Fatal(colors.ToRed(err.Error()))
	}

	logs.Info("正在启动服务...")
	gin.SetMode(ginMode)
	if err := web.Listen(); err != nil {
		log.Fatal(colors.ToRed(err.Error()))
	}
}

// parseFlag 转换命令行参数
func parseFlag() (dataRoot string) {
	ph := flag.Int("p", 8095, "HTTP 服务监听端口")
	phs := flag.Int("ps", 8094, "HTTPS 服务监听端口")
	printVersion := flag.Bool("version", false, "查看程序版本")
	dr := flag.String("dr", ".", "程序数据根目录")
	flag.Parse()

	if *printVersion {
		fmt.Println(constant.CurrentVersion)
		os.Exit(0)
	}

	dataRoot = "."
	if *dr != dataRoot {
		stat, err := os.Stat(*dr)
		if err != nil || !stat.IsDir() {
			log.Fatalf("数据根目录 [%s] 不存在", *dr)
		}
		dataRoot = *dr
	}

	if *ph == *phs {
		log.Fatal("HTTP 和 HTTPS 端口冲突")
	}
	webport.HTTP = strconv.Itoa(*ph)
	webport.HTTPS = strconv.Itoa(*phs)
	return
}

func printBanner() {
	fmt.Printf(colors.ToYellow(`
                                 _           ___                        _ _     _   
                                | |         |__ \                      | (_)   | |  
  __ _  ___ ______ ___ _ __ ___ | |__  _   _   ) |___  _ __   ___ _ __ | |_ ___| |_ 
 / _| |/ _ \______/ _ \ '_ | _ \| '_ \| | | | / // _ \| '_ \ / _ \ '_ \| | / __| __|
| (_| | (_) |    |  __/ | | | | | |_) | |_| |/ /| (_) | |_) |  __/ | | | | \__ \ |_ 
 \__, |\___/      \___|_| |_| |_|_.__/ \__, |____\___/| .__/ \___|_| |_|_|_|___/\__|
  __/ |                                 __/ |         | |                           
 |___/                                 |___/          |_|                           

 Repository: %s
    Version: %s
`), constant.RepoAddr, constant.CurrentVersion)
}
