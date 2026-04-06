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

package cmd

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_runWithRecovery(t *testing.T) {
	tests := []struct {
		name      string
		fn        func()
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "successful function returns no error",
			fn:      func() {},
			wantErr: false,
		},
		{
			name: "function that calls logger.Fatal is recovered",
			fn: func() {
				logger.Fatal("simulated fatal error")
			},
			wantErr:   true,
			errSubstr: "fatal exit with code",
		},
		{
			name: "function that calls logger.Fatalf is recovered",
			fn: func() {
				logger.Fatalf("simulated fatal: %s", "test")
			},
			wantErr:   true,
			errSubstr: "fatal exit with code",
		},
		{
			name: "function that panics is recovered",
			fn: func() {
				panic("something went wrong")
			},
			wantErr:   true,
			errSubstr: "something went wrong",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runWithRecovery(tt.fn)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errSubstr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_runWithRecovery_restores_exit_func(t *testing.T) {
	originalExitFunc := logger.ExitFunc

	_ = runWithRecovery(func() {
		logger.Fatal("trigger recovery")
	})

	require.Equal(t, fmt.Sprintf("%p", originalExitFunc), fmt.Sprintf("%p", logger.ExitFunc),
		"ExitFunc should be restored after recovery")
}
