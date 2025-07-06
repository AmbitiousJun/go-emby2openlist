package music

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/lib/ffmpeg"
	"github.com/bogem/id3v2"
	"github.com/dhowden/tag"
)

// MusicNFO 音乐元数据
type MusicNFO struct {
	XMLName xml.Name `xml:"music"`
	Title   string   `xml:"title,omitempty"`
	Artist  string   `xml:"artist,omitempty"`
	Album   string   `xml:"album,omitempty"`
	Year    int      `xml:"year,omitempty"`
	Genre   string   `xml:"genre,omitempty"`
	Track   int      `xml:"track,omitempty"`
	Lyrics  string   `xml:"lyrics,omitempty"`
	Comment string   `xml:"comment,omitempty"`
}

// WriteNFO 写入音乐元数据到本地
func WriteNFO(filePath string, meta tag.Metadata) error {
	if meta == nil {
		return fmt.Errorf("元数据为空")
	}

	track, _ := meta.Track()
	info := MusicNFO{
		Title:   meta.Title(),
		Artist:  meta.Artist(),
		Album:   meta.Album(),
		Year:    meta.Year(),
		Genre:   meta.Genre(),
		Track:   track,
		Lyrics:  meta.Lyrics(),
		Comment: meta.Comment(),
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("无法创建文件: %w", err)
	}
	defer f.Close()

	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	return enc.Encode(info)
}

// WritePic 写入音乐封面到本地
func WritePic(filePath string, meta tag.Metadata) error {
	if meta == nil || meta.Picture() == nil {
		return fmt.Errorf("元数据为空")
	}
	return os.WriteFile(filePath, meta.Picture().Data, os.ModePerm)
}

// WriteFakeMP3 将音乐元数据写入一个本地虚假 mp3 文件中, 需要先初始化 ffmpeg
//
// 可通过 d 参数指定生成音频的时长
func WriteFakeMP3(filePath string, meta tag.Metadata, d time.Duration) error {
	if meta == nil {
		return fmt.Errorf("元数据为空")
	}

	// 创建 ID3 标签
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
		picFrame := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    pic.MIMEType,
			PictureType: id3v2.PTFrontCover,
			Description: "Cover",
			Picture:     pic.Data,
		}
		id3tag.AddAttachedPicture(picFrame)
	}

	// 写标签到文件头
	buf := bytes.Buffer{}
	if _, err := id3tag.WriteTo(&buf); err != nil {
		return fmt.Errorf("写入标签至缓冲区发生异常: %w", err)
	}

	silent, err := ffmpeg.GenSilentMP3Bytes(d.Seconds())
	if err != nil {
		return fmt.Errorf("生成虚拟静音音频失败: %w", err)
	}
	if _, err := buf.Write(silent); err != nil {
		return fmt.Errorf("写入虚拟静音音频至缓冲区发生异常: %w", err)
	}

	return os.WriteFile(filePath, buf.Bytes(), os.ModePerm)
}
