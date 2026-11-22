package redirect

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kudzeri/go-url-shortener/iternal/lib/logger/handlers/slogdiscard"
	"github.com/Kudzeri/go-url-shortener/iternal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockURLGetter struct {
	mock.Mock
}

func (m *mockURLGetter) GetURL(alias string) (string, error) {
	args := m.Called(alias)
	return args.String(0), args.Error(1)
}

func TestHandler(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		storageError error
		wantStatus   int
		wantError    string
		target       string
		wantLocation string
	}{
		{
			name:         "success",
			alias:        "success",
			wantStatus:   http.StatusFound,
			target:       "https://example.com/docs?q=go",
			wantLocation: "https://example.com/docs?q=go",
		},
		{
			name:       "missing alias",
			wantStatus: http.StatusBadRequest,
			wantError:  "no alias provided",
		},
		{
			name:         "not found",
			alias:        "missing",
			storageError: fmt.Errorf("get URL: %w", storage.ErrNotFound),
			wantStatus:   http.StatusNotFound,
			wantError:    "url not found",
		},
		{
			name:         "storage error",
			alias:        "failure",
			storageError: errors.New("storage unavailable"),
			wantStatus:   http.StatusInternalServerError,
			wantError:    "internal server error",
		},
	}

	store := &mockURLGetter{}
	t.Cleanup(func() { store.AssertExpectations(t) })
	handler := New(slogdiscard.New(), store)
	router := gin.New()
	router.GET("/url/:alias", handler)
	router.GET("/url", handler)

	for _, tc := range cases {
		if tc.alias != "" {
			store.On("GetURL", tc.alias).Return(tc.target, tc.storageError).Once()
		}
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := "/url"
			if tc.alias != "" {
				path += "/" + tc.alias
			}
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))

			require.Equal(t, tc.wantStatus, rr.Code, rr.Body.String())
			require.Equal(t, tc.wantLocation, rr.Header().Get("Location"))
			if tc.wantError != "" {
				require.JSONEq(t, fmt.Sprintf(`{"status":"Error","error":%q}`, tc.wantError), rr.Body.String())
			}
		})
	}
}
