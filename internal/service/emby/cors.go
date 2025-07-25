package emby

import (
	"fmt"
	"io"
	"net/http"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/bytess"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/gin-gonic/gin"
)

// ChangeBaseVideoModuleCorsDefined 调整 emby 的播放器 cors 配置, 使其支持跨域播放
func ChangeBaseVideoModuleCorsDefined(c *gin.Context) {
	// 1 代理请求
	c.Request.Header.Del("If-Modified-Since")
	c.Request.Header.Del("If-None-Match")
	resp, err := https.ProxyRequest(c.Request, config.C.Emby.Host)
	if checkErr(c, err) {
		return
	}
	if resp.StatusCode != http.StatusOK {
		checkErr(c, fmt.Errorf("emby 返回非预期状态码: %d", resp.StatusCode))
		return
	}
	resp.Header.Del("Content-Length")
	defer resp.Body.Close()

	// 2 注入 JS 代码补丁
	modObj := `window.defined['modules/htmlvideoplayer/plugin.js']`
	modObjDefault := modObj + ".default"
	modObjPrototype := modObjDefault + ".prototype"
	modObjCorsFunc := modObjPrototype + ".getCrossOriginValue"
	jsScript := fmt.Sprintf(`(function(){ var modFunc; modFunc = function(){if(!%s||!%s||!%s||!%s){console.log('emby 未初始化完成...');setTimeout(modFunc);return;}%s=function(mediaSource,playMethod){return null;};console.log('cors 脚本补丁已注入')}; modFunc() })()`, modObj, modObjDefault, modObjPrototype, modObjCorsFunc, modObjCorsFunc)

	c.Status(http.StatusOK)
	https.CloneHeader(c.Writer, resp.Header)

	buf := bytess.CommonFixedBuffer()
	defer buf.PutBack()
	io.CopyBuffer(c.Writer, resp.Body, buf.Bytes())

	c.Writer.Write([]byte(jsScript))
	c.Writer.Flush()
}
