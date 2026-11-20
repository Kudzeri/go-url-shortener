package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kudzeri/go-url-shortener/iternal/lib/logger/handlers/slogdiscard"
	"github.com/Kudzeri/go-url-shortener/iternal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveHandler_ValidRequest(t *testing.T) {
	for _, tc := range []struct {
		name  string
		body  string
		alias string
	}{
		{"explicit alias", `{"url":"https://google.com","alias":"google"}`, "google"},
		{"minimum length", `{"url":"https://google.com","alias":"Ab1"}`, "Ab1"},
		{"maximum length", `{"url":"https://google.com","alias":"0123456789abcdefghijklmnopqrstuv"}`, "0123456789abcdefghijklmnopqrstuv"},
		{"omitted alias", `{"url":"https://google.com"}`, ""},
		{"empty alias", `{"url":"https://google.com","alias":""}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			saver := NewMockURLSaver(t)
			wantAlias := tc.alias
			var aliasArgument any = tc.alias
			if tc.alias == "" {
				aliasArgument = mock.MatchedBy(func(alias string) bool { return len(alias) == aliasLength })
			}
			saver.EXPECT().SaveURL("https://google.com", aliasArgument).
				Run(func(_ string, alias string) { wantAlias = alias }).
				Return(1, nil).Once()

			recorder := postURL(t, New(slogdiscard.New(), saver), tc.body)
			require.Equal(t, http.StatusOK, recorder.Code)
			require.Regexp(t, `^[a-zA-Z0-9]+$`, wantAlias)
			require.JSONEq(t, fmt.Sprintf(`{"status":"OK","alias":%q}`, wantAlias), recorder.Body.String())
		})
	}
}

func TestSaveHandler_InvalidRequest(t *testing.T) {
	const badJSON = "Ожидается корректный JSON с полями нужных типов"
	const requiredURL = "Поле URL является обязательным"
	const invalidURL = "Укажите корректный URL"
	const aliasLengthError = "Alias должен содержать от 3 до 32 символов"
	const aliasCharactersError = "Alias может содержать латинские буквы и цифры"

	for _, tc := range []struct {
		name   string
		body   string
		fields map[string]string
	}{
		{"empty body", "", nil},
		{"malformed JSON", `{"url":`, nil},
		{"array JSON", `[]`, nil},
		{"URL is a number", `{"url":123}`, nil},
		{"alias is a boolean", `{"url":"https://google.com","alias":true}`, nil},
		{"missing URL", `{}`, map[string]string{"url": requiredURL}},
		{"empty URL", `{"url":""}`, map[string]string{"url": requiredURL}},
		{"URL without scheme", `{"url":"google.com"}`, map[string]string{"url": invalidURL}},
		{"non-HTTP URL", `{"url":"ftp://google.com"}`, map[string]string{"url": invalidURL}},
		{"alias too short", `{"url":"https://google.com","alias":"ab"}`, map[string]string{"alias": aliasLengthError}},
		{"alias too long", `{"url":"https://google.com","alias":"0123456789abcdefghijklmnopqrstuvw"}`, map[string]string{"alias": aliasLengthError}},
		{"alias contains punctuation", `{"url":"https://google.com","alias":"go-ogle"}`, map[string]string{"alias": aliasCharactersError}},
		{"alias contains Cyrillic", `{"url":"https://google.com","alias":"гугл"}`, map[string]string{"alias": aliasCharactersError}},
		{"both fields invalid", `{"url":"google.com","alias":"ab"}`, map[string]string{"url": invalidURL, "alias": aliasLengthError}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			saver := NewMockURLSaver(t)
			recorder := postURL(t, New(slogdiscard.New(), saver), tc.body)

			require.Equal(t, http.StatusBadRequest, recorder.Code)
			want := map[string]any{"status": "Error", "error": badJSON}
			if tc.fields != nil {
				want["error"] = "Ошибка валидации"
				want["fields"] = tc.fields
			}
			wantJSON, err := json.Marshal(want)
			require.NoError(t, err)
			require.JSONEq(t, string(wantJSON), recorder.Body.String())
			saver.AssertNumberOfCalls(t, "SaveURL", 0)
		})
	}
}

func TestSaveHandler_StorageErrors(t *testing.T) {
	for _, tc := range []struct {
		name         string
		saveErr      error
		wantStatus   int
		wantResponse string
	}{
		{"already exists", storage.ErrAlreadyExists, http.StatusConflict, `{"status":"Error","error":"URL already exists"}`},
		{"wrapped conflict", fmt.Errorf("save URL: %w", storage.ErrAlreadyExists), http.StatusConflict, `{"status":"Error","error":"URL already exists"}`},
		{"storage unavailable", errors.New("storage unavailable"), http.StatusInternalServerError, `{"status":"Error","error":"failed to save URL"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			saver := NewMockURLSaver(t)
			saver.EXPECT().SaveURL("https://google.com", "google").Return(0, tc.saveErr).Once()

			recorder := postURL(t, New(slogdiscard.New(), saver), `{"url":"https://google.com","alias":"google"}`)
			require.Equal(t, tc.wantStatus, recorder.Code)
			require.JSONEq(t, tc.wantResponse, recorder.Body.String())
		})
	}
}

func TestSaveHandler_AliasGenerationError(t *testing.T) {
	saver := NewMockURLSaver(t)
	generateAlias := func(length int) (string, error) {
		require.Equal(t, aliasLength, length)
		return "", errors.New("alias generation failed")
	}
	handler := newHandler(slogdiscard.New(), saver, generateAlias)

	recorder := postURL(t, handler, `{"url":"https://google.com"}`)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.JSONEq(t, `{"status":"Error","error":"Внутренняя ошибка сервера"}`, recorder.Body.String())
	saver.AssertNumberOfCalls(t, "SaveURL", 0)
}

func postURL(t *testing.T, handler gin.HandlerFunc, body string) *httptest.ResponseRecorder {
	t.Helper()

	router := gin.New()
	router.POST("/url", handler)
	req := httptest.NewRequest(http.MethodPost, "/url", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
