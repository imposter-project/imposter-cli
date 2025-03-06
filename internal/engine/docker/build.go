package docker

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"gatehill.io/imposter/internal/fileutil"
	"github.com/docker/docker/api/types"
	"io"
	"os"
	"path/filepath"
)

const buildContextConfigDir = "config"
const bundleConfigDestDir = "/opt/imposter/config"

type buildOutput struct {
	Stream string `json:"stream"`
	Error  string `json:"error"`
}

// buildImage builds a Docker image using the specified build context.
func buildImage(buildCtx *bytes.Buffer, destImageAndTag string) error {
	logger.Tracef("building image with tag %s", destImageAndTag)
	ctx, cli, err := buildCliClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	buildResponse, err := cli.ImageBuild(
		ctx,
		buildCtx,
		types.ImageBuildOptions{
			Dockerfile: "Dockerfile",
			Tags:       []string{destImageAndTag},
			Labels: map[string]string{
				"builtwith": "imposter-cli",
			},
		},
	)
	if err != nil {
		return fmt.Errorf("error building image: %v", err)
	}
	return awaitBuildComplete(buildResponse)
}

// addFilesToTar adds the files in the specified directory to a tar archive.
func addFilesToTar(dir string, parentImage string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tarWriter := tar.NewWriter(buf)

	local, err := fileutil.ListFiles(dir, false)
	if err != nil {
		return nil, err
	}

	for _, localFile := range local {
		err := addFileToTar(tarWriter, dir, localFile)
		if err != nil {
			return nil, err
		}
	}

	err = addDockerfile(tarWriter, parentImage)

	tarWriter.Close()
	return buf, nil
}

func addDockerfile(tarWriter *tar.Writer, parentImage string) error {
	dockerfileContent := fmt.Sprintf(`FROM %s
COPY %s %s
`, parentImage, buildContextConfigDir, bundleConfigDestDir)

	header := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfileContent)),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tarWriter.Write([]byte(dockerfileContent)); err != nil {
		return err
	}

	logger.Tracef("added Dockerfile to build context:\n%s", dockerfileContent)
	return nil
}

// addFileToTar adds the specified file to the tar archive.
func addFileToTar(writer *tar.Writer, baseDir string, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("error stating file: %v", err)
	}

	header, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return fmt.Errorf("error creating tar header for: %v: %v", file, err)
	}

	// update the name to correctly reflect the relative path from the base dir
	// prepending "config/" to the path to match the Dockerfile COPY instruction
	relPath, _ := filepath.Rel(baseDir, file)
	header.Name = buildContextConfigDir + "/" + relPath

	if err := writer.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(writer, f); err != nil {
		return err
	}

	return nil
}

// awaitBuildComplete waits for the build process to complete.
func awaitBuildComplete(buildResponse types.ImageBuildResponse) error {
	defer buildResponse.Body.Close()
	decoder := json.NewDecoder(buildResponse.Body)
	for {
		var output buildOutput
		if err := decoder.Decode(&output); err != nil {
			if err == io.EOF {
				break // end of the build output
			}
			// handle other errors
			return fmt.Errorf("error reading JSON stream: %v\n", err)
		}

		if output.Error != "" {
			return fmt.Errorf("failed to build image: %s", output.Error)
		}

		if output.Stream != "" {
			fmt.Print(output.Stream) // Print build output
		}
	}

	// At this point, the build process is complete.
	logger.Infof("build process completed")
	return nil
}
