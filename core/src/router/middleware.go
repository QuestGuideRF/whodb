package router
import (
	"context"
	"net/http"
	"github.com/clidey/whodb/core/src/analytics"
	"github.com/clidey/whodb/core/src/common"
	"github.com/go-chi/chi/v5/middleware"
)
func contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), common.RouterKey_ResponseWriter, w)
		metadata := analytics.BuildMetadata(r)
		if metadata.RequestID == "" {
			if requestID := middleware.GetReqID(ctx); requestID != "" {
				metadata.RequestID = requestID
			}
		}
		ctx = analytics.WithMetadata(ctx, metadata)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}