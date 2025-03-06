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

package stringutil

import (
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"strings"
)

func GetFirstNonEmpty(candidates ...string) string {
	for _, candidate := range candidates {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

// CombineUnique returns the unique union of originals and candidates
func CombineUnique(originals []string, candidates []string) []string {
	return Unique(append(originals, candidates...))
}

// Unique returns the unique elements from the slice
func Unique(candidates []string) []string {
	var unique []string
	seen := make(map[string]bool, len(candidates))
	for _, candidate := range candidates {
		if !seen[candidate] {
			unique = append(unique, candidate)
			seen[candidate] = true
		}
	}
	return unique
}

func Contains(entries []string, searchTerm string) bool {
	for _, entry := range entries {
		if entry == searchTerm {
			return true
		}
	}
	return false
}

func ContainsPrefix(entries []string, searchTerm string) bool {
	for _, entry := range entries {
		if strings.HasPrefix(entry, searchTerm) {
			return true
		}
	}
	return false
}

// GetMatchingSuffix checks if entry has at least one of the given suffixes.
// If a suffix matches, it is returned, otherwise an empty string is returned.
func GetMatchingSuffix(entry string, suffixes []string) string {
	for _, suffix := range suffixes {
		if strings.HasSuffix(entry, suffix) {
			return suffix
		}
	}
	return ""
}

// Sha1hashString returns a SHA1 checksum of the given input string. This
// *must not* be used for cryptographic purposes, as SHA1 is not secure.
func Sha1hashString(input string) string {
	return Sha1hash([]byte(input))
}

// Sha1hash returns a SHA1 checksum of the given input string. This
// *must not* be used for cryptographic purposes, as SHA1 is not secure.
func Sha1hash(input []byte) string {
	h := sha1.New()
	h.Write(input)
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func ToBool(input string) bool {
	return ToBoolWithDefault(input, false)
}

func ToBoolWithDefault(input string, defaultValue bool) bool {
	parsed, err := strconv.ParseBool(input)
	if err != nil {
		return defaultValue
	}
	return parsed
}
