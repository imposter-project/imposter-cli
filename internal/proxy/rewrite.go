/*
Copyright © 2022 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Proxy 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"bytes"
	"fmt"
	"net/http"
)

func Rewrite(respHeaders *http.Header, respBody *[]byte, upstream string, port int) *[]byte {
	contentType := (*respHeaders).Get("Content-Type")
	if !isTextContentType(contentType) {
		logger.Debugf("unsupported content type '%s' for rewrite - skipping rewrite", contentType)
		return respBody
	}
	rewritten := bytes.ReplaceAll(*respBody, []byte(upstream), []byte(fmt.Sprintf("http://localhost:%d", port)))
	respBody = &rewritten
	return respBody
}
