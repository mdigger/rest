package rest

import (
	"path"
	"strings"
)

// CompressMimeTypes contains Media-Type patterns that can be compressed.
var CompressMimeTypes = map[string][]string{
	"text":        {"*"},
	"application": {"json", "*+json", "xml", "*+xml", "javascript", "x-javascript", "x-font-ttf"},
	"image":       {"*+xml", "bmp", "vnd.microsoft.icon", "x-icon"},
	"audio":       {"wave", "aiff", "basic", "x-wav"},
	"font":        {"eot", "opentype"},
}

// AddCompressMimeType registers a new type for compression.
func AddCompressMimeType(maintype, subtypePattern string) {
	patterns := CompressMimeTypes[maintype]
	if patterns == nil {
		patterns = make([]string, 0, 1)
	}
	patterns = append(patterns, subtypePattern)
	CompressMimeTypes[maintype] = patterns
}

// isCompress returns true if contentType falls under the definition of patterns
// for supporting compression of data types.
func isCompress(contentType string) bool {
	i := strings.Index(contentType, ";")
	if i == -1 {
		i = len(contentType)
	}
	contentType = strings.TrimSpace(strings.ToLower(contentType[:i]))
	i = strings.Index(contentType, "/")
	if i == -1 || i == len(contentType)-1 {
		return false
	}
	var subtype = contentType[i+1:]
	for _, pattern := range CompressMimeTypes[contentType[:i]] {
		ok, err := path.Match(pattern, subtype)
		if err != nil {
			continue
		}
		if ok {
			return true
		}
	}
	return false
}
