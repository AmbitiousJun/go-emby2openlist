package ffmpeg

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/constant"
)

// OpenError ffmpeg 打开文件失败
const OpenError = "Error opening input:"

// mu 任务逐个执行
var mu sync.Mutex

// InspectInfo 检查指定路径文件的元信息
func InspectInfo(path string) (Info, error) {
	if !execOk {
		return Info{}, errors.New("ffmpeg 未初始化")
	}
	mu.Lock()
	defer mu.Unlock()

	cmd := exec.Command(execPath, "-http_proxy", getProxyUrlByPath(path), "-user_agent", constant.CommonDlUserAgent, "-threads", "1", "-i", path)

	outputBytes, _ := cmd.CombinedOutput()
	if bytes.Contains(outputBytes, []byte(OpenError)) {
		return Info{}, errors.New(string(outputBytes[bytes.Index(outputBytes, []byte(OpenError)):]))
	}

	i := Info{}
	if durationReg.Match(outputBytes) {
		i.Duration = resolveDuration(string(outputBytes))
	}

	return i, nil
}

// InspectMusic 检查指定音乐文件的元信息
func InspectMusic(path string) (Music, error) {
	if !execOk {
		return Music{}, errors.New("ffmpeg 未初始化")
	}
	mu.Lock()
	defer mu.Unlock()

	cmd := exec.Command(execPath, "-http_proxy", getProxyUrlByPath(path), "-user_agent", constant.CommonDlUserAgent, "-threads", "1", "-i", path)
	outputBytes, _ := cmd.CombinedOutput()

	if bytes.Contains(outputBytes, []byte(OpenError)) {
		return Music{}, errors.New(string(outputBytes[bytes.Index(outputBytes, []byte(OpenError)):]))
	}

	m := Music{}
	wg := sync.WaitGroup{}
	output := string(outputBytes)
	wg.Add(11)

	go func() {
		defer wg.Done()
		if albumReg.MatchString(output) {
			m.Album = albumReg.FindStringSubmatch(output)[1]
		}
	}()

	go func() {
		defer wg.Done()
		if artistReg.MatchString(output) {
			m.Artist = artistReg.FindStringSubmatch(output)[1]
		}
	}()

	go func() {
		defer wg.Done()
		if commentReg.MatchString(output) {
			m.Comment = commentReg.FindStringSubmatch(output)[1]
		}
	}()

	go func() {
		defer wg.Done()
		if dateReg.MatchString(output) {
			m.Date = dateReg.FindStringSubmatch(output)[1]
		}
	}()

	go func() {
		defer wg.Done()
		if tdorReg.MatchString(output) {
			m.Date = tdorReg.FindStringSubmatch(output)[1]
		}
	}()

	go func() {
		defer wg.Done()
		if genreReg.MatchString(output) {
			m.Genre = genreReg.FindStringSubmatch(output)[1]
		}
	}()

	go func() {
		defer wg.Done()
		m.Title = resolveTitle(output)
	}()

	go func() {
		defer wg.Done()
		m.Duration = resolveDuration(output)
	}()

	go func() {
		defer wg.Done()
		m.Track = resolveTrack(output)
	}()

	go func() {
		defer wg.Done()
		m.Disc = resolveDisc(output)
	}()

	go func() {
		defer wg.Done()
		m.Lyrics = resolveLyrics(output)
	}()

	wg.Wait()
	return m, nil
}

// ExtractMusicCover 解析音乐海报
func ExtractMusicCover(path string) ([]byte, error) {
	if !execOk {
		return nil, errors.New("ffmpeg 未初始化")
	}
	mu.Lock()
	defer mu.Unlock()

	cmd := exec.Command(execPath, "-http_proxy", getProxyUrlByPath(path), "-user_agent", constant.CommonDlUserAgent, "-threads", "1", "-i", path, "-an", "-vframes", "1", "-f", "image2", "-vcodec", "mjpeg", "pipe:1")

	outputBytes, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := exitErr.Stderr
			if bytes.Contains(stderr, []byte(OpenError)) {
				return nil, errors.New(string(stderr[bytes.Index(stderr, []byte(OpenError)):]))
			}
		}
	}

	return append([]byte(nil), outputBytes...), nil
}

// GenSilentMP3Bytes 使用 ffmpeg 生成静音 MP3 并返回字节内容
func GenSilentMP3Bytes(durationSec float64) ([]byte, error) {
	args := []string{
		"-f", "lavfi",
		"-i", "anullsrc=r=8000:cl=mono",
		"-t", fmt.Sprintf("%.2f", durationSec),
		"-acodec", "libmp3lame",
		"-b:a", "8k", // 极低比特率
		"-ar", "8000", // 采样率降为 8000Hz
		"-ac", "1", // 单声道
		"-f", "mp3",
		"pipe:1",
	}

	cmd := exec.Command(execPath, args...)
	outputBytes, _ := cmd.Output()

	return append([]byte(nil), outputBytes...), nil
}
