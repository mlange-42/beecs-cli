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

	indices := []int{}
	parts := strings.Split(str, ",")
	for _, p := range parts {
		if len(p) == 0 {
			return nil, fmt.Errorf("invalid syntax for indices in '%s'", str)
		}
		subParts := strings.Split(p, "-")
		if len(subParts) > 2 || len(subParts) == 0 {
			return nil, fmt.Errorf("invalid syntax for indices in '%s'", str)
		}
		if len(subParts) == 1 {
			value, err := strconv.Atoi(subParts[0])
			if err != nil {
				return nil, fmt.Errorf("error parsing numbers for indices in '%s'", str)
			}
			indices = append(indices, value)
			continue
		}
		lower, err := strconv.Atoi(subParts[0])
		if err != nil {
			return nil, fmt.Errorf("error parsing numbers for indices in '%s'", str)
		}
		upper, err := strconv.Atoi(subParts[1])
		if err != nil {
			return nil, fmt.Errorf("error parsing numbers for indices in '%s'", str)
		}
		values := upper - lower + 1
		for i := 0; i < values; i++ {
			indices = append(indices, lower+i)
		}
	}
	return indices, nil
}
