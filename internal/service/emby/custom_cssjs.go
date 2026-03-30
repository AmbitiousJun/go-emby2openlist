package emby

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/constant"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

const (

	// customUpdateInterval 内存缓存刷新间隔 防止频繁刷新
	customUpdateInterval = 5_000
)

var (
	// customJsList 所有自定义脚本预加载在内存中
	customJsList []string

	// customCssList 所有自定义样式预加载在内存中
	customCssList []string

	// customLastUpdateTimeMillis 内存缓存最后一次更新的时间戳c
	customLastUpdateTimeMillis int64

	// customCacheOpMutex 维护内存缓存更新同步
	customCacheOpMutex sync.Mutex
)

// loadAllCustomCssJs 加载所有自定义脚本
// 当内存中已经存在自定义脚本缓存时, 不进行任何操作
// 可通过 forceRefresh 强制刷新缓存
func loadAllCustomCssJs(forceRefresh bool) {

	customCacheOpMutex.Lock()
	noUpdate := !forceRefresh && (len(customCssList) > 0 || len(customJsList) > 0)

	// 间隔太短 不允许刷新
	if time.Now().UnixMilli()-customLastUpdateTimeMillis < customUpdateInterval {
		noUpdate = true
	}

	defer func() {
		if !noUpdate {
			customLastUpdateTimeMillis = time.Now().UnixMilli()
		}
		customCacheOpMutex.Unlock()
	}()

	if noUpdate {
		return
	}

	// loadRemoteContent 尝试将 originBytes 中的内容解析为 url 并解析远程路径下的文本内容
	// 当 originBytes 中记录的不是一个合法的 url 时, 返回 originBytes 本身
	loadRemoteContent := func(originBytes []byte) ([]byte, error) {
		if len(originBytes) == 0 {
			return []byte{}, nil
		}

		str := strings.TrimSpace(string(originBytes))
		u, err := url.Parse(str)
		if err != nil {
			// 非远程地址
			return originBytes, nil
		}

		resp, err := https.Get(u.String()).Do()
		if err != nil {
			return nil, fmt.Errorf("远程加载失败: %s, err: %v", u.String(), err)
		}
		defer resp.Body.Close()
		if !https.IsSuccessCode(resp.StatusCode) {
			return nil, fmt.Errorf("远程错误响应: %s, err: %s", u.String(), resp.Status)
		}

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("远程读取失败: %s, err: %v", u.String(), err)
		}

		return bytes, nil
	}

	// loadFiles 解析路径 fp 下的后缀为 ext 的文件列表，将它们读取成字符串切片后返回
	loadFiles := func(fp, ext, successLogPrefix string) ([]string, error) {
		if err := os.MkdirAll(fp, os.ModePerm); err != nil {
			return nil, fmt.Errorf("目录初始化失败: %s, err: %v", fp, err)
		}

		files, err := os.ReadDir(fp)
		if err != nil {
			return nil, fmt.Errorf("读取目录失败: %s, err: %v, 无法注入自定义脚本", fp, err)
		}

		res := []string{}
		ch := make(chan string)
		chg := new(errgroup.Group)
		chg.Go(func() error {
			for content := range ch {
				res = append(res, content)
			}
			return nil
		})

		g := new(errgroup.Group)
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			if filepath.Ext(file.Name()) != ext {
				continue
			}

			g.Go(func() error {
				content, err := os.ReadFile(filepath.Join(fp, file.Name()))
				if err != nil {
					return fmt.Errorf("读取文件失败: %s, err: %v", file.Name(), err)
				}

				// 支持远程加载
				content, err = loadRemoteContent(content)
				if err != nil {
					return fmt.Errorf("远程加载失败: %s, err: %v", file.Name(), err)
				}

				ch <- string(content)
				logs.Success("%s已加载: %s", successLogPrefix, file.Name())
				return nil
			})

		}

		if err := g.Wait(); err != nil {
			close(ch)
			return nil, err
		}
		close(ch)
		chg.Wait()
		return res, nil
	}

	fp := filepath.Join(config.BasePath, constant.CustomJsDirName)
	jsList, err := loadFiles(fp, ".js", "自定义脚本")
	if err != nil {
		logs.Error("加载自定义脚本异常: %v", err)
		return
	}
	customJsList = jsList

	fp = filepath.Join(config.BasePath, constant.CustomCssDirName)
	cssList, err := loadFiles(fp, ".css", "自定义样式表")
	if err != nil {
		logs.Error("加载自定义样式表异常: %v", err)
		return
	}
	customCssList = cssList
}

// ProxyIndexHtml 代理 index.html 注入自定义脚本样式文件
func ProxyIndexHtml(c *gin.Context) {
	resp, err := https.ProxyRequest(c.Request, config.C.Emby.Host)
	if checkErr(c, err) {
		return
	}
	defer resp.Body.Close()

	if !https.IsSuccessCode(resp.StatusCode) {
		checkErr(c, fmt.Errorf("远程返回错误响应: %v", resp.Status))
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if checkErr(c, err) {
		return
	}

	content := string(bodyBytes)
	bodyCloseTag := "</body>"
	customJsElm := fmt.Sprintf(`    <script src="%s"></script>`, constant.Route_CustomJs)
	content = strings.ReplaceAll(content, bodyCloseTag, customJsElm+"\n"+bodyCloseTag)

	customCssElm := fmt.Sprintf(`    <link rel="stylesheet" type="text/css" href="%s">`, constant.Route_CustomCss)
	content = strings.ReplaceAll(content, bodyCloseTag, customCssElm+"\n"+bodyCloseTag)

	c.Status(resp.StatusCode)
	resp.Header.Del("Content-Length")
	https.CloneHeader(c.Writer, resp.Header)
	c.Writer.Write([]byte(content))
	c.Writer.Flush()
}

// ProxyCustomJs 代理自定义脚本
func ProxyCustomJs(c *gin.Context) {
	loadAllCustomCssJs(config.C.Emby.CustomCssJs.DebugMode)

	contentBuilder := strings.Builder{}
	for _, script := range customJsList {
		contentBuilder.WriteString("(function() {\n")
		contentBuilder.WriteString("  // 等待 Emby 环境和 apiclient 准备就绪后执行脚本\n")
		contentBuilder.WriteString("  function waitForEmby() {\n")
		contentBuilder.WriteString("    if (typeof ApiClient !== 'undefined' && ApiClient !== null) {\n")
		contentBuilder.WriteString("      try {\n")
		contentBuilder.WriteString("        " + script + "\n")
		contentBuilder.WriteString("      } catch (e) {\n")
		contentBuilder.WriteString("        console.error('Custom script error:', e);\n")
		contentBuilder.WriteString("      }\n")
		contentBuilder.WriteString("    } else {\n")
		contentBuilder.WriteString("      setTimeout(waitForEmby, 100);\n")
		contentBuilder.WriteString("    }\n")
		contentBuilder.WriteString("  }\n")
		contentBuilder.WriteString("  // 启动等待\n")
		contentBuilder.WriteString("  waitForEmby();\n")
		contentBuilder.WriteString("})();\n\n")
	}
	contentBytes := []byte(contentBuilder.String())

	c.Status(http.StatusOK)
	c.Header("Content-Type", "application/javascript")
	c.Header("Content-Length", fmt.Sprintf("%d", len(contentBytes)))
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Writer.Write(contentBytes)
	c.Writer.Flush()
}

// ProxyCustomCss 代理自定义样式表
func ProxyCustomCss(c *gin.Context) {
	loadAllCustomCssJs(config.C.Emby.CustomCssJs.DebugMode)

	contentBuilder := strings.Builder{}
	for _, style := range customCssList {
		fmt.Fprintf(&contentBuilder, "%s\n\n\n", style)
	}
	contentBytes := []byte(contentBuilder.String())

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/css")
	c.Header("Content-Length", fmt.Sprintf("%d", len(contentBytes)))
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Writer.Write(contentBytes)
	c.Writer.Flush()
}
