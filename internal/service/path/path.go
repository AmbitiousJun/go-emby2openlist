package path

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/logs"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/urls"
)

// OpenlistPathRes 路径转换结果
type OpenlistPathRes struct {

	// Success 转换是否成功
	Success bool

	// Path 转换后的路径
	Path string

	// Range 遍历所有 Openlist 根路径生成的子路径
	Range func() ([]string, error)
}

// Emby2Openlist Emby 资源路径转 Openlist 资源路径
func Emby2Openlist(embyPath string) OpenlistPathRes {
	pathRoutes := strings.Builder{}
	pathRoutes.WriteString("[")
	pathRoutes.WriteString("\n【原始路径】 => " + embyPath)

	embyPath = urls.Unescape(embyPath)
	pathRoutes.WriteString("\n\n【URL 解码】 => " + embyPath)

	embyPath = urls.TransferSlash(embyPath)
	pathRoutes.WriteString("\n\n【Windows 反斜杠转换】 => " + embyPath)

	embyMount := config.C.Emby.MountPath
	openlistFilePath := strings.TrimPrefix(embyPath, embyMount)
	pathRoutes.WriteString("\n\n【移除 mount-path】 => " + openlistFilePath)

	if mapPath, ok := config.C.Path.MapEmby2Openlist(openlistFilePath); ok {
		openlistFilePath = mapPath
		pathRoutes.WriteString("\n\n【命中 emby2openlist 映射】 => " + openlistFilePath)
	}
	pathRoutes.WriteString("\n]")
	logs.Tip("embyPath 转换路径: %s", pathRoutes.String())

	rangeFunc := func() ([]string, error) {
		filePath, err := SplitFromSecondSlash(openlistFilePath)
		if err != nil {
			return nil, fmt.Errorf("openlistFilePath 解析异常: %s, error: %v", openlistFilePath, err)
		}

		res := openlist.FetchFsList("/", nil)
		if res.Code != http.StatusOK {
			return nil, fmt.Errorf("请求 openlist fs list 接口异常: %s", res.Msg)
		}

		paths := make([]string, 0, len(res.Data.Content))
		for _, c := range res.Data.Content {
			if !c.IsDir {
				continue
			}
			newPath := fmt.Sprintf("/%s%s", c.Name, filePath)
			paths = append(paths, newPath)
		}
		return paths, nil
	}

	return OpenlistPathRes{
		Success: true,
		Path:    openlistFilePath,
		Range:   rangeFunc,
	}
}

// SplitFromSecondSlash 找到给定字符串 str 中第二个 '/' 字符的位置
// 并以该位置为首字符切割剩余的子串返回
func SplitFromSecondSlash(str string) (string, error) {
	str = urls.TransferSlash(str)
	firstIdx := strings.Index(str, "/")
	if firstIdx == -1 {
		return "", fmt.Errorf("字符串不包含 /: %s", str)
	}

	secondIdx := strings.Index(str[firstIdx+1:], "/")
	if secondIdx == -1 {
		return "", fmt.Errorf("字符串只有单个 /: %s", str)
	}

	return str[secondIdx+firstIdx+1:], nil
}
