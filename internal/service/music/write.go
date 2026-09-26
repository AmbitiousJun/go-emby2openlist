package music

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/lib/ffmpeg"
	"github.com/bogem/id3v2"
)

// MusicNFO 音乐元数据
type MusicNFO struct {
	XMLName xml.Name `xml:"music"`
	Title   string   `xml:"title,omitempty"`
	Artist  string   `xml:"artist,omitempty"`
	Album   string   `xml:"album,omitempty"`
	Year    string   `xml:"year,omitempty"`
	Track   string   `xml:"track,omitempty"`
	Lyrics  string   `xml:"lyrics,omitempty"`
	Comment string   `xml:"comment,omitempty"`
}

// WriteNFO 写入音乐元数据到本地
func WriteNFO(filePath string, meta ffmpeg.Music) error {

	info := MusicNFO{
		Title:   meta.Title,
		Artist:  meta.Artist,
		Album:   meta.Album,
		Year:    meta.Date,
		Track:   meta.Track,
		Lyrics:  meta.Lyrics,
		Comment: meta.Comment,
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

// WriteFakeMP3 将音乐元数据写入一个本地虚假 mp3 文件中, 需要先初始化 ffmpeg
//
// 可通过 d 参数指定生成音频的时长
func WriteFakeMP3(filePath string, meta ffmpeg.Music, pic []byte) error {
	// 创建 ID3 标签
	id3tag := id3v2.NewEmptyTag()

	id3tag.SetTitle(meta.Title)
	id3tag.SetArtist(meta.Artist)
	id3tag.SetAlbum(meta.Album)
	id3tag.SetYear(meta.Date)
	id3tag.SetGenre(meta.Genre)

	id3tag.AddTextFrame("TRCK", id3v2.EncodingUTF8, meta.Track)
	id3tag.AddTextFrame("TPOS", id3v2.EncodingUTF8, meta.Disc)
	id3tag.AddTextFrame("TPE2", id3v2.EncodingUTF8, meta.AlbumArtist)

	if l := meta.Lyrics; l != "" {
		id3tag.AddUnsynchronisedLyricsFrame(id3v2.UnsynchronisedLyricsFrame{
			Encoding:          id3v2.EncodingUTF8,
			Language:          "eng",
			ContentDescriptor: "",
			Lyrics:            l,
		})
	}

	if len(pic) > 0 {
		picFrame := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTFrontCover,
			Description: "Cover",
			Picture:     pic,
		}
		id3tag.AddAttachedPicture(picFrame)
	}

	// 写标签到文件头
	buf := bytes.Buffer{}
	if _, err := id3tag.WriteTo(&buf); err != nil {
		return fmt.Errorf("写入标签至缓冲区发生异常: %w", err)
	}

	silent, err := ffmpeg.GenSilentMP3Bytes(meta.Duration.Seconds())
	if err != nil {
		return fmt.Errorf("生成虚拟静音音频失败: %w", err)
	}
	if _, err := buf.Write(silent); err != nil {
		return fmt.Errorf("写入虚拟静音音频至缓冲区发生异常: %w", err)
	}

	return os.WriteFile(filePath, buf.Bytes(), os.ModePerm)
}
