package aws
import (
	"context"
	"fmt"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/clidey/whodb/core/src/log"
)
func GenerateRDSAuthToken(ctx context.Context, cfg awssdk.Config, endpoint string, port int, region, username string) (string, error) {
	addr := fmt.Sprintf("%s:%d", endpoint, port)
	log.Infof("RDS IAM Auth: generating token for %s (region=%s, user=%s)", addr, region, username)
	token, err := auth.BuildAuthToken(ctx, addr, region, username, cfg.Credentials)
	if err != nil {
		log.Errorf("RDS IAM Auth: token generation failed for %s: %v", addr, err)
		return "", HandleAWSError(err)
	}
	log.Infof("RDS IAM Auth: token generated successfully for %s (token length=%d)", addr, len(token))
	return token, nil
}