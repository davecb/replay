package replay

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io/fs"
	"log"
	"testing"
)

func TestOpenReplay(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr error
	}{
		{
			name:    "open a file",
			file:    "onechunk.log",
			wantErr: nil,
		},
		{
			name:    "fail to open a missing file",
			file:    "absent.file",
			wantErr: &fs.PathError{},
		},
	}
	for _, tt := range tests {
		logger := sugar()
		t.Run(tt.name, func(t *testing.T) {
			testfile := "../../cmd/example/testdata/" + tt.file

			err := Open(testfile, logger)
			t.Logf("wantErr = %v\n", err)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, nil, err)
			}
			Close()
		})
	}
}

func TestGetReplay(t *testing.T) {
	tests := []struct {
		name       string
		file       string
		want       []ConfusionMatrixInfoLog
		wantErr    error
		validChunk bool
	}{
		// GREEN: can read a chunk
		{
			name:    "read a chunk",
			file:    "onechunk.log",
			want:    []ConfusionMatrixInfoLog{},
			wantErr: nil,
		},
		// can't stop at end with an indication of EOF
		{
			name:       "read an EOF", // requires a specific
			wantErr:    nil,           //errUnimplemented,                             // really want nil
			validChunk: false,         // not implemented
		},
	}

	var got []ConfusionMatrixInfoLog
	logger := sugar()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testfile := "../../cmd/example/testdata/" + tt.file

			err := Open(testfile, logger)
			if err != nil {
				t.Fatalf("could not open %q to start a test\n", testfile)
			}

			got, err = Get("capped")
			t.Logf("want = %v, wantErr = %v\n", tt.want, tt.wantErr)
			t.Logf("got  = %v, err     = %v\n", got, err)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			} else {
				assert.Equal(t, tt.want, got)
			}

			Close()
		})
	}
}

func Example() {
	logger := sugar()

	testfile := "../../../cmd/example/testdata/onechunk.log"
	err := Open(testfile, logger)
	if err != nil {
		log.Fatalf("could not open %q to start a test\n", testfile)
	}
	got, err := Get("capped")
	fmt.Printf("replay.GetCapped returned %v, err = %v\n", got, err)
	Close()
	// Output:
	// replay.GetCapped returned map[:{{0 0} 12 4 4}], err = <nil>
}

// sugar just creates a sugared logger for testing
func sugar() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}
