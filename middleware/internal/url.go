package internal

import (
	"net/url"
	"strings"
)

func HostnameSegment(u *url.URL, index int) string {
	return hostnameSegment(u.Hostname(), index)
}

func hostnameSegment(hostname string, index int) string {
	parts := strings.Split(hostname, ".")

	if index < 0 {
		index += len(parts)
	}

	if index >= 0 && index < len(parts) {
		return parts[index]
	}

	return ""
}
