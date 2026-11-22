package graph
import (
	"github.com/clidey/whodb/core/graph/model"
	"github.com/clidey/whodb/core/src/query"
)
func queryWhereConditionFromModel(condition *model.WhereCondition) *query.WhereCondition {
	if condition == nil {
		return nil
	}
	return &query.WhereCondition{
		Type:   query.WhereConditionType(condition.Type),
		Atomic: queryAtomicWhereConditionFromModel(condition.Atomic),
		And:    queryOperationWhereConditionFromModel(condition.And),
		Or:     queryOperationWhereConditionFromModel(condition.Or),
	}
}
func queryAtomicWhereConditionFromModel(condition *model.AtomicWhereCondition) *query.AtomicWhereCondition {
	if condition == nil {
		return nil
	}
	return &query.AtomicWhereCondition{
		ColumnType: condition.ColumnType,
		Key:        condition.Key,
		Operator:   condition.Operator,
		Value:      condition.Value,
	}
}
func queryOperationWhereConditionFromModel(condition *model.OperationWhereCondition) *query.OperationWhereCondition {
	if condition == nil {
		return nil
	}
	children := make([]*query.WhereCondition, 0, len(condition.Children))
	for _, child := range condition.Children {
		children = append(children, queryWhereConditionFromModel(child))
	}
	return &query.OperationWhereCondition{
		Children: children,
	}
}
func querySortConditionsFromModel(sort []*model.SortCondition) []*query.SortCondition {
	conditions := make([]*query.SortCondition, 0, len(sort))
	for _, item := range sort {
		if item == nil {
			conditions = append(conditions, nil)
			continue
		}
		conditions = append(conditions, &query.SortCondition{
			Column:    item.Column,
			Direction: query.SortDirection(item.Direction),
		})
	}
	return conditions
}