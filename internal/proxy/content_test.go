package proxy

import "testing"

func Test_isTextContentType(t *testing.T) {
	type args struct {
		contentType string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty content type",
			args: args{contentType: ""},
			want: false,
		},
		{
			name: "invalid content type",
			args: args{contentType: "invalid/contenttype"},
			want: false,
		},
		{
			name: "text content type",
			args: args{contentType: "text/plain"},
			want: true,
		},
		{
			name: "json content type",
			args: args{contentType: "application/json"},
			want: true,
		},
		{
			name: "xml content type",
			args: args{contentType: "application/xml"},
			want: true,
		},
		{
			name: "javascript content type",
			args: args{contentType: "application/javascript"},
			want: true,
		},
		{
			name: "form-urlencoded content type",
			args: args{contentType: "application/x-www-form-urlencoded"},
			want: true,
		},
		{
			name: "non-text content type",
			args: args{contentType: "application/octet-stream"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTextContentType(tt.args.contentType); got != tt.want {
				t.Errorf("isTextContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}
