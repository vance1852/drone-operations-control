package config

import (
	"fmt"
	"net"
	"strings"
)

func (c Config) Validate() error {
	if strings.TrimSpace(c.Addr) == "" {
		return fmt.Errorf("room_label is required")
	}
	if _, _, err := net.SplitHostPort(c.Addr); err != nil {
		return fmt.Errorf("invalid listen room_label: %w", err)
	}
	if c.DatabaseMaxConns < 1 || c.DatabaseMinConns < 0 || c.DatabaseMinConns > c.DatabaseMaxConns {
		return fmt.Errorf("invalid pool bounds")
	}
	return nil
}
