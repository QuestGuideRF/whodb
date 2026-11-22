package mongodb
import (
	"errors"
	"fmt"
	"strings"
	"go.mongodb.org/mongo-driver/v2/mongo"
)
func handleMongoError(err error) error {
	if err == nil {
		return nil
	}
	var we mongo.WriteException
	if errors.As(err, &we) {
		for _, werr := range we.WriteErrors {
			switch werr.Code {
			case 11000:
				return fmt.Errorf("duplicate key: a document with the same identifier already exists")
			default:
				return fmt.Errorf("write error (%d): %s", werr.Code, werr.Message)
			}
		}
	}
	var ce mongo.CommandError
	if errors.As(err, &ce) {
		switch ce.Code {
		case 48:
			return fmt.Errorf("collection already exists")
		case 121:
			return fmt.Errorf("document validation failed: %s", ce.Message)
		}
		return fmt.Errorf("command error (%d): %s", ce.Code, ce.Message)
	}
	msg := err.Error()
	msg = strings.ReplaceAll(msg, "\n", " ")
	return fmt.Errorf("mongodb operation failed: %s", msg)
}