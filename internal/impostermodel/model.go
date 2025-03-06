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

package impostermodel

type ResponseConfig struct {
	StatusCode  int                `json:"statusCode,omitempty"`
	StaticFile  string             `json:"staticFile,omitempty"`
	StaticData  string             `json:"staticData,omitempty"`
	ExampleName string             `json:"exampleName,omitempty"`
	ScriptFile  string             `json:"scriptFile,omitempty"`
	Headers     *map[string]string `json:"headers,omitempty"`
}

type RequestBody struct {
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type Resource struct {
	Path           string             `json:"path"`
	Method         string             `json:"method"`
	QueryParams    *map[string]string `json:"queryParams,omitempty"`
	RequestBody    *RequestBody       `json:"requestBody,omitempty"`
	RequestHeaders *map[string]string `json:"requestHeaders,omitempty"`
	Response       *ResponseConfig    `json:"response,omitempty"`
}

type PluginConfig struct {
	Plugin    string          `json:"plugin"`
	SpecFile  string          `json:"specFile,omitempty"`
	Response  *ResponseConfig `json:"response,omitempty"`
	Resources []Resource      `json:"resources,omitempty"`
}
