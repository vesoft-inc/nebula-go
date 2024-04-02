package nebula_go

import (
	"regexp"
	"strconv"
)

func IndexOf(collection []string, element string) int {
	for i, item := range collection {
		if item == element {
			return i
		}
	}

	return -1
}

func parseTTL(s string) (string, uint, error) {
	col, err := parseTTLCol(s)
	if err != nil {
		return "", 0, err
	}

	duration, err := parseTTLDuration(s)
	if err != nil {
		return "", 0, err
	}

	return col, duration, nil
}

func parseTTLCol(s string) (string, error) {
	reg, err := regexp.Compile(`ttl_col = "(\w+)"`)
	ss := reg.FindStringSubmatch(s)

	if err != nil {
		return "", err
	}

	if len(ss) == 2 {
		return ss[1], nil
	}

	return "", nil
}

func parseTTLDuration(s string) (uint, error) {
	reg, err := regexp.Compile(`ttl_duration = (\d+)`)

	if err != nil {
		return 0, err
	}

	ss := reg.FindStringSubmatch(s)

	if len(ss) == 2 {
		ttl, err := strconv.Atoi(ss[1])
		if err != nil {
			return 0, err
		}
		return uint(ttl), nil
	}

	return 0, nil
}
