package common
import (
	"regexp"
	"strconv"
	"strings"
)
type TypeSpec struct {
	BaseType string
	Length int
	Precision int
	Scale int
}
var typeSpecRegex = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_ ]*?)(?:\((\d+)(?:,\s*(\d+))?\))?$`)
func ParseTypeSpec(fullType string) TypeSpec {
	fullType = strings.TrimSpace(fullType)
	matches := typeSpecRegex.FindStringSubmatch(fullType)
	if matches == nil {
		return TypeSpec{BaseType: strings.ToUpper(fullType)}
	}
	spec := TypeSpec{
		BaseType: strings.ToUpper(strings.TrimSpace(matches[1])),
	}
	if matches[2] != "" {
		if n, err := strconv.Atoi(matches[2]); err == nil {
			if matches[3] != "" {
				spec.Precision = n
			} else {
				spec.Length = n
			}
		}
	}
	if matches[3] != "" {
		if n, err := strconv.Atoi(matches[3]); err == nil {
			spec.Scale = n
			if spec.Length > 0 {
				spec.Precision = spec.Length
				spec.Length = 0
			}
		}
	}
	return spec
}
func FormatTypeSpec(spec TypeSpec) string {
	if spec.Precision > 0 {
		return spec.BaseType + "(" +
			strconv.Itoa(spec.Precision) + "," +
			strconv.Itoa(spec.Scale) + ")"
	}
	if spec.Length > 0 {
		return spec.BaseType + "(" + strconv.Itoa(spec.Length) + ")"
	}
	return spec.BaseType
}
func NormalizeTypeWithMap(typeName string, aliasMap map[string]string) string {
	spec := ParseTypeSpec(typeName)
	upperBase := strings.ToUpper(spec.BaseType)
	if canonical, ok := aliasMap[upperBase]; ok {
		spec.BaseType = canonical
	} else {
		spec.BaseType = upperBase
	}
	return FormatTypeSpec(spec)
}