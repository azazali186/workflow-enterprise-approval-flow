package audit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// DiffResult represents the result of comparing two states
type DiffResult struct {
	Before      map[string]interface{} `json:"before"`
	After       map[string]interface{} `json:"after"`
	Changed     []FieldChange          `json:"changed"`
	HasChanges  bool                   `json:"has_changes"`
}

// FieldChange represents a single field change
type FieldChange struct {
	Field    string      `json:"field"`
	Before   interface{} `json:"before"`
	After    interface{} `json:"after"`
	Type     string      `json:"type"` // "added", "removed", "modified"
}

// ComputeChanges computes the diff between two entities and returns structured change data
// before and after should be pointers to structs or maps
func ComputeChanges(before, after interface{}) *DiffResult {
	beforeMap := toMap(before)
	afterMap := toMap(after)

	result := &DiffResult{
		Before:  beforeMap,
		After:   afterMap,
		Changed: make([]FieldChange, 0),
	}

	// Fields to skip in diff
	skipFields := map[string]bool{
		"id":         true,
		"created_at": true,
		"updated_at": true,
		"deleted_at": true,
		"password":   true, // Never log password changes
	}

	allKeys := make(map[string]bool)
	for k := range beforeMap {
		allKeys[k] = true
	}
	for k := range afterMap {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if skipFields[key] {
			continue
		}

		beforeVal, beforeExists := beforeMap[key]
		afterVal, afterExists := afterMap[key]

		if !beforeExists && afterExists {
			result.Changed = append(result.Changed, FieldChange{
				Field:  key,
				Before: nil,
				After:  afterVal,
				Type:   "added",
			})
		} else if beforeExists && !afterExists {
			result.Changed = append(result.Changed, FieldChange{
				Field:  key,
				Before: beforeVal,
				After:  nil,
				Type:   "removed",
			})
		} else if !valuesEqual(beforeVal, afterVal) {
			result.Changed = append(result.Changed, FieldChange{
				Field:  key,
				Before: beforeVal,
				After:  afterVal,
				Type:   "modified",
			})
		}
	}

	result.HasChanges = len(result.Changed) > 0
	return result
}

// ComputeChangesFromMaps computes diff from two maps directly
func ComputeChangesFromMaps(before, after map[string]interface{}) *DiffResult {
	result := &DiffResult{
		Before:  before,
		After:   after,
		Changed: make([]FieldChange, 0),
	}

	skipFields := map[string]bool{
		"id":         true,
		"created_at": true,
		"updated_at": true,
		"deleted_at": true,
		"password":   true,
	}

	allKeys := make(map[string]bool)
	for k := range before {
		allKeys[k] = true
	}
	for k := range after {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if skipFields[key] {
			continue
		}

		beforeVal, beforeExists := before[key]
		afterVal, afterExists := after[key]

		if !beforeExists && afterExists {
			result.Changed = append(result.Changed, FieldChange{
				Field:  key,
				Before: nil,
				After:  afterVal,
				Type:   "added",
			})
		} else if beforeExists && !afterExists {
			result.Changed = append(result.Changed, FieldChange{
				Field:  key,
				Before: beforeVal,
				After:  nil,
				Type:   "removed",
			})
		} else if !valuesEqual(beforeVal, afterVal) {
			result.Changed = append(result.Changed, FieldChange{
				Field:  key,
				Before: beforeVal,
				After:  afterVal,
				Type:   "modified",
			})
		}
	}

	result.HasChanges = len(result.Changed) > 0
	return result
}

// toMap converts a struct or map to a map[string]interface{}
func toMap(v interface{}) map[string]interface{} {
	if v == nil {
		return map[string]interface{}{}
	}

	// If already a map
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}

	// If a pointer to a map
	if m, ok := v.(*map[string]interface{}); ok && m != nil {
		return *m
	}

	// Convert struct to map using JSON
	data, err := json.Marshal(v)
	if err != nil {
		return map[string]interface{}{}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return map[string]interface{}{}
	}

	return result
}

// valuesEqual compares two interface{} values for equality
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Normalize: both to JSON and compare
	aData, err1 := json.Marshal(a)
	bData, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	}

	return string(aData) == string(bData)
}

// ChangesToJSONMap converts a DiffResult to a JSONMap suitable for audit log storage
func ChangesToJSONMap(diff *DiffResult) map[string]interface{} {
	if diff == nil || !diff.HasChanges {
		return nil
	}

	changes := make(map[string]interface{})
	for _, change := range diff.Changed {
		changes[change.Field] = map[string]interface{}{
			"before":  change.Before,
			"after":   change.After,
			"type":    change.Type,
		}
	}

	return changes
}

// SummaryOfChanges generates a human-readable summary of changes
func SummaryOfChanges(diff *DiffResult) string {
	if diff == nil || !diff.HasChanges {
		return ""
	}

	var parts []string
	for _, change := range diff.Changed {
		switch change.Type {
		case "added":
			parts = append(parts, fmt.Sprintf("Added %s", change.Field))
		case "removed":
			parts = append(parts, fmt.Sprintf("Removed %s", change.Field))
		case "modified":
			parts = append(parts, fmt.Sprintf("Changed %s from %v to %v", change.Field, change.Before, change.After))
		}
	}

	return strings.Join(parts, "; ")
}

// FilterSensitiveFields removes sensitive fields from a map
func FilterSensitiveFields(m map[string]interface{}) map[string]interface{} {
	sensitiveFields := map[string]bool{
		"password":          true,
		"old_password":      true,
		"new_password":      true,
		"access_token":      true,
		"refresh_token":     true,
		"secret":            true,
		"api_key":           true,
		"private_key":       true,
	}

	result := make(map[string]interface{})
	for k, v := range m {
		if !sensitiveFields[k] {
			result[k] = v
		} else {
			result[k] = "[REDACTED]"
		}
	}
	return result
}

// StructToMapWithJSONTags converts a struct to map using json tags (for accurate field names)
func StructToMapWithJSONTags(v interface{}) map[string]interface{} {
	if v == nil {
		return map[string]interface{}{}
	}

	data, err := json.Marshal(v)
	if err != nil {
		return map[string]interface{}{}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return map[string]interface{}{}
	}

	return result
}

// IgnoreUnexported filters out zero-value fields from a struct map
// Useful for partial updates where only non-zero fields should be shown
func IgnoreUnexported(v interface{}) map[string]interface{} {
	result := StructToMapWithJSONTags(v)

	// Remove zero values
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldVal := val.Field(i)

			// Get json tag name
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			jsonName := strings.Split(jsonTag, ",")[0]

			// Remove if zero value
			if fieldVal.IsZero() {
				delete(result, jsonName)
			}
		}
	}

	return result
}
