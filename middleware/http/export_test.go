package http

func InternalHostnameSegment(hostname string, level int) string {
	return hostnameSegment(hostname, level)
}
