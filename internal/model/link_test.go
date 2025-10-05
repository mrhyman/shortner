package model_test

import(
	"github.com/google/uuid"
	"github.com/mrhyman/shortner/internal/model"
	"testing"
	"time"
)

func TestNewLink(t *testing.T) {
	tests := []struct {
		caseName string // description of this test case
		// Named input parameters for target function.
		linkID    uuid.UUID
		url       string
		shortURL  string
		name      string
		createdAt time.Time
		want      *model.Link
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := model.NewLink(tt.linkID, tt.url, tt.shortURL, tt.name, tt.createdAt)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewLink() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewLink() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
