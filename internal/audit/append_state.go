package audit

import (
	"crypto/sha256"
	"encoding/hex"
)

// beginAppend derives a provisional link marker that commits the current
// chain head to the identity of the event being appended. It must not mutate
// chain state: a rejected append (for example, an event that fails to
// JSON-encode after passing validation) leaves the chain head untouched so
// that subsequent appends continue from the real prior node.
func (c *Chain) beginAppend(event Event) string {
	marker := sha256.Sum256([]byte(c.previous + event.RequestID + event.ObjectID))
	return hex.EncodeToString(marker[:])
}
