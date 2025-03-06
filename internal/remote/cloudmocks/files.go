package cloudmocks

import (
	"fmt"
	"gatehill.io/imposter/internal/fileutil"
	"strings"
)

func (m CloudMocksRemote) syncFiles(dir string) error {
	logger.Debugf("synchronising files from: %s", dir)

	r, err := m.listRemote()
	if err != nil {
		return err
	}

	local, err := fileutil.ListFiles(dir, false)
	if err != nil {
		return err
	}

	err = m.uploadFiles(local)
	if err != nil {
		return err
	}

	delta := m.calculateDelta(dir, r, local)
	err = m.deleteRemote(delta)
	if err != nil {
		return err
	}

	return nil
}

func (m CloudMocksRemote) listRemote() ([]string, error) {
	var resp []string
	err := m.request("GET", fmt.Sprintf("/api/mocks/%s/files", m.Config[configKeyMockId]), &resp)
	if err != nil {
		return nil, fmt.Errorf("error listing files: %s", err)
	}
	return resp, nil
}

// calculateDelta determines the remote files that are not present in dir
func (m CloudMocksRemote) calculateDelta(dir string, remote []string, local []string) []string {
	var delta []string
	for _, r := range remote {
		if !arrayContains(local, dir+"/", r) {
			delta = append(delta, r)
		}
	}
	logger.Debugf("found %v remote files not present in local", len(delta))
	return delta
}

func arrayContains(search []string, trimPrefix string, term string) bool {
	found := false
	for _, s := range search {
		trimmed := strings.TrimPrefix(s, trimPrefix)
		if trimmed == term {
			found = true
			break
		}
	}
	return found
}

func (m CloudMocksRemote) uploadFiles(files []string) error {
	for _, f := range files {
		logger.Infof("uploading: %s", f)
		err := m.upload(fmt.Sprintf("/api/mocks/%s/spec", m.Config[configKeyMockId]), f)
		if err != nil {
			return fmt.Errorf("failed to upload file: %s: %s", f, err)
		}
	}
	return nil
}

func (m CloudMocksRemote) deleteRemote(files []string) error {
	for _, f := range files {
		logger.Infof("deleting remote file: %s", f)
		var resp interface{}
		err := m.request("DELETE", fmt.Sprintf("/api/mocks/%s/files/%s", m.Config[configKeyMockId], f), &resp)
		if err != nil {
			return fmt.Errorf("failed to delete remote file: %s: %s", f, err)
		}
	}
	return nil
}
