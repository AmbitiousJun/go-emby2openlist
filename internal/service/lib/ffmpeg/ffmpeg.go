package ffmpeg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// mu 任务逐个执行
var mu sync.Mutex

// InspectInfo 检查指定路径文件的元信息
func InspectInfo(path string) (Info, error) {
	if !execOk {
		return Info{}, errors.New("ffmpeg 未初始化")
	}
	mu.Lock()
	defer mu.Unlock()

	cmd := exec.Command(execPath, "-i", path)
	out, err := cmd.StderrPipe()
	if err != nil {
		return Info{}, fmt.Errorf("无法获取 Stdout 管道: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return Info{}, fmt.Errorf("启动 ffmpeg 进程失败: %w", err)
	}

	scanner := bufio.NewScanner(out)
	i := Info{}

	durationReg := regexp.MustCompile(`Duration: (\d+):(\d+):(\d+)\.(\d+)`)
	for scanner.Scan() {
		line := scanner.Text()

		if durationReg.MatchString(line) {
			res := durationReg.FindStringSubmatch(line)
			hour, _ := strconv.Atoi(res[1])
			minute, _ := strconv.Atoi(res[2])
			second, _ := strconv.Atoi(res[3])
			minSecond, _ := strconv.Atoi(res[4])
			i.Duration = time.Hour*time.Duration(hour) +
				time.Minute*time.Duration(minute) +
				time.Second*time.Duration(second) +
				time.Millisecond*time.Duration(minSecond)*100
			break
		}
	}

	cmd.Process.Kill()
	cmd.Wait()
	return i, nil
}

// GenSilentMP3Bytes 使用 ffmpeg 生成静音 MP3 并返回字节内容
func GenSilentMP3Bytes(durationSec float64) ([]byte, error) {
	args := []string{
		"-f", "lavfi",
		"-i", "anullsrc=r=44100:cl=mono",
		"-t", fmt.Sprintf("%d", int(durationSec)),
		"-acodec", "libmp3lame",
		"-q:a", "9",
		"-f", "mp3",
		"pipe:1", // 输出到 stdout
	}

	cmd := exec.Command(execPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil // 或者接 stderr 输出查看错误

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
