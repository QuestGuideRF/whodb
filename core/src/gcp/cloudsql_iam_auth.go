package gcp
import (
	"context"
	"fmt"
	"os"
	"github.com/clidey/whodb/core/src/log"
	"golang.org/x/oauth2/google"
)
func GenerateCloudSQLIAMAuthToken(ctx context.Context, serviceAccountKeyPath, username string) (string, error) {
	log.Infof("Cloud SQL IAM Auth: generating token for user=%s", username)
	scopes := []string{"https://www.googleapis.com/auth/sqlservice.login"}
	var creds *google.Credentials
	var err error
	if serviceAccountKeyPath != "" {
		data, readErr := os.ReadFile(serviceAccountKeyPath)
		if readErr != nil {
			return "", fmt.Errorf("failed to read service account key file: %w", readErr)
		}
		creds, err = google.CredentialsFromJSONWithParams(ctx, data, google.CredentialsParams{Scopes: scopes})
	} else {
		creds, err = google.FindDefaultCredentials(ctx, scopes...)
	}
	if err != nil {
		log.Errorf("Cloud SQL IAM Auth: failed to find credentials: %v", err)
		return "", fmt.Errorf("failed to find GCP credentials for Cloud SQL IAM auth: %w", err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		log.Errorf("Cloud SQL IAM Auth: failed to obtain token: %v", err)
		return "", HandleGCPError(err)
	}
	log.Infof("Cloud SQL IAM Auth: token generated successfully for user=%s (token length=%d)", username, len(token.AccessToken))
	return token.AccessToken, nil
}