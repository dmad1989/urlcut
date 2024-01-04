package serverapi

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	postResponsePattern = `^http:\/\/localhost:8080\/\w+`
	targetURL           = "http://localhost:8080/"
	positiveURL         = "http://ya.ru"
)

func initEnv() *server {
	storage := store.New()
	cut := cutter.New(storage)
	return New(cut)
}

type postRequest struct {
	httpMethod string
	body       io.Reader
}

type expectedPostResponse struct {
	code        int
	bodyPattern string
	bodyMessage string
}

func TestInitHandler(t *testing.T) {
	tests := []struct {
		name    string
		request postRequest
		expResp expectedPostResponse
	}{{
		name: "InitHandler - Negative - wrong method",
		request: postRequest{
			httpMethod: http.MethodPut,
			body:       strings.NewReader("")},
		expResp: expectedPostResponse{
			code:        http.StatusBadRequest,
			bodyPattern: "",
			bodyMessage: "wrong http method"},
	},
		{
			name: "InitHandler - Positive",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader(positiveURL)},
			expResp: expectedPostResponse{
				code:        http.StatusCreated,
				bodyPattern: postResponsePattern,
				bodyMessage: ""},
		}}
	serv := initEnv()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.httpMethod, targetURL, tt.request.body)
			w := httptest.NewRecorder()

			serv.initHandlers(w, request)
			res := w.Result()
			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")

			checkPostBody(res, t, tt.expResp.bodyPattern, tt.expResp.bodyMessage)
		})
	}
}

func TestCutterHandler(t *testing.T) {
	tests := []struct {
		name    string
		request postRequest
		expResp expectedPostResponse
	}{{
		name: "negative - wrong method",
		request: postRequest{
			httpMethod: http.MethodGet,
			body:       strings.NewReader("")},
		expResp: expectedPostResponse{
			code:        http.StatusBadRequest,
			bodyPattern: "",
			bodyMessage: "wrong http method"},
	},
		{
			name: "negative - empty Body",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader("")},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyPattern: "",
				bodyMessage: "empty body not expected"},
		},
		{
			name: "negative - not url",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader("==fsaw=ae")},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyPattern: "",
				bodyMessage: `parse "==fsaw=ae": invalid URI for request`},
		},
		{
			name: "positive",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader(positiveURL)},
			expResp: expectedPostResponse{
				code:        http.StatusCreated,
				bodyPattern: postResponsePattern,
				bodyMessage: ""},
		},
	}
	serv := initEnv()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.httpMethod, targetURL, tt.request.body)
			w := httptest.NewRecorder()
			serv.cutterHandler(w, request)
			res := w.Result()
			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")
			checkPostBody(res, t, tt.expResp.bodyPattern, tt.expResp.bodyMessage)
		})
	}
}

func checkPostBody(res *http.Response, t *testing.T, wantedPattern string, wantedMessage string) {
	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)
	assert.NotEmpty(t, resBody)
	if res.StatusCode == http.StatusCreated {
		assert.Regexpf(t, regexp.MustCompile(wantedPattern), string(resBody), "body must be like %s", wantedPattern)
	} else {
		assert.Equal(t, wantedMessage, string(resBody))
	}
}

func doCut(serv *server) (string, error) {
	reqPost := httptest.NewRequest(http.MethodPost, targetURL, strings.NewReader(positiveURL))
	postRecoder := httptest.NewRecorder()
	serv.cutterHandler(postRecoder, reqPost)
	res := postRecoder.Result()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	err = res.Body.Close()
	if err != nil {
		return "", err
	}
	return string(resBody), nil
}
func TestRedirectHandler(t *testing.T) {

	type redirectRequest struct {
		httpMethod string
		url        string
	}
	type expectedResponse struct {
		code        int
		bodyMessage string
	}

	serv := initEnv()
	redirectedURL, err := doCut(serv)
	require.NoError(t, err)

	tests := []struct {
		name    string
		request redirectRequest
		expResp expectedResponse
	}{{
		name: "negative - wrong method",
		request: redirectRequest{
			httpMethod: http.MethodPost,
			url:        "http://localhost:8080/",
		},
		expResp: expectedResponse{
			code:        http.StatusBadRequest,
			bodyMessage: "wrong http method"},
	},
		{
			name: "negative - empty path",
			request: redirectRequest{
				httpMethod: http.MethodGet,
				url:        "http://localhost:8080/",
			},
			expResp: expectedResponse{
				code:        http.StatusBadRequest,
				bodyMessage: "url path is empty"},
		},

		{
			name: "negative - empty path",
			request: redirectRequest{
				httpMethod: http.MethodGet,
				url:        "http://localhost:8080/222",
			},
			expResp: expectedResponse{
				code:        http.StatusBadRequest,
				bodyMessage: "requested url not found"},
		},
		{
			name: "positive",
			request: redirectRequest{
				httpMethod: http.MethodGet,
				url:        redirectedURL,
			},
			expResp: expectedResponse{
				code:        http.StatusTemporaryRedirect,
				bodyMessage: fmt.Sprintf("<a href=\"%s\">Temporary Redirect</a>.\n\n", positiveURL)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(tt.request.httpMethod, tt.request.url, nil)
			serv.redirectHandler(w, request)
			res := w.Result()
			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)
			assert.NotEmpty(t, resBody)
			assert.Equal(t, tt.expResp.bodyMessage, string(resBody))
		})
	}
}
