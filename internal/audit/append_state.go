package audit

import (
	"crypto/sha256"
	"encoding/hex"
)

func (c *Chain) beginAppend(event Event) {
	marker := sha256.Sum256([]byte(c.previous + event.RequestID + event.ObjectID))
	c.previous = hex.EncodeToString(marker[:])
}
