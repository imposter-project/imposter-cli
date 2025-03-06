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

package openapi

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

type OperationResponse struct {
	Description string

	// key is content type
	Content map[string]interface{}
}

type Operation struct {
	// key is status code
	Responses   map[string]OperationResponse
	Description string
}

type PartialModel struct {
	// key is path
	Paths map[string]map[string]Operation
}

func Parse(specFile string) (*PartialModel, error) {
	reader, err := os.Open(specFile)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	o := PartialModel{}
	err = yaml.Unmarshal(raw, &o)
	if err != nil {
		return nil, fmt.Errorf("error: %v\n", err)
	}

	logger.Tracef("openapi parsed:\n%v\n\n", o)
	return &o, nil
}
