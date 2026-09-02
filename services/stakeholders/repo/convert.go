package repo

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
