package delete

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

type mockURLDeleter struct {
	mock.Mock
}

func (m *mockURLDeleter) DeleteUrl(alias string) error {
	args := m.Called(alias)
	return args.Error(0)
}

func TestHandler(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		storageError error
		wantStatus   int
		wantError    string
	}{{
		name:       "success",
		alias:      "success",
		wantStatus: http.StatusNoContent,
	}, {
		name:       "missing alias",
		wantStatus: http.StatusBadRequest,
		wantError:  "no alias provided",
	}, {
		name:         "not found",
		alias:        "missing",
		storageError: fmt.Errorf("delete URL: %w", storage.ErrNotFound),
		wantStatus:   http.StatusNotFound,
		wantError:    "alias not found",
	}, {
		name:         "storage error",
		alias:        "failure",
		storageError: errors.New("storage unavailable"),
		wantStatus:   http.StatusInternalServerError,
		wantError:    "",
	}}

	store := &mockURLDeleter{}
	t.Cleanup(func() { store.AssertExpectations(t) })
	handler := New(slogdiscard.New(), store)
	router := gin.New()
	router.DELETE("/url/:alias", handler)
	// Register a route without a parameter to exercise the handler's validation.
	router.DELETE("/url", handler)

	for _, tc := range cases {
		if tc.alias != "" {
			store.On("DeleteUrl", tc.alias).Return(tc.storageError).Once()
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
			router.ServeHTTP(rr, httptest.NewRequest(http.MethodDelete, path, nil))

			require.Equal(t, tc.wantStatus, rr.Code, rr.Body.String())
			if tc.wantError != "" {
				require.JSONEq(t, fmt.Sprintf(`{"status":"Error","error":%q}`, tc.wantError), rr.Body.String())
			} else {
				require.Empty(t, rr.Body.String())
			}
		})
	}
}
