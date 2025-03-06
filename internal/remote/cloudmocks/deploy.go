package cloudmocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gatehill.io/imposter/internal/remote"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type createMockResponse struct {
	MockId string `json:"mockId"`
}

type getStatusResponse struct {
	Status       string `json:"status"`
	LastModified int    `json:"lastModified"`
}

type getEndpointResponse struct {
	BaseUrl string `json:"baseUrl"`
	SpecUrl string `json:"specUrl"`
}

func (m CloudMocksRemote) Deploy() error {
	if m.Config[configKeyUrl] == "" {
		return fmt.Errorf("URL cannot be null")
	} else if token, _ := m.getObfuscatedToken(); token == "" {
		return fmt.Errorf("auth token cannot be null")
	}

	err := m.ensureMockExists()
	if err != nil {
		return err
	}

	err = m.setMockState("DRAFT")
	if err != nil {
		return err
	}

	err = m.syncFiles(m.Dir)
	if err != nil {
		return err
	}

	err = m.setMockState("LIVE")
	if err != nil {
		return err
	}
	if success := m.waitForStatus("ACTIVE", make(chan bool)); !success {
		return fmt.Errorf("timed out waiting for mock to reach active status")
	}
	return nil
}

func (m CloudMocksRemote) Undeploy() error {
	err := m.setMockState("DRAFT")
	if err != nil {
		return err
	}
	return nil
}

func (m CloudMocksRemote) GetEndpoint() (*remote.EndpointDetails, error) {
	endpoint, err := m.getEndpoint()
	if err != nil {
		return nil, err
	}
	details := &remote.EndpointDetails{
		BaseUrl:   endpoint.BaseUrl,
		SpecUrl:   endpoint.SpecUrl,
		StatusUrl: remote.MustJoinPath(endpoint.BaseUrl, "/system/status"),
	}
	return details, nil
}

func (m CloudMocksRemote) ensureMockExists() error {
	if m.Config[configKeyMockId] == "" {
		logger.Debugf("creating new mock")

		var resp createMockResponse
		err := m.request("POST", "/api/mocks", &resp)
		if err != nil {
			return fmt.Errorf("failed to create mock: %s", err)
		}

		logger.Debugf("created mock with ID: %s", resp.MockId)
		m.Config[configKeyMockId] = resp.MockId
		err = m.SaveConfig()
		if err != nil {
			return fmt.Errorf("failed to save mock ID: %s", err)
		}

	} else {
		logger.Debugf("using existing mock with ID: %s", m.Config[configKeyMockId])
	}
	return nil
}

func (m CloudMocksRemote) setMockState(state string) error {
	logger.Debugf("setting mock state to: %s", state)
	var resp interface{}
	err := m.request("PATCH", fmt.Sprintf("/api/mocks/%s/state/%s", m.Config[configKeyMockId], state), &resp)
	if err != nil {
		return fmt.Errorf("failed to set mock state to: %s: %s", state, err)
	}
	logger.Infof("set mock state to: %s", state)
	return nil
}

func (m CloudMocksRemote) getStatus() (*getStatusResponse, error) {
	var resp getStatusResponse
	err := m.request("GET", fmt.Sprintf("/api/mocks/%s/status", m.Config[configKeyMockId]), &resp)
	if err != nil {
		return nil, fmt.Errorf("error getting status: %s", err)
	}
	return &resp, nil
}

func (m CloudMocksRemote) getEndpoint() (*getEndpointResponse, error) {
	var resp getEndpointResponse
	err := m.request("GET", fmt.Sprintf("/api/mocks/%s/endpoint", m.Config[configKeyMockId]), &resp)
	if err != nil {
		return nil, fmt.Errorf("error getting endpoint: %s", err)
	}
	return &resp, nil
}

func (m CloudMocksRemote) waitForStatus(s string, shutDownC chan bool) bool {
	logger.Infof("waiting for mock status to be: %s...", s)

	finishedC := make(chan bool)
	timeout := time.NewTimer(120 * time.Second)
	defer timeout.Stop()

	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)
			status, err := m.getStatus()
			if err != nil {
				continue
			}
			logger.Tracef("mock status: %v", status.Status)
			if status.Status == s {
				finishedC <- true
				break
			} else if status.Status == "FAILED" {
				finishedC <- false
			}
		}
	}()

	finished := false
	select {
	case <-timeout.C:
		finished = true
		logger.Fatalf("timed out waiting for mock status to be: %s", s)
		return false
	case success := <-finishedC:
		finished = success
		logger.Tracef("finished probe with desired mock status: %v", success)
		return success
	case <-shutDownC:
		if !finished {
			logger.Debugf("aborted status probe")
		}
		return false
	}
}

func (m CloudMocksRemote) request(method string, path string, response interface{}) error {
	url := m.Config[configKeyUrl] + path
	client := http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	token, _ := m.getCleartextToken()
	req.Header = http.Header{
		"Authorization": []string{"Bearer " + token},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed to %s: %s", url, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("error requesting %s - HTTP status: %d", url, resp.StatusCode)
	}

	if resp.ContentLength > 0 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response from %s: %s", url, err)
		}
		_ = resp.Body.Close()
		err = json.Unmarshal(body, response)
		if err != nil {
			return fmt.Errorf("failed to unmarshall response from: %s: %s", url, err)
		}
	}
	return nil
}

func (m CloudMocksRemote) upload(path string, src string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return err
	}
	fileContents, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fi.Name())
	if err != nil {
		return err
	}
	part.Write(fileContents)

	err = writer.Close()
	if err != nil {
		return err
	}

	url := m.Config[configKeyUrl] + path
	client := http.Client{}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	token, _ := m.getCleartextToken()
	req.Header = http.Header{
		"Authorization": []string{"Bearer " + token},
		"Content-Type":  []string{"multipart/form-data; boundary=" + writer.Boundary()},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed to %s: %s", url, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("error requesting %s - HTTP status: %d", url, resp.StatusCode)
	}
	return nil
}
