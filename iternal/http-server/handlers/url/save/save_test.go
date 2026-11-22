package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	resp "github.com/Kudzeri/go-url-shortener/iternal/lib/api/response"
	"github.com/Kudzeri/go-url-shortener/iternal/lib/logger/handlers/slogdiscard"
	"github.com/Kudzeri/go-url-shortener/iternal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	const target = "https://example.com/docs"
	const alias = "testalias"
	const validBody = `{"url":"https://example.com/docs","alias":"testalias"}`
	const generatedAliasPattern = `^[a-zA-Z0-9]{6}$`

	cases := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
		wantFields map[string]string
		mockAlias  string
		autoAlias  bool
		storeError error
	}{
		{
			name:       "explicit alias",
			body:       validBody,
			wantStatus: http.StatusOK,
			mockAlias:  alias,
		},
		{
			name:       "omitted alias",
			body:       `{"url":"https://example.com/docs"}`,
			wantStatus: http.StatusOK,
			autoAlias:  true,
		},
		{
			name:       "empty alias",
			body:       `{"url":"https://example.com/docs","alias":""}`,
			wantStatus: http.StatusOK,
			autoAlias:  true,
		},
		{
			name:       "missing URL",
			body:       `{"alias":"testalias"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Ошибка валидации",
			wantFields: map[string]string{"url": "Поле URL является обязательным"},
		},
		{
			name:       "invalid URL",
			body:       `{"url":"not a URL"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Ошибка валидации",
			wantFields: map[string]string{"url": "Укажите корректный URL"},
		},
		{
			name:       "short alias",
			body:       `{"url":"https://example.com/docs","alias":"ab"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Ошибка валидации",
			wantFields: map[string]string{"alias": "Alias должен содержать от 3 до 32 символов"},
		},
		{
			name:       "long alias",
			body:       `{"url":"https://example.com/docs","alias":"` + strings.Repeat("a", 33) + `"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Ошибка валидации",
			wantFields: map[string]string{"alias": "Alias должен содержать от 3 до 32 символов"},
		},
		{
			name:       "non-alphanumeric alias",
			body:       `{"url":"https://example.com/docs","alias":"bad-alias"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "Ошибка валидации",
			wantFields: map[string]string{"alias": "Alias может содержать латинские буквы и цифры"},
		},
		{
			name:       "malformed JSON",
			body:       `{"url":`,
			wantStatus: http.StatusBadRequest,
			wantError:  "waiting correct json format",
		},
		{
			name:       "wrong field type",
			body:       `{"url":42}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "waiting correct json format",
		},
		{
			name:       "already exists",
			body:       validBody,
			wantStatus: http.StatusConflict,
			wantError:  "URL already exists",
			mockAlias:  alias,
			storeError: fmt.Errorf("insert: %w", storage.ErrAlreadyExists),
		},
		{
			name:       "storage error",
			body:       validBody,
			wantStatus: http.StatusInternalServerError,
			wantError:  "failed to save URL",
			mockAlias:  alias,
			storeError: errors.New("storage unavailable"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			store := NewMockURLSaver(t)
			wantAlias := tc.mockAlias

			switch {
			case tc.autoAlias:
				store.EXPECT().SaveURL(target, mock.IsType("")).
					Run(func(_, generated string) {
						require.Regexp(t, generatedAliasPattern, generated)
						wantAlias = generated
					}).
					Return(1, nil).Once()
			case tc.mockAlias != "":
				store.EXPECT().SaveURL(target, tc.mockAlias).Return(1, tc.storeError).Once()
			}

			router := gin.New()
			router.POST("/url", New(slogdiscard.New(), store))
			req := httptest.NewRequest(http.MethodPost, "/url", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, tc.wantStatus, rr.Code, rr.Body.String())
			var got Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
			if tc.wantStatus == http.StatusOK {
				require.Equal(t, Response{Response: *resp.OK(), Alias: wantAlias}, got)
				return
			}

			require.Equal(t, Response{Response: resp.Response{
				Status: resp.StatusError,
				Error:  tc.wantError,
				Fields: tc.wantFields,
			}}, got)
		})
	}
}
