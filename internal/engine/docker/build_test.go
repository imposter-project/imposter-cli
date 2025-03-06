package docker

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

func Test_addFilesToTar(t *testing.T) {
	config := []byte("plugin: rest\nresponse: { content: \"hello\" }")
	dockerfile := []byte("FROM imposter:latest\nCOPY config /opt/imposter/config\n")

	tempDir := t.TempDir()
	t.Logf("temp dir: %s", tempDir)
	err := os.WriteFile(tempDir+"/test-config-yaml", config, 0644)
	if err != nil {
		t.Fatal(fmt.Errorf("error writing test config file: %s", err.Error()))
	}

	type args struct {
		dir         string
		parentImage string
	}
	type want struct {
		header tar.Header
		body   []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []want
		wantErr bool
	}{
		{
			name: "should add files to tar",
			args: args{
				dir:         tempDir,
				parentImage: "imposter:latest",
			},
			want: []want{
				{
					header: tar.Header{
						Name: "config/test-config-yaml",
						Size: int64(len(config)),
					},
					body: config,
				},
				{
					header: tar.Header{
						Name: "Dockerfile",
						Size: int64(len(dockerfile)),
					},
					body: dockerfile,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addFilesToTar(tt.args.dir, tt.args.parentImage)
			if (err != nil) != tt.wantErr {
				t.Errorf("addFilesToTar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tarReader := tar.NewReader(got)

			for _, f := range tt.want {
				header, err := tarReader.Next()
				if err != nil {
					t.Errorf("addFilesToTar() error reading tar header = %v", err)
					return
				}
				if header.Name != f.header.Name {
					t.Errorf("addFilesToTar() header name = %v, want %v", header.Name, f.header.Name)
				}
				if header.Size != f.header.Size {
					t.Errorf("addFilesToTar() header size = %v, want %v", header.Size, f.header.Size)
				}

				body, err := io.ReadAll(tarReader)
				if err != nil {
					t.Errorf("addFilesToTar() error reading tar body = %v", err)
					return
				}
				if !reflect.DeepEqual(body, f.body) {
					t.Errorf("addFilesToTar() body = %v, want %v", string(body), string(f.body))
				}
			}
		})
	}
}
