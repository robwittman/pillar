package service

import (
	"strings"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/tidwall/gjson"
)

// EvaluateFilter checks if a JSON payload matches all filter conditions.
// An empty conditions list matches everything.
func EvaluateFilter(filter domain.TriggerFilter, payload []byte) bool {
	for _, cond := range filter.Conditions {
		result := gjson.GetBytes(payload, cond.Path)

		switch cond.Op {
		case "eq":
			if result.String() != cond.Value {
				return false
			}
		case "neq":
			if result.String() == cond.Value {
				return false
			}
		case "contains":
			if !strings.Contains(result.String(), cond.Value) {
				return false
			}
		case "exists":
			if !result.Exists() {
				return false
			}
		default:
			return false
		}
	}
	return true
}
