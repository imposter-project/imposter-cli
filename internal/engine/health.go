/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package engine

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"time"
)

const defaultStartTimeout = 30 * time.Second

func getStartTimeout() time.Duration {
	startTimeout := viper.GetInt("startTimeout")
	if startTimeout == 0 {
		return defaultStartTimeout
	}
	return time.Duration(startTimeout) * time.Second
}

// IsMockUp invokes the status endpoint on the specified port and returns
// a boolean indicating whether it is healthy.
func IsMockUp(port int) (success bool) {
	if err := CheckMockStatus(port); err != nil {
		logger.Errorf("healthcheck request failed for mock: %s", err)
		return false
	}
	return true
}

// CheckMockStatus invokes the status endpoint on the specified port and
// checks it returns an HTTP 200 status.
func CheckMockStatus(port int) error {
	url := getStatusUrl(port)
	logger.Tracef("checking mock engine at %v", url)
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("healthcheck request failed for mock at %s: %s", url, err)
	}
	if _, err := io.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("healthcheck body read failed for mock at %s: %s", url, err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode == 200 {
		logger.Tracef("healthcheck passed for mock at %s", url)
		return nil
	}
	return fmt.Errorf("healthcheck status was %d for mock at %s: %s", resp.StatusCode, url, err)
}

func WaitUntilUp(port int, shutDownC chan bool) (success bool) {
	url := getStatusUrl(port)
	return WaitForUrl(fmt.Sprintf("status endpoint to return HTTP 200 at %v", url), url, shutDownC)
}

func getStatusUrl(port int) string {
	return fmt.Sprintf("http://localhost:%d/system/status", port)
}

func WaitForUrl(desc string, url string, abortC chan bool) (success bool) {
	return WaitForOp(desc, getStartTimeout(), abortC, func() bool {
		resp, err := http.Get(url)
		if err != nil {
			return false
		}
		if _, err := io.ReadAll(resp.Body); err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == 200
	})
}

func WaitForOp(desc string, timeoutDuration time.Duration, abortC chan bool, operation func() bool) (success bool) {
	logger.Tracef("waiting for %s", desc)

	successC := make(chan bool)
	timeout := time.NewTimer(timeoutDuration)
	defer timeout.Stop()

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			if operation() {
				successC <- true
				break
			}
		}
	}()

	finished := false
	select {
	case <-timeout.C:
		finished = true
		logger.Fatalf("timed out waiting for %s", desc)
		return false
	case <-successC:
		finished = true
		logger.Tracef("successfully waited for %s", desc)
		return true
	case <-abortC:
		if !finished {
			logger.Debugf("aborted waiting for %s", desc)
		}
		return false
	}
}

func PopulateHealth(mock *ManagedMock) {
	if mock.Port != 0 {
		if IsMockUp(mock.Port) {
			mock.Health = MockHealthHealthy
		} else {
			mock.Health = MockHealthUnhealthy
		}
	} else {
		mock.Health = MockHealthUnknown
	}
}
