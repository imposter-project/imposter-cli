package prefs

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"testing"
)

func setup() {
	tmpMetaDir, err := os.MkdirTemp(os.TempDir(), "imposter-prefs")
	if err != nil {
		panic(fmt.Errorf("unable to create test prefs dir: %s", err))
	}
	fmt.Printf("using test prefs dir: %s\n", tmpMetaDir)
	viper.Set("prefs.dir", tmpMetaDir)
}

func cleanup() {
	viper.Set("prefs.dir", nil)
}

func createTempPrefs(t *testing.T) Prefs {
	p := Load("prefs.json")
	return p
}

func TestReadMetaProperty(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name          string
		args          args
		writeTestProp interface{}
		want          interface{}
		wantErr       bool
	}{
		{name: "read missing property", writeTestProp: nil, args: args{key: "foo"}, want: nil, wantErr: false},
		{name: "read existing property", writeTestProp: "baz", args: args{key: "bar"}, want: "baz", wantErr: false},
	}
	setup()
	t.Cleanup(cleanup)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createTempPrefs(t)
			if tt.writeTestProp != nil {
				err := p.WriteProperty(tt.args.key, tt.writeTestProp)
				if err != nil {
					t.Errorf("could not write test prop: %s: %s", tt.writeTestProp, err)
				}
			}
			got, err := p.readProperty(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("readProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readProperty() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadMetaPropertyString(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name          string
		args          args
		writeTestProp string
		want          string
		wantErr       bool
	}{
		{name: "read missing string", writeTestProp: "", args: args{key: "foo"}, want: "", wantErr: false},
		{name: "read existing string", writeTestProp: "baz", args: args{key: "bar"}, want: "baz", wantErr: false},
	}
	setup()
	t.Cleanup(cleanup)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createTempPrefs(t)
			if tt.writeTestProp != "" {
				err := p.WriteProperty(tt.args.key, tt.writeTestProp)
				if err != nil {
					t.Errorf("could not write test prop: %s: %s", tt.writeTestProp, err)
				}
			}
			got, err := p.ReadPropertyString(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadPropertyString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadPropertyString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadMetaPropertyInt(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name          string
		args          args
		writeTestProp int
		want          int
		wantErr       bool
	}{
		{name: "read missing int", writeTestProp: 0, args: args{key: "foo"}, want: 0, wantErr: false},
		{name: "read existing int", writeTestProp: 1, args: args{key: "bar"}, want: 1, wantErr: false},
	}
	setup()
	t.Cleanup(cleanup)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createTempPrefs(t)
			if tt.writeTestProp != 0 {
				err := p.WriteProperty(tt.args.key, tt.writeTestProp)
				if err != nil {
					t.Errorf("could not write test prop: %d: %s", tt.writeTestProp, err)
				}
			}
			got, err := p.ReadPropertyInt(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestReadMetaPropertyInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TestReadMetaPropertyInt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteMetaProperty(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "write string", args: args{key: "qux", value: "corge"}, wantErr: false},
		{name: "write int", args: args{key: "grault", value: 7}, wantErr: false},
	}
	setup()
	t.Cleanup(cleanup)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createTempPrefs(t)
			if err := p.WriteProperty(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("WriteProperty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
