package engine
import "github.com/clidey/whodb/core/src/source"
type StorageUnit = source.StorageUnit
type Record = source.Record
type ExternalModel = source.ExternalModel
type Column = source.Column
type GetRowsResult = source.RowsResult
type GraphUnitRelationshipType = source.GraphRelationshipType
type GraphUnitRelationship = source.GraphRelationship
type GraphUnit = source.GraphUnit
type ChatMessage = source.ChatMessage
type ForeignKeyRelationship = source.ForeignKeyRelationship
type SSLStatus = source.SSLStatus
type TypeCategory = source.TypeCategory
type TypeDefinition = source.TypeDefinition
const (
	TypeCategoryNumeric TypeCategory = source.TypeCategoryNumeric
	TypeCategoryText TypeCategory = source.TypeCategoryText
	TypeCategoryBinary TypeCategory = source.TypeCategoryBinary
	TypeCategoryDatetime TypeCategory = source.TypeCategoryDatetime
	TypeCategoryBoolean TypeCategory = source.TypeCategoryBoolean
	TypeCategoryJSON TypeCategory = source.TypeCategoryJSON
	TypeCategoryOther TypeCategory = source.TypeCategoryOther
)
const (
	GraphUnitRelationshipTypeOneToOne GraphUnitRelationshipType = source.GraphRelationshipTypeOneToOne
	GraphUnitRelationshipTypeOneToMany GraphUnitRelationshipType = source.GraphRelationshipTypeOneToMany
	GraphUnitRelationshipTypeManyToOne GraphUnitRelationshipType = source.GraphRelationshipTypeManyToOne
	GraphUnitRelationshipTypeManyToMany GraphUnitRelationshipType = source.GraphRelationshipTypeManyToMany
	GraphUnitRelationshipTypeUnknown GraphUnitRelationshipType = source.GraphRelationshipTypeUnknown
)