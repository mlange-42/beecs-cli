package util

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseIndices(str string) ([]int, error) {
	if len(str) == 0 {
		return nil, nil
	}

	if strings.Contains(str, "-") {
		parts := strings.Split(str, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid syntax for indices in '%s'", str)
		}
		lower, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("error parsing numbers for indices in '%s'", str)
		}
		upper, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("error parsing numbers for indices in '%s'", str)
		}
		indices := make([]int, upper-lower+1)
		for i := range indices {
			indices[i] = lower + i
		}
		return indices, nil
	}
	parts := strings.Split(str, ",")
	indices := make([]int, len(parts))
	var err error
	for i, p := range parts {
		indices[i], err = strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("error parsing numbers for indices in '%s'", str)
		}
	}
	return indices, nil
}
