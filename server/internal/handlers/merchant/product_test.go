package merchant

import (
	"fz_yyc_api/internal/models"
	"reflect"
	"testing"
)

func TestParseStringArray(t *testing.T) {
	tests := []struct {
		name string
		raw  models.JSON
		want []string
	}{
		{
			name: "json array",
			raw:  models.JSON(`["https://cdn.example.com/a.png","https://cdn.example.com/b.png"]`),
			want: []string{"https://cdn.example.com/a.png", "https://cdn.example.com/b.png"},
		},
		{
			name: "json string",
			raw:  models.JSON(`"https://cdn.example.com/a.png"`),
			want: []string{"https://cdn.example.com/a.png"},
		},
		{
			name: "plain string",
			raw:  models.JSON(`https://cdn.example.com/a.png`),
			want: []string{"https://cdn.example.com/a.png"},
		},
		{
			name: "comma separated plain string",
			raw:  models.JSON(`https://cdn.example.com/a.png, https://cdn.example.com/b.png`),
			want: []string{"https://cdn.example.com/a.png", "https://cdn.example.com/b.png"},
		},
		{
			name: "quoted plain string",
			raw:  models.JSON(`'https://cdn.example.com/a.png'`),
			want: []string{"https://cdn.example.com/a.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseStringArray(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseStringArray(%q) = %#v, want %#v", string(tt.raw), got, tt.want)
			}
		})
	}
}
