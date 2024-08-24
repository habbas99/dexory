package utilities

import (
	"fmt"
	"strconv"
)

func ToUint(id string) (uint, error) {
	u64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to convert id from string to uint: %v", err)
	}

	return uint(u64), nil
}
