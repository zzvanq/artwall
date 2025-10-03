package dl

import "testing"

func Test_GetDownloaders(t *testing.T) {
	tests := []struct {
		name    string
		sources []Source
		bits    int
		want    int
		wantErr bool
	}{
		{
			name:    "test nga downloader",
			sources: []Source{{Id: 1, ListUrl: "test.com"}},
			bits:    1,
			want:    1,
		},
		{
			name:    "test nonexistent downloader",
			sources: []Source{{Id: 2, ListUrl: "test.com"}},
			bits:    1024,
			want:    0,
		},
		// {
		// 	name: "nga and wallhaven downloader",
		// 	sources: []Source{
		// 		{Id: 1, ListUrl: "test.com"},
		// 		{Id: 2, ListUrl: "test.com"},
		// 	},
		// 	bits: 3,
		// 	want: 2,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDownloaders(tt.sources, tt.bits, "test")
			if len(got) != tt.want {
				t.Errorf("want %d downloaders, got %d", tt.want, len(got))
			}
		})
	}
}
