package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/lib/ffmpeg"
	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/util/https"
	"github.com/bogem/id3v2"
	"github.com/dhowden/tag"
)

func main() {
	if err := ffmpeg.AutoDownloadExec("."); err != nil {
		panic(err)
	}

	url := "http://0.0.0.0:12345/d/%E9%9F%B3%E4%B9%901/%E6%99%A8%E5%86%B0%20-%20%E5%90%AC%E9%97%BB%E8%BF%9C%E6%96%B9%E6%9C%89%E4%BD%A0.mp3?sign=k2VcJsxzmbUkuKrcwJG7rG2VDBzD2nfRi2McLh8MOwk=:0"
	size := 1024 * 1024
	resp, err := https.Get(url).AddHeader("Range", fmt.Sprintf("bytes=0-%d", size-1)).Do()
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		panic("远程服务器不支持 Range 请求")
	}

	data, _ := io.ReadAll(resp.Body)
	reader := bytes.NewReader(data)

	// 2. 提取 metadata
	meta, err := tag.ReadFrom(reader)
	if err != nil {
		panic(fmt.Errorf("解析标签失败: %w", err))
	}

	// 4. 创建 ID3 标签
	id3tag := id3v2.NewEmptyTag()

	id3tag.SetTitle(meta.Title())
	id3tag.SetArtist(meta.Artist())
	id3tag.SetAlbum(meta.Album())
	id3tag.SetYear(fmt.Sprintf("%d", meta.Year()))
	id3tag.SetGenre(meta.Genre())

	if t, _ := meta.Track(); t > 0 {
		id3tag.AddTextFrame("TRCK", id3v2.EncodingUTF8, fmt.Sprintf("%d", t))
	}

	if l := meta.Lyrics(); l != "" {
		id3tag.AddUnsynchronisedLyricsFrame(id3v2.UnsynchronisedLyricsFrame{
			Encoding:          id3v2.EncodingUTF8,
			Language:          "eng",
			ContentDescriptor: "",
			Lyrics:            l,
		})
	}

	if pic := meta.Picture(); pic != nil {
		log.Println("检测到封面")
		picFrame := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    pic.MIMEType,
			PictureType: id3v2.PTFrontCover,
			Description: "Cover",
			Picture:     pic.Data,
		}
		id3tag.AddAttachedPicture(picFrame)
	}

	// 5. 写标签到文件头
	buf := bytes.Buffer{}
	if _, err := id3tag.WriteTo(&buf); err != nil {
		panic(err)
	}

	info, err := ffmpeg.InspectInfo(url)
	if err != nil {
		panic(err)
	}
	silent, _ := ffmpeg.GenSilentMP3Bytes(info.Duration.Seconds())
	buf.Write(silent)

	if err := os.WriteFile("output.mp3", buf.Bytes(), os.ModePerm); err != nil {
		panic(err)
	}

	fmt.Println("标签已写入 output.mp3")
}
