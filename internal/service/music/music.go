package music

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/dhowden/tag"
)

// MaxBytes 尝试从前 1KB 数据中读取出标签数据
const MaxBytes = 1024 * 1024

var mu sync.Mutex

func ExtractRemoteTag(remotePath string) (tag.Metadata, error) {
	mu.Lock()
	defer mu.Unlock()

	resp, err := https.Get(remotePath).
		AddHeader("Range", fmt.Sprintf("bytes=0-%d", MaxBytes-1)).
		Do()
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器不支持 Range 请求")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取数据失败: %w", err)
	}

	reader := bytes.NewReader(data)
	meta, err := tag.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("metadata 解析失败，可能数据不足: %w", err)
	}
	return meta, nil
}
