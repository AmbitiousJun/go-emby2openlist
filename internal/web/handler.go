package web

import (
	"io/fs"
	"net/http"
	"regexp"
	"strings"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/constant"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	web_static "github.com/AmbitiousJun/go-emby2openlist/v2/web"
	"github.com/gin-gonic/gin"
)

// MatchRouteKey 存储在 gin 上下文的路由匹配字段
const MatchRouteKey = "matchRoute"

// globalDftHandler 全局默认兜底的请求处理器
func globalDftHandler(c *gin.Context) {
	if c.Request.Method == http.MethodHead {
		c.String(http.StatusOK, "")
		return
	}

	if handleWebStatic(c) {
		return
	}

	// 依次匹配路由规则, 找到其他的处理器
	for _, rule := range rules {
		reg := rule[0].(*regexp.Regexp)
		if reg.MatchString(c.Request.RequestURI) {
			c.Set(MatchRouteKey, reg.String())
			c.Set(constant.RouteSubMatchGinKey, reg.FindStringSubmatch(c.Request.RequestURI))
			rule[1].(gin.HandlerFunc)(c)
			return
		}
	}
}

// handleWebStatic 处理 web 静态资源
func handleWebStatic(c *gin.Context) (ok bool) {
	if !config.C.Ge2o.Web.IsEnabled() {
		return false
	}

	path := c.Request.URL.Path

	if strings.TrimRight(path, "/") == constant.Route_SelfBase {
		c.Redirect(http.StatusMovedPermanently, constant.Route_Web+"/")
		return true
	}

	feFS, err := fs.Sub(web_static.EmbedFS, "dist")
	if err != nil {
		logs.Error("获取静态资源文件系统失败: %v", err)
		return false
	}

	if !strings.HasPrefix(path, constant.Route_Web) {
		return false
	}

	serveIndexHtml := func() {
		data, err := fs.ReadFile(feFS, "index.html")
		if err != nil {
			logs.Error("读取 index.html 失败: %v", err)
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}

	filePath := strings.TrimPrefix(path, constant.Route_Web)
	filePath = strings.TrimPrefix(filePath, "/")

	if filePath == "" {
		serveIndexHtml()
		return true
	}

	// 检查文件是否存在
	file, err := feFS.Open(filePath)
	if err != nil {
		serveIndexHtml()
		return true
	}
	file.Close()

	// 文件存在 直接返回相应的静态资源
	c.FileFromFS("/"+filePath, http.FS(feFS))
	return true
}

// compileRules 编译路由的正则表达式
func compileRules(rs [][2]any) [][2]any {
	newRs := make([][2]any, 0, len(rs))
	for _, rule := range rs {
		reg, err := regexp.Compile(rule[0].(string))
		if err != nil {
			logs.Error("路由正则编译失败, pattern: %v, error: %v", rule[0], err)
			continue
		}
		rule[0] = reg

		rawHandler, ok := rule[1].(func(*gin.Context))
		if !ok {
			logs.Error("错误的请求处理器, pattern: %v", rule[0])
			continue
		}
		var handler gin.HandlerFunc = rawHandler
		rule[1] = handler
		newRs = append(newRs, rule)
	}
	return newRs
}
