package repo

import "fmt"

// asString and asBool safely coerce a Neo4j record value, returning the
// zero value instead of panicking if it's nil or an unexpected type (a
// bare `.(string)`/`.(bool)` type assertion crashes the request on any
// record with a missing/null property).
func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func asBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// asInt coerces a Neo4j integer property, which the driver surfaces as
// int64, down to a plain int.
func asInt(v any) int {
	if i, ok := v.(int64); ok {
		return int(i)
	}
	return 0
}

// asOptionalString formats a possibly-nil Neo4j value (e.g. a datetime
// property that hasn't been set yet, like Notification.deliveredAt before
// delivery succeeds) as a string, returning "" instead of the literal
// "<nil>" that fmt.Sprintf("%v", nil) would otherwise produce.
func asOptionalString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
