package cutter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/dmad1989/urlcut/internal/mocks"
)

func TestCut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pgerr := &pgconn.PgError{
		Code: pgerrcode.UniqueViolation,
	}
	type mockParams struct {
		addErrReturn error
		getURLReturn string
		getErrReturn error
		getTimes     int
	}
	type expected struct {
		isEmptyRes bool
		isNoError  bool
	}
	m := mocks.NewMockStore(ctrl)
	tests := []struct {
		name       string
		mockParams mockParams
		expected   expected
	}{{
		name: "positive",
		mockParams: mockParams{
			addErrReturn: nil,
			getURLReturn: "sss",
			getErrReturn: nil,
			getTimes:     0,
		},
		expected: expected{
			isEmptyRes: false,
			isNoError:  true,
		},
	},
		{
			name: "negative - add custom error",
			mockParams: mockParams{
				addErrReturn: errors.New("from db"),
				getURLReturn: "",
				getErrReturn: nil,
				getTimes:     0,
			},
			expected: expected{
				isEmptyRes: true,
				isNoError:  false,
			},
		},
		{
			name: "negative - add UniqueURLError ",
			mockParams: mockParams{
				addErrReturn: NewUniqueURLError("code", errors.New("not unique URL")),
				getURLReturn: "",
				getErrReturn: nil,
				getTimes:     0,
			},
			expected: expected{
				isEmptyRes: true,
				isNoError:  false,
			},
		},
		{
			name: "negative - add UniqueViolation get no error",
			mockParams: mockParams{
				addErrReturn: pgerr,
				getURLReturn: "cuten",
				getErrReturn: nil,
				getTimes:     1,
			},
			expected: expected{
				isEmptyRes: false,
				isNoError:  false,
			},
		},
		{
			name: "negative - add UniqueViolation get  error",
			mockParams: mockParams{
				addErrReturn: pgerr,
				getURLReturn: "cuten",
				getErrReturn: errors.New("from db"),
				getTimes:     1,
			},
			expected: expected{
				isEmptyRes: true,
				isNoError:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.mockParams.addErrReturn).MaxTimes(1)
			m.EXPECT().GetShortURL(gomock.Any(), gomock.Any()).Return(tt.mockParams.getURLReturn, tt.mockParams.getErrReturn).MaxTimes(tt.mockParams.getTimes)
			app := New(m)
			res, err := app.Cut(context.TODO(), "someurl")
			assert.Equal(t, tt.expected.isEmptyRes, res == "")
			if tt.expected.isNoError {
				assert.Empty(t, err)
				return
			}
			assert.NotEmpty(t, err)
		})
	}
}

func TestCheckUrls(t *testing.T) {
	defer goleak.VerifyNone(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockStore(ctrl)

	tests := []struct {
		name     string
		inputSl  []string
		maxTimes int
		dbError  error
	}{{
		name:     "10els",
		inputSl:  make([]string, 10),
		maxTimes: 1,
	},
		{
			name:     "100els",
			inputSl:  make([]string, 100),
			maxTimes: 1,
		},
		{
			name:     "1000els",
			inputSl:  make([]string, 1000),
			maxTimes: 10,
		},
		{
			name:     "1001els",
			inputSl:  make([]string, 1001),
			maxTimes: 11,
		},
		{
			name:     "999els",
			inputSl:  make([]string, 999),
			maxTimes: 10,
		},
		{
			name:     "165els",
			inputSl:  make([]string, 165),
			maxTimes: 2,
		},
		{
			name:     "100els_error",
			inputSl:  make([]string, 100),
			maxTimes: 1,
			dbError:  errors.New("custom db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.EXPECT().DeleteURLs(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.dbError).MaxTimes(tt.maxTimes)
			app := New(m)
			go app.DeleteUrls("", tt.inputSl)
		})

	}
	time.Sleep(time.Second * 30)
	defer goleak.VerifyNone(t)
}

func BenchmarkGetKeyByValue(b *testing.B) {
	b.StopTimer()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	m := mocks.NewMockStore(ctrl)
	a := New(m)
	m.EXPECT().GetOriginalURL(gomock.Any(), "someString").Return("returnString", nil).AnyTimes()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		a.GetKeyByValue(context.TODO(), "someString")
	}
}

func BenchmarkUploadBatch(b *testing.B) {
	b.StopTimer()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	m := mocks.NewMockStore(ctrl)
	a := New(m)
	batch := make(jsonobject.Batch, 200)
	for i := 0; i < 200; i++ {
		str, err := randStringBytes(i)
		if err != nil {
			panic("randStringBytes out of control")
		}
		batch = append(batch, jsonobject.BatchItem{ID: str, OriginalURL: str})
	}
	m.EXPECT().UploadBatch(gomock.Any(), gomock.Any()).Return(batch, nil).AnyTimes()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		a.UploadBatch(context.TODO(), batch)
	}
}

func BenchmarkDeleteUrls(b *testing.B) {
	b.StopTimer()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	m := mocks.NewMockStore(ctrl)
	a := New(m)
	m.EXPECT().DeleteURLs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	ids := make(jsonobject.ShortIds, 200)
	for i := 0; i < 200; i++ {
		str, err := randStringBytes(i)
		if err != nil {
			panic("randStringBytes out of control")
		}
		ids = append(ids, str)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		a.DeleteUrls("customID", ids)
	}
}

func BenchmarkCut(b *testing.B) {
	b.StopTimer()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	m := mocks.NewMockStore(ctrl)
	a := New(m)
	m.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().GetShortURL(gomock.Any(), gomock.Any()).Return("url", nil).AnyTimes()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		a.Cut(context.TODO(), "someurl")
	}
}
