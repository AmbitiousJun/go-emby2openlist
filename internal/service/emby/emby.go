package emby

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/bytess"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/jsons"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"

	"github.com/gin-gonic/gin"
)

func ProxySocket() func(*gin.Context) {

	var proxy *httputil.ReverseProxy
	var once = sync.Once{}

	initFunc := func() {
		origin := config.C.Emby.Host
		u, err := url.Parse(origin)
		if err != nil {
			panic("转换 emby host 异常: " + err.Error())
		}

		proxy = httputil.NewSingleHostReverseProxy(u)

		// 禁用系统代理
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = nil
		proxy.Transport = transport

		proxy.Director = func(r *http.Request) {
			r.URL.Scheme = u.Scheme
			r.URL.Host = u.Host
		}
	}

	return func(c *gin.Context) {
		once.Do(initFunc)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// HandleImages 处理图片请求
//
// 修改图片质量参数为配置值
func HandleImages(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Del("quality")
	q.Del("Quality")
	q.Set("Quality", strconv.Itoa(config.C.Emby.ImagesQuality))
	c.Request.RequestURI = c.Request.URL.Path + "?" + q.Encode()
	ProxyOrigin(c)
}

// ProxyOrigin 将请求代理到源服务器
func ProxyOrigin(c *gin.Context) {
	if c == nil {
		return
	}
	origin := config.C.Emby.Host

	// 传递客户端 IP 到 emby
	c.Request.Header.Set("X-Forwarded-For", c.ClientIP())
	c.Request.Header.Set("X-Real-IP", c.ClientIP())

	if err := https.ProxyPass(c.Request, c.Writer, origin); err != nil {
		logs.Error("代理异常: %v", err)
	}
}

// TestProxyUri 用于测试的代理,
// 主要是为了查看实际请求的详细信息, 方便测试
func TestProxyUri(c *gin.Context) bool {
	testUris := []string{}

	flag := false
	for _, uri := range testUris {
		if strings.Contains(c.Request.RequestURI, uri) {
			flag = true
			break
		}
	}
	if !flag {
		return false
	}

	type TestInfos struct {
		Uri        string
		Method     string
		Header     map[string]string
		Body       string
		RespStatus int
		RespHeader map[string]string
		RespBody   string
	}

	infos := &TestInfos{
		Uri:        c.Request.URL.String(),
		Method:     c.Request.Method,
		Header:     make(map[string]string),
		RespHeader: make(map[string]string),
	}

	for key, values := range c.Request.Header {
		infos.Header[key] = strings.Join(values, "|")
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logs.Error("测试 uri 执行异常: %v", err)
		return false
	}
	infos.Body = string(bodyBytes)

	origin := config.C.Emby.Host
	resp, err := https.Request(infos.Method, origin+infos.Uri).
		Header(c.Request.Header).
		Body(io.NopCloser(bytes.NewBuffer(bodyBytes))).
		Do()
	if err != nil {
		logs.Error("测试 uri 执行异常: %v", err)
		return false
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		infos.RespHeader[key] = strings.Join(values, "|")
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		logs.Error("测试 uri 执行异常: %v", err)
		return false
	}
	infos.RespBody = string(bodyBytes)
	infos.RespStatus = resp.StatusCode
	logs.Warn("测试 uri 代理信息: %s", jsons.FromValue(infos))

	c.Status(infos.RespStatus)
	c.Writer.Write(bodyBytes)

	return true
}

// ProxyRoot web 首页代理
func ProxyRoot(c *gin.Context) {
	resp, err := https.Request(c.Request.Method, config.C.Emby.Host+c.Request.URL.String()).
		Header(c.Request.Header).
		Body(c.Request.Body).
		DoSingle()

	if checkErr(c, err) {
		return
	}
	defer resp.Body.Close()

	https.CloneHeader(c.Writer, resp.Header)
	c.Status(resp.StatusCode)

	buf := bytess.CommonFixedBuffer()
	defer buf.PutBack()
	io.CopyBuffer(c.Writer, resp.Body, buf.Bytes())
}
