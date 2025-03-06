package stringutil

import (
	"reflect"
	"testing"
)

func TestGetMatchingSuffix(t *testing.T) {
	type args struct {
		entry    string
		suffixes []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no suffixes",
			args: args{
				entry:    "foo",
				suffixes: []string{},
			},
			want: "",
		},
		{
			name: "no match",
			args: args{
				entry:    "foo",
				suffixes: []string{".bar", ".qux"},
			},
			want: "",
		},
		{
			name: "match",
			args: args{
				entry:    "foo.bar",
				suffixes: []string{".bar", ".qux"},
			},
			want: ".bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMatchingSuffix(tt.args.entry, tt.args.suffixes); got != tt.want {
				t.Errorf("GetMatchingSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name       string
		candidates []string
		want       string
	}{
		{
			name:       "all empty",
			candidates: []string{"", "", ""},
			want:       "",
		},
		{
			name:       "first non-empty",
			candidates: []string{"", "foo", "bar"},
			want:       "foo",
		},
		{
			name:       "all non-empty",
			candidates: []string{"foo", "bar", "baz"},
			want:       "foo",
		},
		{
			name:       "no candidates",
			candidates: []string{},
			want:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFirstNonEmpty(tt.candidates...); got != tt.want {
				t.Errorf("GetFirstNonEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCombineUnique(t *testing.T) {
	tests := []struct {
		name       string
		originals  []string
		candidates []string
		want       []string
	}{
		{
			name:       "no duplicates",
			originals:  []string{"foo", "bar"},
			candidates: []string{"baz", "qux"},
			want:       []string{"foo", "bar", "baz", "qux"},
		},
		{
			name:       "with duplicates",
			originals:  []string{"foo", "bar"},
			candidates: []string{"bar", "baz"},
			want:       []string{"foo", "bar", "baz"},
		},
		{
			name:       "empty originals",
			originals:  []string{},
			candidates: []string{"foo", "bar"},
			want:       []string{"foo", "bar"},
		},
		{
			name:       "empty candidates",
			originals:  []string{"foo", "bar"},
			candidates: []string{},
			want:       []string{"foo", "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CombineUnique(tt.originals, tt.candidates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CombineUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name       string
		candidates []string
		want       []string
	}{
		{
			name:       "no duplicates",
			candidates: []string{"foo", "bar", "baz"},
			want:       []string{"foo", "bar", "baz"},
		},
		{
			name:       "with duplicates",
			candidates: []string{"foo", "bar", "foo", "baz", "bar"},
			want:       []string{"foo", "bar", "baz"},
		},
		{
			name:       "all duplicates",
			candidates: []string{"foo", "foo", "foo"},
			want:       []string{"foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unique(tt.candidates); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name       string
		entries    []string
		searchTerm string
		want       bool
	}{
		{
			name:       "found",
			entries:    []string{"foo", "bar", "baz"},
			searchTerm: "bar",
			want:       true,
		},
		{
			name:       "not found",
			entries:    []string{"foo", "bar", "baz"},
			searchTerm: "qux",
			want:       false,
		},
		{
			name:       "empty slice",
			entries:    []string{},
			searchTerm: "foo",
			want:       false,
		},
		{
			name:       "empty search term",
			entries:    []string{"foo", "bar", ""},
			searchTerm: "",
			want:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.entries, tt.searchTerm); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsPrefix(t *testing.T) {
	tests := []struct {
		name       string
		entries    []string
		searchTerm string
		want       bool
	}{
		{
			name:       "exact match",
			entries:    []string{"foo", "bar", "baz"},
			searchTerm: "bar",
			want:       true,
		},
		{
			name:       "prefix match",
			entries:    []string{"foo", "bar", "baz"},
			searchTerm: "ba",
			want:       true,
		},
		{
			name:       "no match",
			entries:    []string{"foo", "bar", "baz"},
			searchTerm: "qux",
			want:       false,
		},
		{
			name:       "empty prefix",
			entries:    []string{"foo", "bar", "baz"},
			searchTerm: "",
			want:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsPrefix(tt.entries, tt.searchTerm); got != tt.want {
				t.Errorf("ContainsPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSha1hashString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			name:  "hello world",
			input: "hello world",
			want:  "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name:  "special characters",
			input: "!@#$%^&*()",
			want:  "bf24d65c9bb05b9b814a966940bcfa50767c8a8d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sha1hashString(tt.input); got != tt.want {
				t.Errorf("Sha1hashString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSha1hash(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "empty bytes",
			input: []byte{},
			want:  "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			name:  "hello world bytes",
			input: []byte("hello world"),
			want:  "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name:  "binary data",
			input: []byte{0x00, 0xFF, 0x42},
			want:  "7efad0e9852eab25a68211cf38693671e2b4a77d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sha1hash(tt.input); got != tt.want {
				t.Errorf("Sha1hash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "true string",
			input: "true",
			want:  true,
		},
		{
			name:  "false string",
			input: "false",
			want:  false,
		},
		{
			name:  "1 string",
			input: "1",
			want:  true,
		},
		{
			name:  "0 string",
			input: "0",
			want:  false,
		},
		{
			name:  "invalid string",
			input: "not a bool",
			want:  false,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToBool(tt.input); got != tt.want {
				t.Errorf("ToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToBoolWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue bool
		want         bool
	}{
		{
			name:         "true string",
			input:        "true",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "false string",
			input:        "false",
			defaultValue: true,
			want:         false,
		},
		{
			name:         "invalid with true default",
			input:        "not a bool",
			defaultValue: true,
			want:         true,
		},
		{
			name:         "invalid with false default",
			input:        "not a bool",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "empty with true default",
			input:        "",
			defaultValue: true,
			want:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToBoolWithDefault(tt.input, tt.defaultValue); got != tt.want {
				t.Errorf("ToBoolWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
