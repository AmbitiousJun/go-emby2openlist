package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/dhowden/tag"
)

func main() {
	const maxBytes = 1024 * 1024
	url := "http://0.0.0.0:12345/d/%E9%9F%B3%E4%B9%901/%E6%99%A8%E5%86%B0%20-%20%E5%90%AC%E9%97%BB%E8%BF%9C%E6%96%B9%E6%9C%89%E4%BD%A0.mp3?sign=k2VcJsxzmbUkuKrcwJG7rG2VDBzD2nfRi2McLh8MOwk=:0"

	resp, err := https.Get(url).AddHeader("Range", fmt.Sprintf("bytes=0-%d", maxBytes-1)).Do()
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		panic("服务器不支持 Range 请求")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	reader := bytes.NewReader(data)
	meta, err := tag.ReadFrom(reader)
	if err != nil {
		panic(fmt.Errorf("metadata 解析失败，可能数据不足: %w", err))
	}

	fmt.Println("标题:", meta.Title())
	fmt.Println("艺术家:", meta.Artist())
	fmt.Println("专辑:", meta.Album())
	fmt.Println("歌词:", meta.Lyrics())
	if pic := meta.Picture(); pic != nil {
		fmt.Printf("封面图: %s, 大小: %d 字节\n", pic.MIMEType, len(pic.Data))
	}
}
