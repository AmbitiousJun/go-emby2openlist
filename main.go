package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/constant"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist/localtree"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs/colors"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/web"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/web/webport"
	"github.com/gin-gonic/gin"
)

func main() {
	parseFlag()

	if err := config.ReadFromFile("config.yml"); err != nil {
		log.Fatal(err)
	}

	printBanner()

	logs.Info("正在初始化本地目录树模块...")
	if err := localtree.Init(); err != nil {
		log.Fatal(colors.ToRed(err.Error()))
	}

	logs.Info("正在启动服务...")
	if err := web.Listen(); err != nil {
		log.Fatal(colors.ToRed(err.Error()))
	}
}

// parseFlag 转换命令行参数
func parseFlag() {
	ph := flag.Int("port-http", 8095, "HTTP 服务监听端口")
	phs := flag.Int("port-https", 8094, "HTTPS 服务监听端口")
	printVersion := flag.Bool("version", false, "查看程序版本")
	prod := flag.Bool("prod", false, "是否以生产环境运行")
	flag.Parse()

	if *printVersion {
		fmt.Println(constant.CurrentVersion)
		os.Exit(0)
	}

	if *prod {
		gin.SetMode(gin.ReleaseMode)
	}

	if *ph == *phs {
		log.Fatal("HTTP 和 HTTPS 端口冲突")
	}
	webport.HTTP = strconv.Itoa(*ph)
	webport.HTTPS = strconv.Itoa(*phs)
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
