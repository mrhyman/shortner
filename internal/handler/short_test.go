package handler_test

import(
	"github.com/mrhyman/shortner/internal/handler"
	"net/http"
	"testing"
)

func TestHandler_ShortLinkHandler(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		res http.ResponseWriter
		req *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			var h handler.Handler
			h.ShortLinkHandler(tt.res, tt.req)
		})
	}
}
