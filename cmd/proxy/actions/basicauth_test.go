package actions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type checkAuthTest struct {
	name     string
	user     string
	pass     string
	httpUser string
	httpPass string
	want     bool
}

var checkAuthTests = [...]checkAuthTest{
	{
		name:     "success",
		user:     "athens",
		pass:     "1234",
		httpUser: "athens",
		httpPass: "1234",
		want:     true,
	},
	{
		name:     "username incorrect",
		user:     "athens",
		pass:     "1234",
		httpUser: "not athens",
		httpPass: "1234",
		want:     false,
	},
	{
		name:     "password incorrect",
		user:     "athens",
		pass:     "1234",
		httpUser: "athens",
		httpPass: "not 1234",
		want:     false,
	},
}

func TestCheckAuth(t *testing.T) {
	for _, testCase := range checkAuthTests {
		t.Run(testCase.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.SetBasicAuth(testCase.httpUser, testCase.httpPass)

			got := checkAuth(r, testCase.user, testCase.pass)
			require.Equal(t, got, testCase.want)
		})
	}
}
