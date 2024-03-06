package cutter

import (
	"errors"
	"testing"
	"time"

	"github.com/dmad1989/urlcut/internal/mocks"
	"github.com/golang/mock/gomock"
	"go.uber.org/goleak"
)

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
