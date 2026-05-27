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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_WaitForOp(t *testing.T) {
	t.Run("succeeds when the operation passes", func(t *testing.T) {
		success, timedOut := WaitForOp("op", time.Second, nil, func() bool {
			return true
		})
		assert.True(t, success)
		assert.False(t, timedOut)
	})

	t.Run("reports timeout without terminating the process", func(t *testing.T) {
		success, timedOut := WaitForOp("op", 50*time.Millisecond, nil, func() bool {
			return false
		})
		assert.False(t, success, "operation that never passes must not report success")
		assert.True(t, timedOut, "timeout must be distinguishable so callers can clean up")
	})

	t.Run("reports abort separately from timeout", func(t *testing.T) {
		abortC := make(chan bool, 1)
		abortC <- true
		success, timedOut := WaitForOp("op", time.Minute, abortC, func() bool {
			return false
		})
		assert.False(t, success)
		assert.False(t, timedOut, "an external abort is not a timeout")
	})
}
