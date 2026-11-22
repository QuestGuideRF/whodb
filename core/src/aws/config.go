package aws
import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func LoadAWSConfig(ctx context.Context, creds *engine.Credentials) (aws.Config, error) {
	awsCreds, err := ParseFromWhoDB(creds)
	if err != nil {
		return aws.Config{}, err
	}
	return loadConfigFromAWSCredentials(ctx, awsCreds)
}
func loadConfigFromAWSCredentials(ctx context.Context, awsCreds *AWSCredentialConfig) (aws.Config, error) {
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(awsCreds.Region),
	}
	if awsCreds.IsProfileAuth() && awsCreds.ProfileName != "" {
		options = append(options, awsconfig.WithSharedConfigProfile(awsCreds.ProfileName))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		log.WithFields(map[string]any{
			"region":     awsCreds.Region,
			"authMethod": awsCreds.AuthMethod,
		}).WithError(err).Error("Failed to load AWS configuration")
		return aws.Config{}, HandleAWSError(err)
	}
	log.WithFields(map[string]any{
		"region":         awsCreds.Region,
		"authMethod":     awsCreds.AuthMethod,
		"hasProfileName": awsCreds.ProfileName != "",
	}).Debug("AWS configuration loaded successfully")
	return cfg, nil
}