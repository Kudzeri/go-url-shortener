package save

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kudzeri/go-url-shortener/iternal/lib/logger/handlers/slogdiscard"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name       string
		alias      string
		url        string
		wantStatus int
		respError  string
		mockError  error
	}{
		{
			name:       "Success",
			alias:      "testalias",
			url:        "https://google.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Empty alias",
			alias:      "",
			url:        "https://google.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Empty URL",
			url:        "",
			alias:      "somealias",
			wantStatus: http.StatusBadRequest,
			respError:  "Ошибка валидации",
		},
		{
			name:       "Invalid URL",
			url:        "some invalid URL",
			alias:      "somealias",
			wantStatus: http.StatusBadRequest,
			respError:  "Ошибка валидации",
		},
		{
			name:       "SaveURL Error",
			alias:      "testalias",
			url:        "https://google.com",
			wantStatus: http.StatusInternalServerError,
			respError:  "failed to save URL",
			mockError:  errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := NewMockURLSaver(t)

			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()
			}

			router := gin.New()
			router.POST("/url", New(slogdiscard.New(), urlSaverMock))

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.wantStatus, rr.Code, rr.Body.String())

			body := rr.Body.String()

			var resp Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)

			// TODO: add more checks
		})
	}
}
