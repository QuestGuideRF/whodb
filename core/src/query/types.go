package query
type WhereConditionType string
const (
	WhereConditionTypeAtomic WhereConditionType = "Atomic"
	WhereConditionTypeAnd WhereConditionType = "And"
	WhereConditionTypeOr WhereConditionType = "Or"
)
type AtomicWhereCondition struct {
	ColumnType string
	Key        string
	Operator   string
	Value      string
}
type OperationWhereCondition struct {
	Children []*WhereCondition
}
type WhereCondition struct {
	Type   WhereConditionType
	Atomic *AtomicWhereCondition
	And    *OperationWhereCondition
	Or     *OperationWhereCondition
}
type SortDirection string
const (
	SortDirectionAsc SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)
type SortCondition struct {
	Column    string
	Direction SortDirection
}