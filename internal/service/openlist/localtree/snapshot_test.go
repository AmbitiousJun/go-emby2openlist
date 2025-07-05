package localtree_test

import (
	"testing"

	"github.com/AmbitiousJun/go-emby2openlist/v2/internal/service/openlist/localtree"
)

func TestSnapshot_Check(t *testing.T) {
	type args struct {
		path string
	}
	s := localtree.NewSnapshot()
	s.Put("/电视剧", true)
	s.Put("/电视剧/aaa.mp4", false)
	tests := []struct {
		name       string
		s          localtree.Snapshot
		args       args
		wantIsDir  bool
		wantExists bool
	}{
		{name: "1", s: s, args: args{path: "/电视剧"}, wantIsDir: true, wantExists: true},
		{name: "2", s: s, args: args{path: "/电视剧/aaa.mp4"}, wantIsDir: false, wantExists: true},
		{name: "3", s: s, args: args{path: "/电视剧/bbb.mp4"}, wantIsDir: false, wantExists: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsDir, gotExists := tt.s.Check(tt.args.path)
			if gotIsDir != tt.wantIsDir {
				t.Errorf("Snapshot.Check() gotIsDir = %v, want %v", gotIsDir, tt.wantIsDir)
			}
			if gotExists != tt.wantExists {
				t.Errorf("Snapshot.Check() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}
