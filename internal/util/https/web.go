package https

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/bytess"
)

// ExtractReqBody 克隆并提取请求体, 返回一个新的可被读取的流
func ExtractReqBody(r io.ReadCloser) ([]byte, io.ReadCloser, error) {
	if r == nil {
		return nil, nil, nil
	}
	defer r.Close()

	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	newBody := io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes, newBody, nil
}

// ClientRequestHost 获取客户端请求的 Host
func ClientRequestHost(r *http.Request) string {
	if r == nil {
		return ""
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

// ClientRequestUrl 获取客户端请求的完整地址
func ClientRequestUrl(r *http.Request) string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("%s%s", ClientRequestHost(r), r.URL.String())
}

// IsErrorStatus 判断一个请求响应是否是错误响应
//
// 判断标准是响应码以 4xx 5xx 开头
func IsErrorStatus(code int) bool {
	str := strconv.Itoa(code)
	return strings.HasPrefix(str, "4") || strings.HasPrefix(str, "5")
}

// CloneHeader 克隆 http 头部到 web 的响应头中
func CloneHeader(dst http.ResponseWriter, src http.Header) {
	if dst == nil || src == nil {
		return
	}
	for key, values := range src {
		dst.Header().Del(key)
		for _, value := range values {
			dst.Header().Add(key, value)
		}
	}
}

// ProxyRequest 代理请求, 返回远程响应
func ProxyRequest(r *http.Request, remote string) (*http.Response, error) {
	if r == nil || remote == "" {
		return nil, errors.New("参数为空")
	}

	// 1 解析远程地址
	rawUrl := remote + r.RequestURI
	_, err := url.Parse(rawUrl)
	if err != nil {
		return nil, fmt.Errorf("解析远程地址失败: %v", err)
	}

	// 2 发送请求
	return Request(r.Method, rawUrl).
		Header(r.Header).
		Body(r.Body).
		Context(r.Context()).
		Do()
}

// ProxyPass 代理转发请求
func ProxyPass(r *http.Request, w http.ResponseWriter, remote string) error {
	if r == nil || remote == "" {
		return errors.New("参数为空")
	}

	// 1 代理请求
	resp, err := ProxyRequest(r, remote)
	if err != nil {
		return fmt.Errorf("代理请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 2 回写响应头
	CloneHeader(w, resp.Header)
	w.WriteHeader(resp.StatusCode)

	// 3 回写响应体
	buf := bytess.CommonFixedBuffer()
	defer buf.PutBack()
	io.CopyBuffer(w, resp.Body, buf.Bytes())
	return err
}
