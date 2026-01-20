package webproxy

import (
	"net/url"
	"os"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
)

var (

	// HttpUrl http 代理地址
	HttpUrl *url.URL = nil

	// HttpsUrl https 代理地址
	HttpsUrl *url.URL = nil
)

func init() {
	env := os.Getenv("HTTP_PROXY")
	u, err := url.Parse(env)
	if err != nil {
		logs.Warn("解析 http 代理地址 [%s] 异常: %v", env, err)
	} else if u.String() != "" {
		HttpUrl = u
		logs.Success("启用 http 代理: [%s]", env)
	}

	env = os.Getenv("HTTPS_PROXY")
	u, err = url.Parse(env)
	if err != nil {
		logs.Warn("解析 https 代理地址 [%s] 异常: %v", env, err)
	} else if u.String() != "" {
		HttpsUrl = u
		logs.Success("启用 https 代理: [%s]", env)
	}
}
