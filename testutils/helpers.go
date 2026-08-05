package testutils

import (
	"net/http"
	"os"
	"testing"

	"golang.org/x/oauth2"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func GetRecorder(t *testing.T) *recorder.Recorder {
	t.Helper()

	fixtureName := "testdata/" + t.Name()

	hook := func(i *cassette.Interaction) error {
		if i.Request.Headers != nil && i.Request.Headers.Get("Authorization") != "" {
			i.Request.Headers.Set("Authorization", "REDACTED")
		}

		return nil
	}

	var opts []recorder.Option
	opts = []recorder.Option{
		recorder.WithRealTransport(&oauth2.Transport{
			Base: http.DefaultTransport,
			Source: oauth2.ReuseTokenSource(nil, oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: os.Getenv("BACKSTAGE_TOKEN"),
					TokenType:   "Bearer",
				},
			)),
		}),
		recorder.WithHook(hook, recorder.BeforeSaveHook),
		recorder.WithMatcher(cassette.MatcherFunc(func(r1 *http.Request, r2 cassette.Request) bool {
			// doesn't match automatically when providing real transport
			return r1.URL.String() == r2.URL
		})),
		recorder.WithMode(recorder.ModeRecordOnce),
	}

	r, err := recorder.New(fixtureName, opts...)
	if err != nil {
		t.Fatal(err)
	}

	return r
}
