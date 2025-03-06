package proxy

import (
	"mime"
	"regexp"
)

// textMediaTypes are the media types that are eligible for
// request capture and response rewriting.
var textMediaTypes = []string{
	"text/.+",
	"application/javascript",
	"application/json",
	"application/xml",
	"application/x-www-form-urlencoded",
}

// isTextContentType returns true if the media type is eligible for
// request capture and response rewriting.
func isTextContentType(contentType string) bool {
	if contentType == "" {
		logger.Warnf("missing content type")
		return false
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		logger.Warnf("failed to parse content type: %v: %v", contentType, err)
		return false
	}
	for _, rewriteMediaType := range textMediaTypes {
		if matched, _ := regexp.MatchString(rewriteMediaType, mediaType); matched {
			return true
		}
	}
	return false
}
