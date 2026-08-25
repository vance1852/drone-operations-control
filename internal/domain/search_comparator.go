package domain

import "strings"

func searchComparator(field SortField, items []DroneTask) func(int, int) bool {
	switch field {
	case SortExpiry:
		return func(left, right int) bool {
			return items[left].ExpiresAt.Before(items[right].ExpiresAt)
		}
	case SortCode:
		return func(left, right int) bool {
			return strings.ToLower(items[left].TaskCode) < strings.ToLower(items[right].TaskCode)
		}
	case SortCreated:
		return func(left, right int) bool {
			return items[left].ID < items[right].ID
		}
	default:
		return func(left, right int) bool {
			return items[left].ID < items[right].ID
		}
	}
}
