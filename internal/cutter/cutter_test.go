package cutter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dmad1989/urlcut/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestCut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pgerr := &pgconn.PgError{
		Code: pgerrcode.UniqueViolation,
	}
	type mockParams struct {
		addErrReturn error
		getUrlReturn string
		getErrReturn error
		getTimes     int
	}
	type expected struct {
		isEmptyRes bool
		isNoError  bool
		err        error
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
			getUrlReturn: "sss",
			getErrReturn: nil,
			getTimes:     0,
		},
		expected: expected{
			isEmptyRes: false,
			isNoError:  true,
			err:        nil,
		},
	},
		{
			name: "negative - add custom error",
			mockParams: mockParams{
				addErrReturn: errors.New("some db error"),
				getUrlReturn: "",
				getErrReturn: nil,
				getTimes:     0,
			},
			expected: expected{
				isEmptyRes: true,
				isNoError:  false,
				err:        errors.New("some db error"),
			},
		},
		{
			name: "negative - add UniqueURLError ",
			mockParams: mockParams{
				addErrReturn: NewUniqueURLError("code", errors.New("unique error text")),
				getUrlReturn: "",
				getErrReturn: nil,
				getTimes:     0,
			},
			expected: expected{
				isEmptyRes: true,
				isNoError:  false,
				err:        NewUniqueURLError("code", errors.New("unique error text")),
			},
		},
		{
			name: "negative - add UniqueViolation get no error",
			mockParams: mockParams{
				addErrReturn: pgerr,
				getUrlReturn: "cuten",
				getErrReturn: nil,
				getTimes:     1,
			},
			expected: expected{
				isEmptyRes: false,
				isNoError:  false,
				err:        NewUniqueURLError("cuten", pgerr),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.mockParams.addErrReturn).MaxTimes(1)
			m.EXPECT().GetShortURL(gomock.Any(), gomock.Any()).Return(tt.mockParams.getUrlReturn, tt.mockParams.getErrReturn).MaxTimes(tt.mockParams.getTimes)
			app := New(m)
			res, err := app.Cut(context.TODO(), "someurl")
			assert.Equal(t, tt.expected.isEmptyRes, res == "")
			if tt.expected.isNoError {
				assert.Empty(t, err)
				return
			}
			assert.NotEmpty(t, err)
			assert.ErrorContains(t, err, tt.expected.err.Error())
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
