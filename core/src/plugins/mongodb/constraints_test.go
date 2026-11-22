package mongodb
import (
	"testing"
)
func TestParseMongoDBJsonSchema(t *testing.T) {
	schema := map[string]any{
		"bsonType": "object",
		"required": []any{"name", "price", "status", "category"},
		"properties": map[string]any{
			"name": map[string]any{
				"bsonType":  "string",
				"maxLength": float64(100),
			},
			"price": map[string]any{
				"bsonType": "double",
				"minimum":  float64(0),
				"maximum":  float64(99999.99),
			},
			"status": map[string]any{
				"bsonType": "string",
				"enum":     []any{"active", "inactive", "discontinued"},
			},
			"category": map[string]any{
				"bsonType": "string",
				"enum":     []any{"electronics", "clothing", "food", "other"},
			},
			"stock_quantity": map[string]any{
				"bsonType": "int",
				"minimum":  float64(0),
				"maximum":  float64(10000),
			},
			"description": map[string]any{
				"bsonType":  "string",
				"maxLength": float64(500),
			},
			"rating": map[string]any{
				"bsonType": "double",
				"minimum":  float64(0),
				"maximum":  float64(5),
			},
		},
	}
	constraints := parseMongoDBJsonSchema(schema)
	t.Run("required fields are not nullable", func(t *testing.T) {
		requiredFields := []string{"name", "price", "status", "category"}
		for _, field := range requiredFields {
			if constraints[field] == nil {
				t.Errorf("expected constraints for required field %q", field)
				continue
			}
			if nullable, ok := constraints[field]["nullable"].(bool); !ok || nullable {
				t.Errorf("expected field %q to have nullable=false, got %v", field, constraints[field]["nullable"])
			}
		}
	})
	t.Run("optional fields are nullable", func(t *testing.T) {
		optionalFields := []string{"stock_quantity", "description", "rating"}
		for _, field := range optionalFields {
			if constraints[field] == nil {
				t.Errorf("expected constraints for optional field %q", field)
				continue
			}
			if nullable, ok := constraints[field]["nullable"].(bool); !ok || !nullable {
				t.Errorf("expected field %q to have nullable=true, got %v", field, constraints[field]["nullable"])
			}
		}
	})
	t.Run("enum values extracted as check_values", func(t *testing.T) {
		statusValues, ok := constraints["status"]["check_values"].([]string)
		if !ok {
			t.Fatalf("expected check_values for status field")
		}
		expected := []string{"active", "inactive", "discontinued"}
		if len(statusValues) != len(expected) {
			t.Errorf("expected %d status values, got %d", len(expected), len(statusValues))
		}
		for i, v := range expected {
			if statusValues[i] != v {
				t.Errorf("expected status value %q at index %d, got %q", v, i, statusValues[i])
			}
		}
		categoryValues, ok := constraints["category"]["check_values"].([]string)
		if !ok {
			t.Fatalf("expected check_values for category field")
		}
		expectedCat := []string{"electronics", "clothing", "food", "other"}
		if len(categoryValues) != len(expectedCat) {
			t.Errorf("expected %d category values, got %d", len(expectedCat), len(categoryValues))
		}
	})
	t.Run("numeric min/max extracted as check_min/check_max", func(t *testing.T) {
		if min, ok := constraints["price"]["check_min"].(float64); !ok || min != 0 {
			t.Errorf("expected price check_min=0, got %v", constraints["price"]["check_min"])
		}
		if max, ok := constraints["price"]["check_max"].(float64); !ok || max != 99999.99 {
			t.Errorf("expected price check_max=99999.99, got %v", constraints["price"]["check_max"])
		}
		if min, ok := constraints["rating"]["check_min"].(float64); !ok || min != 0 {
			t.Errorf("expected rating check_min=0, got %v", constraints["rating"]["check_min"])
		}
		if max, ok := constraints["rating"]["check_max"].(float64); !ok || max != 5 {
			t.Errorf("expected rating check_max=5, got %v", constraints["rating"]["check_max"])
		}
		if min, ok := constraints["stock_quantity"]["check_min"].(float64); !ok || min != 0 {
			t.Errorf("expected stock_quantity check_min=0, got %v", constraints["stock_quantity"]["check_min"])
		}
		if max, ok := constraints["stock_quantity"]["check_max"].(float64); !ok || max != 10000 {
			t.Errorf("expected stock_quantity check_max=10000, got %v", constraints["stock_quantity"]["check_max"])
		}
	})
	t.Run("maxLength extracted as length", func(t *testing.T) {
		if length, ok := constraints["name"]["length"].(int); !ok || length != 100 {
			t.Errorf("expected name length=100, got %v", constraints["name"]["length"])
		}
		if length, ok := constraints["description"]["length"].(int); !ok || length != 500 {
			t.Errorf("expected description length=500, got %v", constraints["description"]["length"])
		}
	})
	t.Run("bsonType extracted as type", func(t *testing.T) {
		if typ, ok := constraints["name"]["type"].(string); !ok || typ != "string" {
			t.Errorf("expected name type=string, got %v", constraints["name"]["type"])
		}
		if typ, ok := constraints["price"]["type"].(string); !ok || typ != "double" {
			t.Errorf("expected price type=double, got %v", constraints["price"]["type"])
		}
		if typ, ok := constraints["stock_quantity"]["type"].(string); !ok || typ != "int" {
			t.Errorf("expected stock_quantity type=int, got %v", constraints["stock_quantity"]["type"])
		}
	})
}
func TestParseMongoDBJsonSchemaEmptySchema(t *testing.T) {
	constraints := parseMongoDBJsonSchema(map[string]any{})
	if len(constraints) != 0 {
		t.Errorf("expected empty constraints for empty schema, got %v", constraints)
	}
}
func TestParseMongoDBJsonSchemaNoProperties(t *testing.T) {
	schema := map[string]any{
		"bsonType": "object",
		"required": []any{"field1"},
	}
	constraints := parseMongoDBJsonSchema(schema)
	if len(constraints) != 0 {
		t.Errorf("expected empty constraints when no properties defined, got %v", constraints)
	}
}
func TestParseMongoDBJsonSchemaInt32Values(t *testing.T) {
	schema := map[string]any{
		"bsonType": "object",
		"properties": map[string]any{
			"rating": map[string]any{
				"bsonType": "double",
				"minimum":  int32(0),
				"maximum":  int32(5),
			},
			"name": map[string]any{
				"bsonType":  "string",
				"maxLength": int32(100),
			},
		},
	}
	constraints := parseMongoDBJsonSchema(schema)
	t.Run("int32 minimum/maximum converted to float64", func(t *testing.T) {
		if minVal, ok := constraints["rating"]["check_min"].(float64); !ok || minVal != 0 {
			t.Errorf("expected rating check_min=0 (float64), got %v (%T)", constraints["rating"]["check_min"], constraints["rating"]["check_min"])
		}
		if maxVal, ok := constraints["rating"]["check_max"].(float64); !ok || maxVal != 5 {
			t.Errorf("expected rating check_max=5 (float64), got %v (%T)", constraints["rating"]["check_max"], constraints["rating"]["check_max"])
		}
	})
	t.Run("int32 maxLength converted to int", func(t *testing.T) {
		if length, ok := constraints["name"]["length"].(int); !ok || length != 100 {
			t.Errorf("expected name length=100 (int), got %v (%T)", constraints["name"]["length"], constraints["name"]["length"])
		}
	})
}