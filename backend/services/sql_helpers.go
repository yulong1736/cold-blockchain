package services

import "strings"

func inClauseStrings(values []string) (placeholder string, args []interface{}) {
	if len(values) == 0 {
		return "", nil
	}
	args = make([]interface{}, len(values))
	for i := range values {
		args[i] = values[i]
	}
	placeholder = strings.Repeat("?,", len(values)-1) + "?"
	return placeholder, args
}

func inClauseUints(values []uint) (placeholder string, args []interface{}) {
	if len(values) == 0 {
		return "", nil
	}
	args = make([]interface{}, len(values))
	for i := range values {
		args[i] = values[i]
	}
	placeholder = strings.Repeat("?,", len(values)-1) + "?"
	return placeholder, args
}
