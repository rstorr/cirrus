package dynamo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// attributeValueToString converts a DynamoDB AttributeValue to a human-readable string
func attributeValueToString(av types.AttributeValue) string {
	if av == nil {
		return "<null>"
	}

	switch v := av.(type) {
	case *types.AttributeValueMemberS:
		// String
		return v.Value

	case *types.AttributeValueMemberN:
		// Number
		return v.Value

	case *types.AttributeValueMemberBOOL:
		// Boolean
		if v.Value {
			return "true"
		}
		return "false"

	case *types.AttributeValueMemberNULL:
		// Null
		return "<null>"

	case *types.AttributeValueMemberL:
		// List
		var items []string
		for _, item := range v.Value {
			items = append(items, attributeValueToString(item))
		}
		return "[" + strings.Join(items, ", ") + "]"

	case *types.AttributeValueMemberM:
		// Map
		var pairs []string
		for key, val := range v.Value {
			pairs = append(pairs, fmt.Sprintf("%s: %s", key, attributeValueToString(val)))
		}
		return "{" + strings.Join(pairs, ", ") + "}"

	case *types.AttributeValueMemberSS:
		// String Set
		return "[" + strings.Join(v.Value, ", ") + "]"

	case *types.AttributeValueMemberNS:
		// Number Set
		return "[" + strings.Join(v.Value, ", ") + "]"

	case *types.AttributeValueMemberBS:
		// Binary Set
		var items []string
		for _, b := range v.Value {
			items = append(items, fmt.Sprintf("<binary:%d bytes>", len(b)))
		}
		return "[" + strings.Join(items, ", ") + "]"

	case *types.AttributeValueMemberB:
		// Binary
		return fmt.Sprintf("<binary:%d bytes>", len(v.Value))

	default:
		// Fallback to JSON marshaling
		bytes, err := json.Marshal(av)
		if err != nil {
			return "<unknown>"
		}
		return string(bytes)
	}
}

// attributeValueToType returns the type name of an AttributeValue
func attributeValueToType(av types.AttributeValue) string {
	if av == nil {
		return "NULL"
	}

	switch av.(type) {
	case *types.AttributeValueMemberS:
		return "String"
	case *types.AttributeValueMemberN:
		return "Number"
	case *types.AttributeValueMemberBOOL:
		return "Boolean"
	case *types.AttributeValueMemberNULL:
		return "Null"
	case *types.AttributeValueMemberL:
		return "List"
	case *types.AttributeValueMemberM:
		return "Map"
	case *types.AttributeValueMemberSS:
		return "String Set"
	case *types.AttributeValueMemberNS:
		return "Number Set"
	case *types.AttributeValueMemberBS:
		return "Binary Set"
	case *types.AttributeValueMemberB:
		return "Binary"
	default:
		return "Unknown"
	}
}

// truncateString truncates a string to maxLen and adds ellipsis if needed
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
