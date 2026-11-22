package elasticsearch
import (
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/query"
	"github.com/clidey/whodb/core/src/log"
)
func convertAtomicConditionToES(atomic *query.AtomicWhereCondition) (map[string]any, error) {
	if atomic == nil {
		return nil, fmt.Errorf("atomic condition is nil")
	}
	switch atomic.Operator {
	case "match", "MATCH":
		return map[string]any{
			"match": map[string]any{
				atomic.Key: atomic.Value,
			},
		}, nil
	case "match_phrase_prefix", "MATCH_PHRASE_PREFIX":
		return map[string]any{
			"match_phrase_prefix": map[string]any{
				atomic.Key: atomic.Value,
			},
		}, nil
	case "=", "eq", "EQ", "equals", "EQUALS", "term", "TERM":
		if atomic.Key == "_id" {
			return map[string]any{
				"ids": map[string]any{
					"values": []any{atomic.Value},
				},
			}, nil
		}
		return map[string]any{
			"term": map[string]any{
				atomic.Key: atomic.Value,
			},
		}, nil
	case "!=", "ne", "NE", "not equals", "NOT EQUALS":
		return map[string]any{
			"bool": map[string]any{
				"must_not": []map[string]any{
					{
						"term": map[string]any{
							atomic.Key: atomic.Value,
						},
					},
				},
			},
		}, nil
	case "exists", "EXISTS":
		return map[string]any{
			"exists": map[string]any{
				"field": atomic.Key,
			},
		}, nil
	case "not exists", "NOT EXISTS":
		return map[string]any{
			"bool": map[string]any{
				"must_not": []map[string]any{
					{
						"exists": map[string]any{
							"field": atomic.Key,
						},
					},
				},
			},
		}, nil
	case ">", "gt", "GT":
		return map[string]any{
			"range": map[string]any{
				atomic.Key: map[string]any{
					"gt": atomic.Value,
				},
			},
		}, nil
	case ">=", "gte", "GTE":
		return map[string]any{
			"range": map[string]any{
				atomic.Key: map[string]any{
					"gte": atomic.Value,
				},
			},
		}, nil
	case "<", "lt", "LT":
		return map[string]any{
			"range": map[string]any{
				atomic.Key: map[string]any{
					"lt": atomic.Value,
				},
			},
		}, nil
	case "<=", "lte", "LTE":
		return map[string]any{
			"range": map[string]any{
				atomic.Key: map[string]any{
					"lte": atomic.Value,
				},
			},
		}, nil
	case "like", "LIKE", "contains", "CONTAINS":
		return map[string]any{
			"wildcard": map[string]any{
				atomic.Key: map[string]any{
					"value": fmt.Sprintf("*%v*", atomic.Value),
				},
			},
		}, nil
	case "prefix", "PREFIX", "starts with", "STARTS WITH":
		return map[string]any{
			"prefix": map[string]any{
				atomic.Key: atomic.Value,
			},
		}, nil
	case "terms", "TERMS":
		values := parseCSVToSlice(atomic.Value)
		return map[string]any{
			"terms": map[string]any{
				atomic.Key: values,
			},
		}, nil
	case "ids", "IDS":
		var ids []any
		ids = parseCSVToSlice(atomic.Value)
		return map[string]any{
			"ids": map[string]any{
				"values": ids,
			},
		}, nil
	case "range", "RANGE":
		minBound, maxBound := parseRangeBounds(fmt.Sprintf("%v", atomic.Value))
		rangeClause := map[string]any{}
		if minBound != "" {
			rangeClause["gte"] = minBound
		}
		if maxBound != "" {
			rangeClause["lte"] = maxBound
		}
		return map[string]any{
			"range": map[string]any{
				atomic.Key: rangeClause,
			},
		}, nil
	case "match_phrase", "MATCH_PHRASE":
		return map[string]any{
			"match_phrase": map[string]any{
				atomic.Key: atomic.Value,
			},
		}, nil
	case "wildcard", "WILDCARD":
		return map[string]any{
			"wildcard": map[string]any{
				atomic.Key: map[string]any{
					"value": atomic.Value,
				},
			},
		}, nil
	case "regexp", "REGEXP":
		return map[string]any{
			"regexp": map[string]any{
				atomic.Key: atomic.Value,
			},
		}, nil
	case "fuzzy", "FUZZY":
		return map[string]any{
			"fuzzy": map[string]any{
				atomic.Key: atomic.Value,
			},
		}, nil
	default:
		log.WithField("operator", atomic.Operator).Warn("Unknown operator, defaulting to match query")
		return map[string]any{
			"match": map[string]any{
				atomic.Key: atomic.Value,
			},
		}, nil
	}
}
func convertWhereConditionToES(where *query.WhereCondition) (map[string]any, error) {
	if where == nil {
		return map[string]any{}, nil
	}
	switch where.Type {
	case query.WhereConditionTypeAtomic:
		if where.Atomic == nil {
			err := fmt.Errorf("atomic condition must have an atomicwherecondition")
			log.WithError(err).Error("Invalid atomic where condition: missing atomic condition")
			return nil, err
		}
		clause, err := convertAtomicConditionToES(where.Atomic)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"must": []map[string]any{clause},
		}, nil
	case query.WhereConditionTypeAnd:
		if where.And == nil || len(where.And.Children) == 0 {
			return map[string]any{}, nil
		}
		mustClauses := []map[string]any{}
		for _, child := range where.And.Children {
			if child.Type == query.WhereConditionTypeAtomic && child.Atomic != nil {
				clause, err := convertAtomicConditionToES(child.Atomic)
				if err != nil {
					log.WithError(err).Error("Failed to convert atomic condition in AND clause to ElasticSearch query")
					return nil, err
				}
				mustClauses = append(mustClauses, clause)
			} else {
				childCondition, err := convertWhereConditionToES(child)
				if err != nil {
					log.WithError(err).Error("Failed to convert child condition in AND clause to ElasticSearch query")
					return nil, err
				}
				mustClauses = append(mustClauses, map[string]any{
					"bool": childCondition,
				})
			}
		}
		return map[string]any{"must": mustClauses}, nil
	case query.WhereConditionTypeOr:
		if where.Or == nil || len(where.Or.Children) == 0 {
			return map[string]any{}, nil
		}
		shouldClauses := []map[string]any{}
		for _, child := range where.Or.Children {
			if child.Type == query.WhereConditionTypeAtomic && child.Atomic != nil {
				clause, err := convertAtomicConditionToES(child.Atomic)
				if err != nil {
					log.WithError(err).Error("Failed to convert atomic condition in OR clause to ElasticSearch query")
					return nil, err
				}
				shouldClauses = append(shouldClauses, clause)
			} else {
				childCondition, err := convertWhereConditionToES(child)
				if err != nil {
					log.WithError(err).Error("Failed to convert child condition in OR clause to ElasticSearch query")
					return nil, err
				}
				shouldClauses = append(shouldClauses, map[string]any{
					"bool": childCondition,
				})
			}
		}
		return map[string]any{
			"should":               shouldClauses,
			"minimum_should_match": 1,
		}, nil
	default:
		err := fmt.Errorf("unknown whereconditiontype: %v", where.Type)
		return nil, err
	}
}
func parseCSVToSlice(raw string) []any {
	if strings.TrimSpace(raw) == "" {
		return []any{}
	}
	parts := strings.Split(raw, ",")
	values := make([]any, 0, len(parts))
	for _, p := range parts {
		values = append(values, strings.TrimSpace(p))
	}
	return values
}
func parseRangeBounds(raw string) (string, string) {
	parts := strings.Split(raw, ",")
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), ""
	}
	minBound := strings.TrimSpace(parts[0])
	maxBound := strings.TrimSpace(parts[1])
	return minBound, maxBound
}