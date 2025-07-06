package music_test

import (
	"testing"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/music"
)

func TestWriteNFO(t *testing.T) {
	rmt := "http://0.0.0.0:12345/d/%E9%9F%B3%E4%B9%901/%E6%99%A8%E5%86%B0%20-%20%E5%90%AC%E9%97%BB%E8%BF%9C%E6%96%B9%E6%9C%89%E4%BD%A0.mp3?sign=k2VcJsxzmbUkuKrcwJG7rG2VDBzD2nfRi2McLh8MOwk=:0"
	meta, err := music.ExtractRemoteTag(rmt)
	if err != nil {
		t.Fatal(err)
		return
	}

	if err = music.WriteNFO("晨冰 - 听闻远方有你.nfo", meta); err != nil {
		t.Fatal(err)
	}

	if err = music.WritePic("晨冰 - 听闻远方有你."+meta.Picture().Ext, meta); err != nil {
		t.Fatal(err)
	}
}
