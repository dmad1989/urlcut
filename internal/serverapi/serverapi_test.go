package serverapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmad1989/urlcut/internal/config"
	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/dmad1989/urlcut/internal/mocks"
	"github.com/dmad1989/urlcut/internal/store"
)

const (
	postResponsePatternF = `^http:\/\/%s\/.+`
	postResponsePattern  = `^http:\/\/localhost:8080\/.+`
	targetURL            = "http://localhost:8080/"
	positiveURL          = "http://ya.ru"
	JSONBodyRequest      = `{"url":"http://mail.ru/"}`
	JSONResponse         = `{"result":"%s/%s"}`
	JSONPathPattern      = "%s/api/shorten"
	JSONBatchPathPattern = "%s/api/shorten/batch"
)

var (
	errorReader = errors.New("test reader")
	errorUnique cutter.UniqueURLError
)

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errorReader
}

type TestConfig struct {
	url           string
	shortAddress  string
	fileStoreName string
	dbConnName    string
	trustedSubnet string
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
func (c TestConfig) GetUserContextKey() config.ContextKey {
	return config.ContextKey{}
}

func (c TestConfig) GetEnableHTTPS() bool {
	return false
}

func (c TestConfig) GetTrustedSubnet() string {
	return c.trustedSubnet
}

func initEnv() (serv *Server, testserver *httptest.Server) {
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
	errorUnique.Code = "notunique"
	return
}

type postRequest struct {
	body       io.Reader
	httpMethod string
	jsonHeader bool
}

type expectedPostResponse struct {
	bodyPattern string
	bodyMessage string
	code        int
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
		expResp expectedPostResponse
		name    string
		request postRequest
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
		// {
		// 	name: "negative - error Read Body",
		// 	request: postRequest{
		// 		httpMethod: http.MethodPost,
		// 		body:       errReader(0)},
		// 	expResp: expectedPostResponse{
		// 		code:        http.StatusBadRequest,
		// 		bodyMessage: strings.Join([]string{"cutterHandler: reading request body: ", errorReader.Error()}, ""),
		// 	},
		// },
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
			defer func() {
				require.NoError(t, res.Body.Close())
			}()
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

func doCut(t *testing.T, testserver *httptest.Server) (string, error) {
	request, err := http.NewRequest(http.MethodPost, testserver.URL, strings.NewReader(positiveURL))
	require.NoError(t, err)
	res, err := testserver.Client().Do(request)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, res.Body.Close())
	}()
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
		bodyMessage string
		code        int
	}
	_, testserver := initEnv()
	defer testserver.Close()
	redirectedURL, err := doCut(t, testserver)
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
			defer func() {
				require.NoError(t, res.Body.Close())
			}()
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serv, testserver := initEnv()
	defer testserver.Close()
	url := fmt.Sprintf(JSONPathPattern, testserver.URL)
	type mockParams struct {
		shortAddress string
		cutterResult string
		cutterError  error
	}

	tests := []struct {
		name    string
		request postRequest
		expResp expectedPostResponse
		mock    mockParams
	}{
		{
			name: "negative - no json header",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader("JSONBodyRequest")},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: `cutterJsonHandler: content-type have to be application/json`},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress()[7:],
				cutterResult: "",
				cutterError:  nil},
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
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress()[7:],
				cutterResult: "",
				cutterError:  nil},
		},

		{
			name: "negative - error Read Body",
			request: postRequest{
				httpMethod: http.MethodPost,
				jsonHeader: true,
				body:       errReader(0)},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: strings.Join([]string{"cutterJsonHandler: reading request body: ", errorReader.Error()}, ""),
			},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress()[7:],
				cutterResult: "",
				cutterError:  nil},
		},

		{
			name: "negative - error from cutter",
			request: postRequest{
				httpMethod: http.MethodPost,
				body: strings.NewReader(`
					{
						"correlation_id": "1",
						"original_url": "http://yaga.ru/"
					}`),
				jsonHeader: true},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: "cutterJsonHandler: getting code for url: from cutter"},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress()[7:],
				cutterResult: "",
				cutterError:  errors.New("from cutter")},
		},
		{
			name: "negative - Unique error from cutter",
			request: postRequest{
				httpMethod: http.MethodPost,
				body: strings.NewReader(`
					{
						"correlation_id": "1",
						"original_url": "http://yaga.ru/"
					}`),
				jsonHeader: true},
			expResp: expectedPostResponse{
				code:        http.StatusConflict,
				bodyMessage: fmt.Sprintf(JSONResponse, serv.config.GetShortAddress(), "notunique")},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress(),
				cutterResult: "notunique",
				cutterError:  &errorUnique,
			},
		},
		{
			name: "positive",
			request: postRequest{
				httpMethod: http.MethodPost,
				jsonHeader: true,
				body:       strings.NewReader(JSONBodyRequest)},
			expResp: expectedPostResponse{
				code:        http.StatusCreated,
				bodyMessage: fmt.Sprintf(JSONResponse, serv.config.GetShortAddress(), "positive")},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress(),
				cutterResult: "positive",
				cutterError:  nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := mocks.NewMockICutter(ctrl)
			c := mocks.NewMockConfiger(ctrl)
			c.EXPECT().GetShortAddress().Return(tt.mock.shortAddress).MaxTimes(1)
			a.EXPECT().Cut(gomock.Any(), gomock.Any()).Return(tt.mock.cutterResult, tt.mock.cutterError).MaxTimes(1)

			s := New(a, c)
			request, err := http.NewRequest(tt.request.httpMethod, url, tt.request.body)
			if tt.request.jsonHeader {
				request.Header.Set("Content-Type", "application/json")
			}
			require.NoError(t, err)

			w := httptest.NewRecorder()
			s.cutterJSONHandler(w, request)
			res := w.Result()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, res.Body.Close())
			}()

			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")

			b, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			resBody := string(b)
			err = res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.expResp.bodyMessage, string(resBody))
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

		defer func() {
			require.NoError(t, resp.Body.Close())
		}()

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

func TestCutterJSONBatchHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serv, testserver := initEnv()
	defer testserver.Close()
	url := fmt.Sprintf(JSONBatchPathPattern, testserver.URL)

	type mockParams struct {
		//shortAddressTimes int
		shortAddress string
		uploadError  error
		uploadResult jsonobject.Batch
	}
	tests := []struct {
		name    string
		request postRequest
		expResp expectedPostResponse
		mock    mockParams
	}{
		{
			name: "negative - no json header",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader(""),
				jsonHeader: false},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: "JSONBatchHandler: content-type have to be application/json"},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress()[7:],
				uploadResult: jsonobject.Batch{},
				uploadError:  nil},
		},
		{
			name: "negative - empty Body",
			request: postRequest{
				httpMethod: http.MethodPost,
				body:       strings.NewReader(""),
				jsonHeader: true},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: "JSONBatchHandler: decoding request: EOF"},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress()[7:],
				uploadResult: jsonobject.Batch{},
				uploadError:  nil},
		},
		{
			name: "negative - error Read Body",
			request: postRequest{
				httpMethod: http.MethodPost,
				jsonHeader: true,
				body:       errReader(0)},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: strings.Join([]string{"JSONBatchHandler: reading request body: ", errorReader.Error()}, ""),
			},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress()[7:],
				uploadResult: jsonobject.Batch{},
				uploadError:  nil},
		},
		{
			name: "negative - error from cutter",
			request: postRequest{
				httpMethod: http.MethodPost,
				body: strings.NewReader(`[
					{
						"correlation_id": "1",
						"original_url": "http://yaga.ru/"
					}]`),
				jsonHeader: true},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: "JSONBatchHandler: getting code for url: from cutter"},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress()[7:],
				uploadResult: jsonobject.Batch{},
				uploadError:  errors.New("from cutter")},
		},
		{
			name: "positive",
			request: postRequest{
				httpMethod: http.MethodPost,
				body: strings.NewReader(`[
					{
						"correlation_id": "1",
						"original_url": "http://yaga.ru/"
					}]`),
				jsonHeader: true},
			expResp: expectedPostResponse{
				code: http.StatusCreated,
				bodyMessage: `[
					{
						"correlation_id": "1",
						"original_url": "http://yaga.ru/",
						
					}]`},
			mock: mockParams{
				shortAddress: serv.config.GetShortAddress()[7:],
				uploadResult: jsonobject.Batch{jsonobject.BatchItem{ID: "1", OriginalURL: "", ShortURL: "tt"}},
				uploadError:  nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//init mocks
			a := mocks.NewMockICutter(ctrl)
			c := mocks.NewMockConfiger(ctrl)
			c.EXPECT().GetShortAddress().Return(tt.mock.shortAddress).MaxTimes(1)
			a.EXPECT().UploadBatch(gomock.Any(), gomock.Any()).Return(tt.mock.uploadResult, tt.mock.uploadError).MaxTimes(1)
			s := New(a, c)
			//init request
			request, err := http.NewRequest(tt.request.httpMethod, url, tt.request.body)
			if tt.request.jsonHeader {
				request.Header.Set("Content-Type", "application/json")
			}
			require.NoError(t, err)

			w := httptest.NewRecorder()
			s.cutterJSONBatchHandler(w, request)
			res := w.Result()
			defer func() {
				require.NoError(t, res.Body.Close())
			}()
			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")
			checkPostBody(res, t, tt.expResp.bodyPattern, tt.expResp.bodyMessage)
		})
	}
}

func TestUserUrlsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	batches := func(cap int) jsonobject.Batch {
		batch := make(jsonobject.Batch, 0, cap)
		for i := 0; i < cap; i++ {
			batch = append(batch, jsonobject.BatchItem{ID: "id", OriginalURL: "url"})
		}
		return batch
	}

	serv, testserver := initEnv()
	defer testserver.Close()
	sAddr := serv.config.GetShortAddress()[7:]
	url := fmt.Sprintf("%s/api/user/urls", testserver.URL)
	type mockParams struct {
		shortAddressTimes int
		shortAddress      string
		getUrlsError      error
		getURLResult      jsonobject.Batch
	}

	type request struct {
		ctx context.Context
	}

	tests := []struct {
		name    string
		r       request
		expResp expectedPostResponse
		mock    mockParams
	}{
		{
			name: "negative - error in context",
			r: request{
				ctx: context.WithValue(
					context.Background(),
					config.ErrorCtxKey,
					errors.New("in context")),
			},
			expResp: expectedPostResponse{
				code:        http.StatusUnauthorized,
				bodyMessage: "in context\n"},
			mock: mockParams{
				shortAddressTimes: 1,
				shortAddress:      sAddr,
				getUrlsError:      nil,
				getURLResult:      jsonobject.Batch{},
			},
		},
		{
			name: "negative - error no rows from getUserUrls",
			r: request{
				ctx: context.Background(),
			},
			expResp: expectedPostResponse{
				code:        http.StatusNoContent,
				bodyMessage: ""},
			mock: mockParams{
				shortAddressTimes: 1,
				shortAddress:      sAddr,
				getUrlsError:      sql.ErrNoRows,
				getURLResult:      jsonobject.Batch{},
			},
		},
		{
			name: "negative - other error from getUserUrls",
			r: request{
				ctx: context.Background(),
			},
			expResp: expectedPostResponse{
				code:        http.StatusBadRequest,
				bodyMessage: "userUrlsHandler: getting all urls: other problem"},
			mock: mockParams{
				shortAddressTimes: 1,
				shortAddress:      sAddr,
				getUrlsError:      errors.New("other problem"),
				getURLResult:      jsonobject.Batch{},
			},
		},
		{
			name: "negative - no result from getUserUrls",
			r: request{
				ctx: context.Background(),
			},
			expResp: expectedPostResponse{
				code:        http.StatusNoContent,
				bodyMessage: ""},
			mock: mockParams{
				shortAddressTimes: 1,
				shortAddress:      sAddr,
				getUrlsError:      nil,
				getURLResult:      jsonobject.Batch{},
			},
		},
		{
			name: "positive - 1 row from getUserUrls",
			r: request{
				ctx: context.Background(),
			},
			expResp: expectedPostResponse{
				code:        http.StatusOK,
				bodyMessage: fmt.Sprintf("[{\"correlation_id\":\"id\",\"original_url\":\"url\",\"short_url\":\"%s/\"}]", sAddr)},
			mock: mockParams{
				shortAddressTimes: 1,
				shortAddress:      sAddr,
				getUrlsError:      nil,
				getURLResult:      batches(1),
			},
		},
		{
			name: "positive - 5 row from getUserUrls",
			r: request{
				ctx: context.Background(),
			},
			expResp: expectedPostResponse{
				code:        http.StatusOK,
				bodyMessage: fmt.Sprintf("[{\"correlation_id\":\"id\",\"original_url\":\"url\",\"short_url\":\"%s/\"},{\"correlation_id\":\"id\",\"original_url\":\"url\",\"short_url\":\"%s/\"},{\"correlation_id\":\"id\",\"original_url\":\"url\",\"short_url\":\"%s/\"},{\"correlation_id\":\"id\",\"original_url\":\"url\",\"short_url\":\"%s/\"},{\"correlation_id\":\"id\",\"original_url\":\"url\",\"short_url\":\"%s/\"}]", sAddr, sAddr, sAddr, sAddr, sAddr)},
			mock: mockParams{
				shortAddressTimes: 5,
				shortAddress:      serv.config.GetShortAddress()[7:],
				getUrlsError:      nil,
				getURLResult:      batches(5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//init mocks
			a := mocks.NewMockICutter(ctrl)
			c := mocks.NewMockConfiger(ctrl)
			c.EXPECT().GetShortAddress().Return(tt.mock.shortAddress).MaxTimes(tt.mock.shortAddressTimes)
			a.EXPECT().GetUserURLs(gomock.Any()).Return(tt.mock.getURLResult, tt.mock.getUrlsError).MaxTimes(1)
			s := New(a, c)
			//init request
			request, err := http.NewRequestWithContext(tt.r.ctx, http.MethodGet, url, strings.NewReader(""))
			require.NoError(t, err)

			w := httptest.NewRecorder()
			s.userUrlsHandler(w, request)
			res := w.Result()
			defer func() {
				require.NoError(t, res.Body.Close())
			}()
			assert.Equal(t, tt.expResp.code, res.StatusCode, "statusCode error")
			if res.StatusCode == http.StatusOK {
				assert.Equal(t, res.Header.Get("Content-Type"), "application/json")
			}

			b, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			resBody := string(b)
			err = res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.expResp.bodyMessage, string(resBody))
		})
	}
}

func BenchmarkCutterJSONHandler(b *testing.B) {
	b.StopTimer()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	a := mocks.NewMockICutter(ctrl)
	c := mocks.NewMockConfiger(ctrl)
	s := New(a, c)
	a.EXPECT().Cut(gomock.Any(), gomock.Any()).Return("returnString", nil).AnyTimes()
	_, testserver := initEnv()
	defer testserver.Close()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.cutterJSONHandler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, testserver.URL, strings.NewReader(positiveURL)))
	}
}

func BenchmarkCutterHandler(b *testing.B) {
	b.StopTimer()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	a := mocks.NewMockICutter(ctrl)
	c := mocks.NewMockConfiger(ctrl)
	s := New(a, c)
	_, testserver := initEnv()
	defer testserver.Close()
	a.EXPECT().Cut(gomock.Any(), gomock.Any()).Return("returnString", nil).AnyTimes()
	c.EXPECT().GetShortAddress().Return(testserver.URL).AnyTimes()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.cutterHandler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, testserver.URL, strings.NewReader(positiveURL)))
	}
}

func BenchmarkCutterJSONBatchHandler(b *testing.B) {
	b.StopTimer()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	a := mocks.NewMockICutter(ctrl)
	c := mocks.NewMockConfiger(ctrl)
	s := New(a, c)
	_, testserver := initEnv()
	defer testserver.Close()
	batch := make(jsonobject.Batch, 200)
	for i := 0; i < 200; i++ {
		str, err := randStringBytes(i)
		if err != nil {
			panic("randStringBytes out of control")
		}
		batch = append(batch, jsonobject.BatchItem{ID: str, OriginalURL: str})
	}

	a.EXPECT().UploadBatch(gomock.Any(), gomock.Any()).Return(batch, nil).AnyTimes()
	c.EXPECT().GetShortAddress().Return(testserver.URL).AnyTimes()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.cutterJSONBatchHandler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, testserver.URL, strings.NewReader(positiveURL)))
	}
}

func BenchmarkUserUrlsHandler(b *testing.B) {
	b.StopTimer()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	a := mocks.NewMockICutter(ctrl)
	c := mocks.NewMockConfiger(ctrl)
	s := New(a, c)
	_, testserver := initEnv()
	defer testserver.Close()
	batch := make(jsonobject.Batch, 200)
	for i := 0; i < 200; i++ {
		str, err := randStringBytes(i)
		if err != nil {
			panic("randStringBytes out of control")
		}
		batch = append(batch, jsonobject.BatchItem{ID: str, OriginalURL: str})
	}

	a.EXPECT().GetUserURLs(gomock.Any()).Return(batch, nil).AnyTimes()
	c.EXPECT().GetShortAddress().Return(testserver.URL).AnyTimes()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.userUrlsHandler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, testserver.URL, strings.NewReader(positiveURL)))
	}
}

func BenchmarkDeleteUserUrlsHandler(b *testing.B) {
	b.StopTimer()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	a := mocks.NewMockICutter(ctrl)
	c := mocks.NewMockConfiger(ctrl)
	s := New(a, c)
	_, testserver := initEnv()
	defer testserver.Close()
	a.EXPECT().DeleteUrls(gomock.Any(), gomock.Any()).AnyTimes()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.deleteUserUrlsHandler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, testserver.URL, strings.NewReader(positiveURL)))
	}
}

func randStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("randStringBytes: Generating random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
