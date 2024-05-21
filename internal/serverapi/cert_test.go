package serverapi

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCert(t *testing.T) {
	deleteFile := func(path string) {
		err := os.Remove(path)
		require.NoError(t, err, "delete file")
	}

	checkExists := func(path string) {
		fi, err := os.Stat(path)
		require.NoErrorf(t, err, "%s file not found", path)
		defer deleteFile(path)
		require.True(t, fi.Size() > 0)
	}

	type params struct {
		certPath string
		keyPath  string
	}

	tests := []struct {
		name          string
		p             params
		expectedError string
	}{{
		name:          "no cert path",
		p:             params{certPath: "", keyPath: ""},
		expectedError: "CreateCert: CERTIFICATE: saveToFile: path Create: open : The system cannot find the file specified.",
	},
		{
			name:          "no key path",
			p:             params{certPath: "cert.pem", keyPath: ""},
			expectedError: "CreateCert: RSA PRIVATE KEY: saveToFile: path Create: open : The system cannot find the file specified.",
		},
		{
			name:          "positive",
			p:             params{certPath: "cert.pem", keyPath: "key.pem"},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateCert(tt.p.certPath, tt.p.keyPath)
			if err != nil {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				require.Empty(t, tt.expectedError, "no error catched? but expected")
				checkExists(tt.p.certPath)
				checkExists(tt.p.keyPath)
			}
		})
	}

}
