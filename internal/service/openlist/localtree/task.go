package localtree

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
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/lib/ffmpeg"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/music"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/colors"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/mp4s"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/trys"
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

var (
	vw = VirtualWriter{}
	sw = StrmWriter{}
	mw = MusicWriter{}
	rw = RawWriter{}
)

// LoadTaskWriter 根据文件容器加载 TaskWriter
func LoadTaskWriter(container string) TaskWriter {
	cfg := config.C.Openlist.LocalTreeGen
	if cfg.IsVirtual(container) {
		return &vw
	}
	if cfg.IsStrm(container) {
		return &sw
	}
	if cfg.IsMusic(container) {
		return &mw
	}
	return &rw
}

// VirtualWriter 写同名空文件, 尝试写入媒体时长
type VirtualWriter struct{}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (vw *VirtualWriter) Path(path string) string {
	return path
}

// Write 将文件信息写入到本地文件系统中
func (vw *VirtualWriter) Write(task FileTask, localPath string) error {
	// 默认写入时长 3 小时
	dftDuration := time.Hour * 3
	if !config.C.Openlist.LocalTreeGen.FFmpegEnable {
		return os.WriteFile(localPath, mp4s.GenWithDuration(dftDuration), os.ModePerm)
	}

	rmtUrl := sw.OpenlistPath(task)

	var info ffmpeg.Info
	err := trys.Try(func() (err error) {
		info, err = ffmpeg.InspectInfo(rmtUrl)
		return
	}, 3, time.Second)
	if err != nil {
		return fmt.Errorf("调用 ffmpeg 失败: %w", err)
	}
	logf(colors.Gray, "提取文件时长 [%s]: %v", filepath.Base(task.Path), info.Duration)

	return os.WriteFile(localPath, mp4s.GenWithDuration(info.Duration), os.ModePerm)
}

// StrmWriter 写文件对应的 openlist strm 文件
type StrmWriter struct{}

// OpenlistPath 生成媒体的 openlist http 访问地址
func (sw *StrmWriter) OpenlistPath(task FileTask) string {
	segs := strings.Split(strings.TrimPrefix(task.Path, "/"), "/")
	for i, seg := range segs {
		segs[i] = url.PathEscape(seg)
	}

	return fmt.Sprintf(
		"%s/d/%s?sign=%s",
		config.C.Openlist.Host,
		strings.Join(segs, "/"),
		task.Sign,
	)
}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (sw *StrmWriter) Path(path string) string {
	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext) + ".strm"
}

// Write 将文件信息写入到本地文件系统中
func (sw *StrmWriter) Write(task FileTask, localPath string) error {
	return os.WriteFile(localPath, []byte(sw.OpenlistPath(task)), os.ModePerm)
}

// MusicWriter 写同名空文件, 同时尝试写入时长和音乐标签元数据信息
type MusicWriter struct{}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (mw *MusicWriter) Path(path string) string {
	return path
}

// Write 将文件信息写入到本地文件系统中
func (mw *MusicWriter) Write(task FileTask, localPath string) error {
	if !config.C.Openlist.LocalTreeGen.FFmpegEnable {
		// 必须开启 ffmpeg 才能生成, 改用 strm 替代
		return sw.Write(task, localPath)
	}

	rmtUrl := sw.OpenlistPath(task)

	var meta ffmpeg.Music
	err := trys.Try(func() (err error) {
		meta, err = ffmpeg.InspectMusic(rmtUrl)
		return
	}, 3, time.Second)
	if err != nil {
		return fmt.Errorf("提取音乐元数据失败 [%s]: %w", filepath.Base(task.Path), err)
	}
	if meta.Duration == 0 {
		meta.Duration = time.Second
	}

	var pic []byte
	err = trys.Try(func() (err error) {
		pic, err = ffmpeg.ExtractMusicCover(rmtUrl)
		return
	}, 3, time.Second)
	if err != nil {
		return fmt.Errorf("提取音乐封面失败 [%s]: %w", filepath.Base(task.Path), err)
	}

	logf(colors.Gray, "提取音乐元数据 [%s]: [标题: %s] [艺术家: %s] [时长: %v]", filepath.Base(task.Path), meta.Title, meta.Artist, meta.Duration)

	return music.WriteFakeMP3(localPath, meta, pic)
}

// RawWriter 请求 openlist 源文件写入本地
type RawWriter struct {
	mu sync.Mutex
}

// Path 将 openlist 文件路径中的文件名
// 转换为本地文件系统中的文件名
func (rw *RawWriter) Path(path string) string {
	return path
}

// Write 将文件信息写入到本地文件系统中
func (rw *RawWriter) Write(task FileTask, localPath string) error {
	// 防止并发访问网盘触发风控
	rw.mu.Lock()
	defer rw.mu.Unlock()

	header := http.Header{"User-Agent": []string{"libmpv"}}
	res := openlist.FetchFsGet(task.Path, header)
	if res.Code != http.StatusOK {
		return fmt.Errorf("请求 openlist 文件失败: %s", res.Msg)
	}

	u := res.Data.RawUrl

	err := trys.Try(func() (err error) {
		file, err := os.Create(localPath)
		if err != nil {
			return fmt.Errorf("创建文件失败 [%s]: %w", localPath, err)
		}
		defer file.Close()

		resp, err := https.Get(u).Do()
		if err != nil {
			return fmt.Errorf("请求 openlist 直链失败: %w", err)
		}
		defer resp.Body.Close()

		if !https.IsSuccessCode(resp.StatusCode) {
			return fmt.Errorf("请求 openlist 直链失败, 响应状态: %s", resp.Status)
		}

		if _, err = io.Copy(file, resp.Body); err != nil {
			return fmt.Errorf("写入 openlist 源文件到本地磁盘失败, 拷贝异常: %w", err)
		}

		logf(colors.Gray, "下载 openlist 源文件 [%s]", filepath.Base(task.Path))
		return
	}, 3, time.Second)

	return err
}
