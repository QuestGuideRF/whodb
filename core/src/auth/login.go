package auth
import (
	"context"
	"github.com/clidey/whodb/core/graph/model"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/source"
)
func LoginSource(_ context.Context, credentials *source.Credentials) (*model.StatusResponse, error) {
	values := credentials.CloneValues()
	log.Debugf("[LoginSource] sourceType=%s, values=%d", credentials.SourceType, len(values))
	if credentials.ID != nil && *credentials.ID != "" {
		storedCredentials := &source.Credentials{
			ID:          credentials.ID,
			SourceType:  credentials.SourceType,
			Values:      values,
			AccessToken: credentials.AccessToken,
			IsProfile:   false,
		}
		if err := SaveCredentials(*credentials.ID, storedCredentials); err != nil {
			warnKeyringUnavailableOnce(err)
		}
	}
	return &model.StatusResponse{Status: true}, nil
}