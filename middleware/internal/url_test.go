package internal_test

import (
	"net/url"
	"testing"

	"github.com/aserto-dev/go-aserto/middleware/internal"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name     string
	hostname string
	level    int
	expected string
}

func TestHostnameSegment(t *testing.T) {
	testCases := []TestCase{
		{"should accept a valid positive index", "http://user.example.com", 0, "user"},
		{"should accept a valid negative index", "http://com.example.user", -1, "user"},
		{"should be empty if index is too high", "http://user.example.com", 5, ""},
		{"should be empty if index is too low", "http://user.example.com", -5, ""},
		{"should be empty hostname is empty", "", 0, ""},
	}

	for _, test := range testCases {
		t.Run(test.name, hostnameSegmentTest(test))
	}
}

func hostnameSegmentTest(test TestCase) func(*testing.T) {
	return func(t *testing.T) {
		u, err := url.Parse(test.hostname)
		require.NoError(t, err)

		actual := internal.HostnameSegment(u, test.level)
		require.Equal(t, test.expected, actual)
	}
}
