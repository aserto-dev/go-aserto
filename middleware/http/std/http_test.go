package std_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpmw "github.com/aserto-dev/go-aserto/middleware/http/std"
	"github.com/aserto-dev/go-aserto/middleware/internal/test"
	assert "github.com/stretchr/testify/require"
)

type TestCase struct {
	*test.Case
	expectedStatusCode int
	middleware         *httpmw.Middleware
}

type testOptions struct {
	test.Options
	expectedStatusCode int
	callback           func(*httpmw.Middleware)
}

func (opts *testOptions) HasStatusCode() bool {
	return opts.expectedStatusCode != 0
}

const DefaultPolicyPath = "GET.foo"

func NewTest(t *testing.T, name string, options *testOptions) *TestCase {
	if !options.HasPolicy() {
		options.PolicyPath = DefaultPolicyPath
	}

	if !options.HasStatusCode() {
		options.expectedStatusCode = http.StatusOK
	}

	base := test.NewTest(t, name, &options.Options)

	mw := httpmw.New(base.Client, test.Policy(""))

	if options.callback == nil {
		mw.Identity.Subject().ID(test.DefaultUsername)
	} else {
		options.callback(mw)
	}

	return &TestCase{Case: base, expectedStatusCode: options.expectedStatusCode, middleware: mw}
}

func TestAuthorizer(t *testing.T) {
	tests := []*TestCase{
		NewTest(
			t,
			"authorized decisions should succeed",
			&testOptions{},
		),
		NewTest(
			t,
			"unauthorized decisions should err",
			&testOptions{
				Options: test.Options{
					Reject: true,
				},
				expectedStatusCode: http.StatusForbidden,
			},
		),
		NewTest(
			t,
			"policy mapper should override policy path",
			&testOptions{
				Options: test.Options{
					PolicyPath: test.OverridePolicyPath,
				},
				callback: func(mw *httpmw.Middleware) {
					mw.WithPolicyPathMapper(
						func(r *http.Request) string {
							return test.OverridePolicyPath
						},
					).Identity.Subject().ID(test.DefaultUsername)
				},
			},
		),
	}

	for _, test := range tests {
		t.Run(
			test.Case.Name,
			testCase(test),
		)
	}
}

func noopHandler(_ http.ResponseWriter, _ *http.Request) {}

func testCase(testCase *TestCase) func(*testing.T) {
	return func(t *testing.T) {
		handler := testCase.middleware.Handler(http.HandlerFunc(noopHandler))

		req := httptest.NewRequest("GET", "https://example.com/foo", http.NoBody)
		req.Header.Add("Authorization", test.DefaultUsername)

		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
	}
}
