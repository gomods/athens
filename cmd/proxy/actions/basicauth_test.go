package actions

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gomods/athens/pkg/log"
	"github.com/sirupsen/logrus"
)

var basicAuthTests = [...]struct {
	name           string
	user           string
	pass           string
	path           string
	logs           string
	expectedStatus int
}{
	{
		name:           "happy_path",
		user:           "correctUser",
		pass:           "correctPass",
		path:           "/",
		logs:           "",
		expectedStatus: 200,
	},
	{
		name:           "incorrect_username",
		user:           "wrongUser",
		pass:           "correctPass",
		path:           "/",
		logs:           "",
		expectedStatus: 401,
	},
	{
		name:           "incorrect_password",
		user:           "correctUser",
		pass:           "wrongPassword",
		path:           "/",
		logs:           "",
		expectedStatus: 401,
	},
	{
		name:           "log_on_healthz",
		user:           "wrongUser",
		pass:           "wrongPassword",
		path:           "/healthz",
		logs:           "",
		expectedStatus: 200,
	},
	{
		name:           "log_on_readyz",
		user:           "wrongUser",
		pass:           "wrongPassword",
		path:           "/readyz",
		logs:           "",
		expectedStatus: 200,
	},
}

func TestBasicAuth(t *testing.T) {
	mwFunc := basicAuth("correctUser", "correctPass")
	handler := mwFunc(http.HandlerFunc(mockHandler))
	for _, tc := range basicAuthTests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tc.path, nil)
			r.SetBasicAuth(tc.user, tc.pass)
			lggr := log.New("none", logrus.DebugLevel, "")
			buf := &bytes.Buffer{}
			lggr.Out = buf
			ctx := log.SetEntryInContext(context.Background(), lggr)
			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Fatalf("expected http status to be %v but got %v", tc.expectedStatus, resp.StatusCode)
			}
			if !strings.Contains(buf.String(), tc.logs) {
				t.Fatalf("expected logs to include: %s but got: %s", tc.logs, buf.String())
			}
		})
	}
}

func mockHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
