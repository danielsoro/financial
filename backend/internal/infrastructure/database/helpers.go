package database

import "strings"

func isDuplicateKey(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}
