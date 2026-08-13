package emby

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/path"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/strs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/trys"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/urls"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/web/cache"

	"github.com/gin-gonic/gin"
)

// Redirect2Transcode 将 master 请求重定向到本地 ts 代理
func Redirect2Transcode(c *gin.Context) {
	templateId := c.Query("template_id")
	if strs.AnyEmpty(templateId) {
		// 尝试从 mediaSourceId 中获取 templateId
		itemInfo, err := resolveItemInfo(c, RouteTranscode)
		if checkErr(c, err) {
			return
		}
		templateId = itemInfo.MsInfo.TemplateId
	}

	apiKey := c.Query(QueryApiKeyName)
	openlistPath := c.Query("openlist_path")
	if strs.AnyEmpty(templateId) {
		ProxyOrigin(c)
		return
	}

	// 只有 template id 时, 需要先获取 openlist path
	if strs.AnyEmpty(openlistPath) {
		Redirect2OpenlistLink(c)
		return
	}

	tu, _ := url.Parse(https.ClientRequestHost(c.Request) + "/videos/proxy_playlist")
	q := tu.Query()
	q.Set("openlist_path", openlistPath)
	q.Set(QueryApiKeyName, apiKey)
	q.Set("template_id", templateId)
	tu.RawQuery = q.Encode()
	c.Redirect(http.StatusTemporaryRedirect, tu.String())
}

// Redirect2OpenlistLink 重定向资源到 openlist 网盘直链
func Redirect2OpenlistLink(c *gin.Context) {
	// 不处理字幕接口
	if strings.Contains(strings.ToLower(c.Request.RequestURI), "subtitles") {
		ProxyOrigin(c)
		return
	}

	// 1 解析要请求的资源信息
	itemInfo, err := resolveItemInfo(c, RouteStream)
	if checkErr(c, err) {
		return
	}
	logs.Info("解析到的 itemInfo: %v", itemInfo)

	// 2 如果请求的是转码资源, 重定向到本地的 m3u8 代理服务
	msInfo := itemInfo.MsInfo
	useTranscode := !msInfo.Empty && msInfo.Transcode
	if useTranscode && msInfo.OpenlistPath != "" {
		u, _ := url.Parse(strings.ReplaceAll(MasterM3U8UrlTemplate, "${itemId}", itemInfo.Id))
		q := u.Query()
		q.Set("template_id", itemInfo.MsInfo.TemplateId)
		q.Set(QueryApiKeyName, itemInfo.ApiKey)
		q.Set("openlist_path", itemInfo.MsInfo.OpenlistPath)
		u.RawQuery = q.Encode()
		logs.Success("重定向 playlist: %s", u.String())
		c.Redirect(http.StatusTemporaryRedirect, u.String())
		return
	}

	// 3 请求资源在 Emby 中的 Path 参数
	embyPath, err := getEmbyFileLocalPath(itemInfo)
	if checkErr(c, err) {
		return
	}

	// 4 如果是远程地址 (strm), 重定向处理
	if urls.IsHttpRemote(embyPath) {
		finalPath := config.C.Emby.Strm.MapPath(embyPath)
		finalPath = getFinalRedirectLink(finalPath, c.Request.Header.Clone())
		logs.Success("重定向 strm: %s", finalPath)
		c.Header(cache.HeaderKeyExpired, cache.Duration(time.Minute*10))
		c.Redirect(http.StatusTemporaryRedirect, finalPath)

		// 异步发送一个播放 Playback 请求, 触发 emby 解析 strm 视频格式
		go sendOpenStreamPlaybackInfoReqToOrigin(itemInfo)

		return
	}

	// 5 如果是本地地址
	if config.C.Emby.IsLocalMediaPath(embyPath) {
		if isLocalRequest(c.Request) {
			logs.Info("内网本地媒体: %s, 直接 302 重定向回 Emby 端口直连播放: %s", embyPath, config.C.Emby.Host)
			c.Redirect(http.StatusTemporaryRedirect, config.C.Emby.Host+c.Request.RequestURI)
			return
		}
		// 如果是外网，由于无法直连内网端口，我们返回原有的 original 路径进行代理回源播放
		// （因为有底层的 Context 机制保护，所以即使代理也不会发生内存暴涨）
		logs.Info("外网本地媒体: %s, 执行代理回源播放", embyPath)
		newUri := strings.Replace(c.Request.RequestURI, "stream", "original", 1)
		newUri = strings.Replace(newUri, "universal", "original", 1)
		c.Redirect(http.StatusTemporaryRedirect, newUri)
		return
	}

	// 6 请求 openlist 资源
	fi := openlist.FetchInfo{
		Header:       c.Request.Header.Clone(),
		UseTranscode: useTranscode,
		Format:       msInfo.TemplateId,
	}
	openlistPathRes := path.Emby2Openlist(embyPath)

	allErrors := strings.Builder{}
	// handleOpenlistResource 根据传递的 path 请求 openlist 资源
	handleOpenlistResource := func(path string) bool {
		logs.Info("尝试请求 Openlist 资源: %s", path)
		fi.Path = path
		res := openlist.FetchResource(fi)

		if res.Code != http.StatusOK {
			allErrors.WriteString(fmt.Sprintf("请求 Openlist 失败, code: %d, msg: %s, path: %s;", res.Code, res.Msg, path))
			return false
		}

		// 处理直链
		if !fi.UseTranscode {
			res.Data.Url = config.C.Emby.Strm.MapPath(res.Data.Url)
			logs.Success("请求成功, 重定向到: %s", res.Data.Url)
			c.Header(cache.HeaderKeyExpired, cache.Duration(time.Minute*10))
			c.Redirect(http.StatusTemporaryRedirect, res.Data.Url)
			return true
		}

		// 代理转码 m3u
		u, _ := url.Parse(https.ClientRequestHost(c.Request) + "/videos/proxy_playlist")
		q := u.Query()
		q.Set("template_id", itemInfo.MsInfo.TemplateId)
		q.Set(QueryApiKeyName, itemInfo.ApiKey)
		q.Set("openlist_path", openlist.PathEncode(path))
		u.RawQuery = q.Encode()
		c.Redirect(http.StatusTemporaryRedirect, u.String())
		return true
	}

	if openlistPathRes.Success && handleOpenlistResource(openlistPathRes.Path) {
		return
	}
	paths, err := openlistPathRes.Range()
	if checkErr(c, err) {
		return
	}
	if slices.ContainsFunc(paths, func(path string) bool {
		return handleOpenlistResource(path)
	}) {
		return
	}

	checkErr(c, fmt.Errorf("获取直链失败: %s", allErrors.String()))
}

// ProxyOriginalResource 拦截 original 接口
func ProxyOriginalResource(c *gin.Context) {
	if strings.Contains(strings.ToLower(c.Request.RequestURI), "subtitles") {
		ProxyOrigin(c)
		return
	}

	itemInfo, err := resolveItemInfo(c, RouteOriginal)
	if checkErr(c, err) {
		return
	}

	embyPath, err := getEmbyFileLocalPath(itemInfo)
	if checkErr(c, err) {
		return
	}

	// 如果是本地媒体
	if config.C.Emby.IsLocalMediaPath(embyPath) {
		if isLocalRequest(c.Request) {
			logs.Info("内网本地媒体: %s, 重定向回 Emby 端口直连播放", embyPath)
			c.Redirect(http.StatusTemporaryRedirect, config.C.Emby.Host+c.Request.RequestURI)
			return
		}
		// 外网请求，执行代理回源
		logs.Info("外网本地媒体: %s, 执行代理回源", embyPath)
		ProxyOrigin(c)
		return
	}
	Redirect2OpenlistLink(c)
}

// checkErr 检查 err 是否为空
// 不为空则根据错误处理策略返回响应
//
// 返回 true 表示请求已经被处理
func checkErr(c *gin.Context, err error) bool {
	if err == nil || c == nil {
		return false
	}

	// 异常接口, 不缓存
	c.Header(cache.HeaderKeyExpired, "-1")

	// 采用拒绝策略, 直接返回错误
	if config.C.Emby.ProxyErrorStrategy == config.PeStrategyReject {
		logs.Error("代理接口失败: %v", err)
		c.String(http.StatusInternalServerError, "代理接口失败, 请检查日志")
		return true
	}

	logs.Error("代理接口失败: %v, 回源处理", err)
	ProxyOrigin(c)
	return true
}

// getFinalRedirectLink 尝试对带有重定向的原始链接进行内部请求, 返回最终链接
//
// 检测到 internal-redirect-enable 配置未启用时, 直接返回原始链接
//
// 请求中途出现任何失败都会返回原始链接
func getFinalRedirectLink(originLink string, header http.Header) string {

	if !config.C.Emby.Strm.InternalRedirectEnable {
		logs.Info("internal-redirect-enable 未启用, 使用原始链接")
		return originLink
	}

	var finalLink string
	err := trys.Try(func() (err error) {
		logs.Info("正在尝试内部重定向, originLink: [%s]", originLink)
		fl, resp, e := https.Get(originLink).Header(header).DoRedirect()
		if e != nil {
			return e
		}
		defer resp.Body.Close()
		finalLink = fl
		return nil
	}, 3, time.Second*2)

	if err != nil {
		logs.Warn("内部重定向失败: %v", err)
		return originLink
	}

	return finalLink
}

// 检查客户端请求是否是通过内网 IP 访问
func isLocalRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	host := r.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host == "localhost" ||
		host == "127.0.0.1" ||
		strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.16.") ||
		strings.HasPrefix(host, "172.17.") ||
		strings.HasPrefix(host, "172.18.") ||
		strings.HasPrefix(host, "172.19.") ||
		strings.HasPrefix(host, "172.20.") ||
		strings.HasPrefix(host, "172.21.") ||
		strings.HasPrefix(host, "172.22.") ||
		strings.HasPrefix(host, "172.23.") ||
		strings.HasPrefix(host, "172.24.") ||
		strings.HasPrefix(host, "172.25.") ||
		strings.HasPrefix(host, "172.26.") ||
		strings.HasPrefix(host, "172.27.") ||
		strings.HasPrefix(host, "172.28.") ||
		strings.HasPrefix(host, "172.29.") ||
		strings.HasPrefix(host, "172.30.") ||
		strings.HasPrefix(host, "172.31.")
}
