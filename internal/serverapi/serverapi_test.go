package serverapi

import (
	"bytes"
	"compress/gzip"
	"context"
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
	postResponsePatternF = `^http:\/\/%s\/.+`
	postResponsePattern  = `^http:\/\/localhost:8080\/.+`
	targetURL            = "http://localhost:8080/"
	positiveURL          = "http://ya.ru"
	JSONBodyRequest      = `{"url":"http://mail.ru/"}`
	JSONPatternResponse  = `^{"result":"http:\/\/%s\/.+`
	JSONPathPattern      = "%s/api/shorten"
)

type TestConfig struct {
	url           string
	shortAddress  string
	fileStoreName string
	dbConnName    string
}

var tconf *TestConfig

func (c TestConfig) GetURL() string {
	return c.url
}

func (c TestConfig) GetShortAddress() string {
	return c.shortAddress
}
func (c TestConfig) GetFileStoreName() string {
	return c.fileStoreName
}
func (c TestConfig) GetDBConnName() string {
	return c.dbConnName
}

func initEnv() (serv *server, testserver *httptest.Server) {
	tconf = &TestConfig{
		url:           ":8080",
		shortAddress:  "http://localhost:8080/",
		fileStoreName: "/tmp/short-url-db.json"}

	storage, err := store.New(context.Background(), tconf)
	if err != nil {
		panic(err)
	}
	cut := cutter.New(storage)
	serv = New(cut, tconf)
	testserver = httptest.NewServer(serv.mux)
	tconf.shortAddress = testserver.URL
	return
}

type postRequest struct {
	httpMethod string
	body       io.Reader
	jsonHeader bool
}

type expectedPostResponse struct {
	code        int
	bodyPattern string
	bodyMessage string
}

func TestInitHandler(t *testing.T) {
	serv, testserver := initEnv()
	defer testserver.Close()
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
			code:        http.StatusMethodNotAllowed,
			bodyPattern: "",
			bodyMessage: ""},
	},
		{
			name: "InitHandler - Positive",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader(positiveURL)},
			expResp: expectedPostResponse{
				code:        http.StatusCreated,
				bodyPattern: fmt.Sprintf(postResponsePatternF, serv.config.GetShortAddress()[7:]),
				bodyMessage: ""},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.request.httpMethod, testserver.URL, tt.request.body)
			require.NoError(t, err)
			res, err := testserver.Client().Do(request)
			require.NoError(t, err)
			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")
			checkPostBody(res, t, tt.expResp.bodyPattern, tt.expResp.bodyMessage)
		})
	}
}

func TestCutterHandler(t *testing.T) {
	serv, testserver := initEnv()
	defer testserver.Close()
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
			code:        http.StatusMethodNotAllowed,
			bodyPattern: "",
			bodyMessage: ""},
	},
		{
			name: "negative - empty Body",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader("")},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyPattern: "",
				bodyMessage: "cutterHandler: empty body not expected"},
		},
		{
			name: "negative - not url",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader("==fsaw=ae")},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyPattern: "",
				bodyMessage: `cutterHandler: parsing URI: ==fsaw=ae : parse "==fsaw=ae": invalid URI for request`},
		},
		{
			name: "positive",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader(positiveURL)},
			expResp: expectedPostResponse{
				code:        http.StatusCreated,
				bodyPattern: fmt.Sprintf(postResponsePatternF, serv.config.GetShortAddress()[7:]),
				bodyMessage: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.request.httpMethod, testserver.URL, tt.request.body)
			require.NoError(t, err)
			res, err := testserver.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")
			checkPostBody(res, t, tt.expResp.bodyPattern, tt.expResp.bodyMessage)
		})
	}
}

func checkPostBody(res *http.Response, t *testing.T, wantedPattern string, wantedMessage string) {
	var resBody string
	if res.Header.Get("Content-Encoding") == "gzip" {
		resBody = unzipBody(t, res.Body)
	} else {
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		resBody = string(b)

	}
	err := res.Body.Close()
	require.NoError(t, err)
	if res.StatusCode == http.StatusCreated || res.StatusCode == http.StatusConflict {
		assert.Regexpf(t, regexp.MustCompile(wantedPattern), string(resBody), "body must be like %s", wantedPattern)
	} else {
		assert.Equal(t, wantedMessage, string(resBody))
	}
}

func doCut(t *testing.T, servStruct *server, testserver *httptest.Server) (string, error) {
	request, err := http.NewRequest(http.MethodPost, testserver.URL, strings.NewReader(positiveURL))
	require.NoError(t, err)
	res, err := testserver.Client().Do(request)
	require.NoError(t, err)
	defer res.Body.Close()
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
	serv, testserver := initEnv()
	defer testserver.Close()
	redirectedURL, err := doCut(t, serv, testserver)
	require.NoError(t, err)
	tests := []struct {
		name    string
		request redirectRequest
		expResp expectedResponse
	}{
		{
			name: "negative - wrong method",
			request: redirectRequest{
				httpMethod: http.MethodPut,
				url:        testserver.URL,
			},
			expResp: expectedResponse{
				code:        http.StatusMethodNotAllowed,
				bodyMessage: ""},
		},
		{
			name: "negative - notfound",
			request: redirectRequest{
				httpMethod: http.MethodGet,
				url:        fmt.Sprintf("%s/C222", testserver.URL),
			},
			expResp: expectedResponse{
				code:        http.StatusBadRequest,
				bodyMessage: "redirectHandler: fetching url fo redirect: getKeyByValue: while getting value by key:C222: no data found in urlMap for value C222"},
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
			request, err := http.NewRequest(tt.request.httpMethod, tt.request.url, nil)
			require.NoError(t, err)
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			res, err := client.Do(request)

			require.NoError(t, err)
			defer res.Body.Close()
			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.expResp.bodyMessage, string(resBody))
		})
	}
}

func TestCutterJSONHandler(t *testing.T) {
	serv, testserver := initEnv()
	defer testserver.Close()
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
			code: http.StatusMethodNotAllowed},
	},
		{
			name: "negative - no json header",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader("JSONBodyRequest")},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: `cutterJsonHandler: content-type have to be application/json`},
		},
		{
			name: "negative - empty Body",
			request: postRequest{
				httpMethod: http.MethodPost,
				jsonHeader: true,
				body:       strings.NewReader("")},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: "cutterJsonHandler: decoding request: EOF"},
		},

		{
			name: "positive",
			request: postRequest{
				httpMethod: http.MethodPost,
				jsonHeader: true,
				body:       strings.NewReader(JSONBodyRequest)},
			expResp: expectedPostResponse{
				code:        http.StatusCreated,
				bodyPattern: fmt.Sprintf(JSONPatternResponse, serv.config.GetShortAddress()[7:]),
				bodyMessage: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.request.httpMethod, fmt.Sprintf(JSONPathPattern, testserver.URL), tt.request.body)
			if tt.request.jsonHeader {
				request.Header.Set("Content-Type", "application/json")
			}
			require.NoError(t, err)
			res, err := testserver.Client().Do(request)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")
			checkPostBody(res, t, tt.expResp.bodyPattern, tt.expResp.bodyMessage)
		})
	}
}
func TestCompression(t *testing.T) {
	_, testserver := initEnv()
	defer testserver.Close()
	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(JSONBodyRequest))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)
		r, err := http.NewRequest("POST", fmt.Sprintf(JSONPathPattern, testserver.URL), buf)
		require.NoError(t, err)

		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := testserver.Client().Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		_, err = io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))
	})

}

func unzipBody(t *testing.T, body io.ReadCloser) string {
	zr, err := gzip.NewReader(body)
	require.NoError(t, err)
	b, err := io.ReadAll(zr)
	require.NoError(t, err)
	return string(b)
}
