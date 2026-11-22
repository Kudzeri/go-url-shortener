package tests

import (
	"net/http"
	"path"

	"github.com/Kudzeri/go-url-shortener/iternal/http-server/handlers/url/save"
	"github.com/Kudzeri/go-url-shortener/iternal/lib/api"
	"github.com/Kudzeri/go-url-shortener/iternal/lib/random"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

import (
	"net/url"
	"testing"
)

const host = "localhost:8081"

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	alias, _ := random.NewRandomString(10)
	e.POST("/url").WithJSON(save.Request{
		URL:   gofakeit.URL(),
		Alias: alias}).WithBasicAuth("admin", "12345").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("alias")
}
func TestURLShortener_SaveRedirectRemove(t *testing.T) {
	cases := []struct {
		name  string
		url   string
		alias string
		err   string
	}{
		{
			name:  "valid",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "invalid URL",
			url:   "invalid url",
			alias: gofakeit.Word(),
			err:   "Ошибка валидации",
		},
		{
			name:  "empty alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := httpexpect.Default(t, u.String())
			wantStatus := http.StatusOK
			if tc.err != "" {
				wantStatus = http.StatusBadRequest
			}

			resp := e.POST("/url").WithJSON(save.Request{
				URL:   tc.url,
				Alias: tc.alias,
			}).
				WithBasicAuth("admin", "12345").
				Expect().
				Status(wantStatus).
				JSON().
				Object()

			if tc.err != "" {
				resp.NotContainsKey("alias")
				resp.Value("error").String().IsEqual(tc.err)
				return
			}

			alias := tc.alias
			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			}

			testRedirect(t, alias, tc.url)

			e.DELETE("/"+path.Join("url", alias)).
				WithBasicAuth("admin", "12345").
				Expect().Status(http.StatusNoContent).
				Body().IsEmpty()

			testRedirectNotFound(t, alias)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path.Join("url", alias),
	}

	redirectToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectToURL)
}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path.Join("url", alias),
	}

	_, err := api.GetRedirect(u.String())
	require.ErrorIs(t, err, api.ErrInvalidStatusCode)
}
