package localtree

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/config"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
)

// FileTask 包含同步必要信息的文件结构
type FileTask struct {
	// Path 文件绝对路径, 与 openlist 对应
	Path string

	// LocalPath 文件要存入本地的路径
	LocalPath string

	// IsDir 是否是目录
	IsDir bool

	// Container 标记文件的容器
	Container string

	// Sign openlist 文件签名
	Sign string
}

func FsGetTask(prefix string, info openlist.FsGet) FileTask {
	container := strings.TrimPrefix(strings.ToLower(filepath.Ext(info.Name)), ".")
	fp := filepath.Join(prefix, info.Name)
	return FileTask{
		Path:      fp,
		LocalPath: fp,
		IsDir:     info.IsDir,
		Sign:      info.Sign,
		Container: container,
	}
}

// TaskWriter 将 openlist 文件写入到本地文件系统
type TaskWriter interface {

	// Path 将 openlist 文件路径中的文件名
	// 转换为本地文件系统中的文件名
	Path(path string) string

	// Write 将文件信息写入到本地文件系统中
	Write(task FileTask, localPath string) error
}

// LoadTaskWriter 根据文件容器加载 TaskWriter
var LoadTaskWriter = (func() func(container string) TaskWriter {
	virtual := VirtualWriter{}
	strm := StrmWriter{}
	raw := RawWriter{}

	return func(container string) TaskWriter {
		cfg := config.C.Openlist.LocalTreeGen
		if cfg.IsVirtual(container) {
			return &virtual
		}
		if cfg.IsStrm(container) {
			return &strm
		}
		return &raw
	}
})()

// VirtualWriter 写同名空文件
type VirtualWriter struct{}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (vw *VirtualWriter) Path(path string) string {
	return path
}

// Write 将文件信息写入到本地文件系统中
func (vw *VirtualWriter) Write(_ FileTask, localPath string) error {
	return os.WriteFile(localPath, []byte{}, os.ModePerm)
}

// StrmWriter 写文件对应的 openlist strm 文件
type StrmWriter struct{}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (sw *StrmWriter) Path(path string) string {
	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext) + ".strm"
}

// Write 将文件信息写入到本地文件系统中
func (sw *StrmWriter) Write(task FileTask, localPath string) error {
	segs := strings.Split(strings.TrimPrefix(task.Path, "/"), "/")
	for i, seg := range segs {
		segs[i] = url.PathEscape(seg)
	}

	rmtUrl := fmt.Sprintf(
		"%s/d/%s?sign=%s",
		config.C.Openlist.Host,
		strings.Join(segs, "/"),
		task.Sign,
	)
	return os.WriteFile(localPath, []byte(rmtUrl), os.ModePerm)
}

// RawWriter 请求 openlist 源文件写入本地
type RawWriter struct{}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (rw *RawWriter) Path(path string) string {
	return path
}

// Write 将文件信息写入到本地文件系统中
func (rw *RawWriter) Write(task FileTask, localPath string) error {
	header := http.Header{"User-Agent": []string{"libmpv"}}
	res := openlist.FetchFsGet(task.Path, header)
	if res.Code != http.StatusOK {
		return fmt.Errorf("请求 openlist 文件失败: %s", res.Msg)
	}

	u := res.Data.RawUrl
	cur, tot := 1, 3

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	var writeErr error
	for cur <= tot {
		resp, err := https.Get(u).Do()
		if err != nil {
			writeErr = err
			cur++
			continue
		}

		if !https.IsSuccessCode(resp.StatusCode) {
			resp.Body.Close()
			writeErr = fmt.Errorf("请求 openlist 直链失败, 响应状态: %s", resp.Status)
			cur++
			continue
		}

		if _, err = io.Copy(file, resp.Body); err != nil {
			resp.Body.Close()
			if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
				return fmt.Errorf("操作本地文件失败, 无法定位指针到文件起始位置: %w", seekErr)
			}
			writeErr = fmt.Errorf("写入 openlist 源文件到本地磁盘失败, 拷贝异常: %w", err)
			cur++
			continue
		}

		resp.Body.Close()
		writeErr = nil
		break
	}

	return writeErr
}
