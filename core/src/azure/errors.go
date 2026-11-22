package azure
import (
	"errors"
	"fmt"
	"strings"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)
var (
	ErrAccessDenied         = errors.New("access denied: check Azure RBAC permissions for this operation")
	ErrInvalidCredentials   = errors.New("invalid Azure credentials: check tenant ID, client ID, and secret")
	ErrExpiredCredentials   = errors.New("Azure credentials have expired: re-authenticate or refresh token")
	ErrSubscriptionNotFound = errors.New("subscription not found: check the subscription ID and permissions")
	ErrResourceNotFound     = errors.New("resource not found: check the resource name and subscription")
	ErrServiceUnavailable   = errors.New("Azure service temporarily unavailable: try again later")
	ErrThrottling           = errors.New("request throttled: too many requests, try again later")
	ErrConnectionFailed     = errors.New("connection failed: check network connectivity and endpoint")
	ErrSubscriptionRequired = errors.New("Azure subscription ID is required")
	ErrTenantIDRequired     = errors.New("service principal auth requires a Tenant ID")
	ErrClientIDRequired     = errors.New("service principal auth requires a Client ID")
	ErrClientSecretRequired = errors.New("service principal auth requires a Client Secret")
	ErrInvalidAuthMethod    = errors.New("invalid auth method: must be one of: default, service-principal")
	ErrAzureProviderDisabled = errors.New("Azure provider is disabled")
)
func HandleAzureError(err error) error {
	if err == nil {
		return nil
	}
	var respErr *azcore.ResponseError
	if errors.As(err, &respErr) {
		return handleResponseError(respErr)
	}
	errMsg := err.Error()
	if strings.Contains(errMsg, "no credential") ||
		strings.Contains(errMsg, "authentication failed") {
		return ErrInvalidCredentials
	}
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "dial tcp") {
		return fmt.Errorf("%w: %s", ErrConnectionFailed, errMsg)
	}
	return err
}
func handleResponseError(respErr *azcore.ResponseError) error {
	code := respErr.ErrorCode
	status := respErr.StatusCode
	switch code {
	case "AuthorizationFailed", "AuthenticationFailed", "AuthenticationFailedInvalidHeader":
		return ErrAccessDenied
	case "InvalidAuthenticationToken", "InvalidAuthenticationTokenTenant":
		return ErrInvalidCredentials
	case "ExpiredAuthenticationToken":
		return ErrExpiredCredentials
	case "SubscriptionNotFound":
		return ErrSubscriptionNotFound
	case "ResourceNotFound", "ResourceGroupNotFound":
		return ErrResourceNotFound
	case "TooManyRequests", "RequestRateLimited":
		return ErrThrottling
	}
	switch {
	case status == 401:
		return ErrInvalidCredentials
	case status == 403:
		return ErrAccessDenied
	case status == 404:
		return ErrResourceNotFound
	case status == 429:
		return ErrThrottling
	case status >= 500:
		return ErrServiceUnavailable
	}
	return fmt.Errorf("Azure error [%s] (HTTP %d): %s", code, status, respErr.Error())
}