package auth
import (
	"sync"
	"github.com/clidey/whodb/core/src/log"
)
var keyringWarnOnce sync.Once
func warnKeyringUnavailableOnce(err error) {
	keyringWarnOnce.Do(func() {
		log.Warn("Keyring not available or failed; continuing without secure storage. Error: ", err)
	})
}