package specs
import "github.com/clidey/whodb/core/src/engine"
var ElasticSearchSupportedOperators = map[string]string{
	"match": "match", "match_phrase": "match_phrase", "match_phrase_prefix": "match_phrase_prefix", "multi_match": "multi_match",
	"bool": "bool", "term": "term", "terms": "terms", "range": "range", "exists": "exists", "prefix": "prefix", "wildcard": "wildcard",
	"regexp": "regexp", "fuzzy": "fuzzy", "ids": "ids", "constant_score": "constant_score", "function_score": "function_score",
	"dis_max": "dis_max", "nested": "nested", "has_child": "has_child", "has_parent": "has_parent",
}
var ElasticSearchTypeDefinitions = []engine.TypeDefinition{
	{ID: "text", Label: "text", Category: engine.TypeCategoryText},
	{ID: "keyword", Label: "keyword", Category: engine.TypeCategoryText},
	{ID: "boolean", Label: "boolean", Category: engine.TypeCategoryBoolean},
	{ID: "long", Label: "long", Category: engine.TypeCategoryNumeric},
	{ID: "double", Label: "double", Category: engine.TypeCategoryNumeric},
	{ID: "date", Label: "date", Category: engine.TypeCategoryDatetime},
	{ID: "object", Label: "object", Category: engine.TypeCategoryOther},
	{ID: "array", Label: "array", Category: engine.TypeCategoryOther},
	{ID: "geo_point", Label: "geo_point", Category: engine.TypeCategoryOther},
	{ID: "nested", Label: "nested", Category: engine.TypeCategoryOther},
	{ID: "mixed", Label: "mixed", Category: engine.TypeCategoryOther},
}
var MongoDBSupportedOperators = map[string]string{
	"eq": "eq", "ne": "ne", "gt": "gt", "gte": "gte", "lt": "lt", "lte": "lte",
	"in": "in", "nin": "nin", "and": "and", "or": "or", "not": "not", "nor": "nor",
	"exists": "exists", "type": "type", "regex": "regex", "expr": "expr", "mod": "mod",
	"all": "all", "elemMatch": "elemMatch", "size": "size", "bitsAllClear": "bitsAllClear",
	"bitsAllSet": "bitsAllSet", "bitsAnyClear": "bitsAnyClear", "bitsAnySet": "bitsAnySet",
	"geoIntersects": "geoIntersects", "geoWithin": "geoWithin", "near": "near", "nearSphere": "nearSphere",
}
var MongoDBTypeDefinitions = []engine.TypeDefinition{
	{ID: "string", Label: "string", Category: engine.TypeCategoryText},
	{ID: "int", Label: "int", Category: engine.TypeCategoryNumeric},
	{ID: "double", Label: "double", Category: engine.TypeCategoryNumeric},
	{ID: "bool", Label: "bool", Category: engine.TypeCategoryBoolean},
	{ID: "date", Label: "date", Category: engine.TypeCategoryDatetime},
	{ID: "objectId", Label: "objectId", Category: engine.TypeCategoryOther},
	{ID: "array", Label: "array", Category: engine.TypeCategoryOther},
	{ID: "object", Label: "object", Category: engine.TypeCategoryOther},
	{ID: "mixed", Label: "mixed", Category: engine.TypeCategoryOther},
}
var RedisOperators = map[string]string{
	"=":           "=",
	"!=":          "!=",
	"<>":          "!=",
	">":           ">",
	">=":          ">=",
	"<":           "<",
	"<=":          "<=",
	"CONTAINS":    "CONTAINS",
	"STARTS WITH": "STARTS WITH",
	"ENDS WITH":   "ENDS WITH",
	"IN":          "IN",
	"NOT IN":      "NOT IN",
}
var MemcachedOperators = map[string]string{
	"=":           "=",
	"!=":          "!=",
	"<>":          "!=",
	">":           ">",
	">=":          ">=",
	"<":           "<",
	"<=":          "<=",
	"CONTAINS":    "CONTAINS",
	"STARTS WITH": "STARTS WITH",
	"ENDS WITH":   "ENDS WITH",
	"IN":          "IN",
	"NOT IN":      "NOT IN",
}