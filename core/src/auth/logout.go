package auth
import (
	"context"
	"net/http"
	"github.com/clidey/whodb/core/graph/model"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/source"
	"github.com/clidey/whodb/core/src/sourcecatalog"
)
func Logout(ctx context.Context) (*model.StatusResponse, error) {
	if creds := GetSourceCredentials(ctx); creds != nil {
		if spec, ok := sourcecatalog.Find(creds.SourceType); ok {
			_ = source.Invalidate(ctx, spec, creds)
		}
		if creds.ID != nil {
			if err := DeleteCredentials(*creds.ID); err != nil {
				warnKeyringUnavailableOnce(err)
			}
		}
	}
	http.SetCookie(ctx.Value(common.RouterKey_ResponseWriter).(http.ResponseWriter), nil)
	return &model.StatusResponse{
		Status: true,
	}, nil
}