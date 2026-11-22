package source
import (
	"strings"
)
type Category string
const (
	CategoryDatabase Category = "Database"
	CategoryCache Category = "Cache"
	CategorySearch Category = "Search"
	CategoryObjectStore Category = "ObjectStore"
	CategoryFileStore Category = "FileStore"
)
type Model string
const (
	ModelRelational Model = "Relational"
	ModelDocument Model = "Document"
	ModelKeyValue Model = "KeyValue"
	ModelSearch Model = "Search"
	ModelGraph Model = "Graph"
	ModelObject Model = "Object"
)
type Surface string
const (
	SurfaceBrowser Surface = "Browser"
	SurfaceQuery Surface = "Query"
	SurfaceGraph Surface = "Graph"
	SurfaceChat Surface = "Chat"
)
type ObjectKind string
const (
	ObjectKindDatabase ObjectKind = "Database"
	ObjectKindSchema ObjectKind = "Schema"
	ObjectKindTable ObjectKind = "Table"
	ObjectKindView ObjectKind = "View"
	ObjectKindCollection ObjectKind = "Collection"
	ObjectKindIndex ObjectKind = "Index"
	ObjectKindKey ObjectKind = "Key"
	ObjectKindItem ObjectKind = "Item"
	ObjectKindFunction ObjectKind = "Function"
	ObjectKindProcedure ObjectKind = "Procedure"
	ObjectKindTrigger ObjectKind = "Trigger"
	ObjectKindSequence ObjectKind = "Sequence"
)
type Action string
const (
	ActionBrowse Action = "Browse"
	ActionInspect Action = "Inspect"
	ActionViewRows Action = "ViewRows"
	ActionViewContent Action = "ViewContent"
	ActionViewDefinition Action = "ViewDefinition"
	ActionCreateChild Action = "CreateChild"
	ActionDelete Action = "Delete"
	ActionInsertData Action = "InsertData"
	ActionUpdateData Action = "UpdateData"
	ActionDeleteData Action = "DeleteData"
	ActionImportData Action = "ImportData"
	ActionGenerateMockData Action = "GenerateMockData"
	ActionExecute Action = "Execute"
	ActionViewGraph Action = "ViewGraph"
)
type View string
const (
	ViewGrid View = "Grid"
	ViewJSON View = "JSON"
	ViewText View = "Text"
	ViewSQL View = "SQL"
	ViewBinary View = "Binary"
	ViewMetadata View = "Metadata"
	ViewGraph View = "Graph"
)
type DataShape string
const (
	DataShapeTabular DataShape = "Tabular"
	DataShapeDocument DataShape = "Document"
	DataShapeContent DataShape = "Content"
	DataShapeGraph DataShape = "Graph"
	DataShapeMetadata DataShape = "Metadata"
)
type ConnectionFieldKind string
const (
	ConnectionFieldKindText ConnectionFieldKind = "Text"
	ConnectionFieldKindPassword ConnectionFieldKind = "Password"
	ConnectionFieldKindFilePath ConnectionFieldKind = "FilePath"
)
type ConnectionFieldSection string
const (
	ConnectionFieldSectionPrimary ConnectionFieldSection = "Primary"
	ConnectionFieldSectionAdvanced ConnectionFieldSection = "Advanced"
)
type CredentialField string
const (
	CredentialFieldHostname CredentialField = "Hostname"
	CredentialFieldUsername CredentialField = "Username"
	CredentialFieldPassword CredentialField = "Password"
	CredentialFieldDatabase CredentialField = "Database"
	CredentialFieldAdvanced CredentialField = "Advanced"
)
type ConnectionField struct {
	Key             string
	Kind            ConnectionFieldKind
	Section         ConnectionFieldSection
	Required        bool
	LabelKey        string
	PlaceholderKey  string
	DefaultValue    string
	SupportsOptions bool
	CredentialField CredentialField
	AdvancedKey     string
}
type ConnectionTransport string
const (
	ConnectionTransportNetwork ConnectionTransport = "Network"
	ConnectionTransportFile ConnectionTransport = "File"
	ConnectionTransportBridge ConnectionTransport = "Bridge"
)
type HostInputMode string
const (
	HostInputModeNone HostInputMode = "None"
	HostInputModeHostname HostInputMode = "Hostname"
	HostInputModeHostnameOrURL HostInputMode = "HostnameOrURL"
)
type HostInputURLParser string
const (
	HostInputURLParserNone HostInputURLParser = "None"
	HostInputURLParserPostgres HostInputURLParser = "Postgres"
	HostInputURLParserMongoSRV HostInputURLParser = "MongoSRV"
)
type ProfileLabelStrategy string
const (
	ProfileLabelStrategyDefault ProfileLabelStrategy = "Default"
	ProfileLabelStrategyHostname ProfileLabelStrategy = "Hostname"
	ProfileLabelStrategyDatabase ProfileLabelStrategy = "Database"
)
type SchemaFidelity string
const (
	SchemaFidelityExact SchemaFidelity = "Exact"
	SchemaFidelitySampled SchemaFidelity = "Sampled"
)
type QueryExplainMode string
const (
	QueryExplainModeNone QueryExplainMode = "None"
	QueryExplainModeExplain QueryExplainMode = "Explain"
	QueryExplainModeExplainAnalyze QueryExplainMode = "ExplainAnalyze"
	QueryExplainModeExplainPipeline QueryExplainMode = "ExplainPipeline"
)
type ConnectionTraits struct {
	Transport               ConnectionTransport
	HostInputMode           HostInputMode
	HostInputURLParser      HostInputURLParser
	SupportsCustomCAContent bool
}
type PresentationTraits struct {
	ProfileLabelStrategy ProfileLabelStrategy
	SchemaFidelity       SchemaFidelity
}
type QueryTraits struct {
	SupportsAnalyze bool
	ExplainMode     QueryExplainMode
}
type MockDataTraits struct {
	SupportsRelationalDependencies bool
}
type TypeTraits struct {
	Connection   ConnectionTraits
	Presentation PresentationTraits
	Query        QueryTraits
	MockData     MockDataTraits
}
type Contract struct {
	Model             Model
	Surfaces          []Surface
	RootActions       []Action
	BrowsePath        []ObjectKind
	DefaultObjectKind ObjectKind
	GraphScopeKind    *ObjectKind
	ObjectTypes       []ObjectType
}
func (c Contract) SupportsSurface(surface Surface) bool {
	for _, candidate := range c.Surfaces {
		if candidate == surface {
			return true
		}
	}
	return false
}
func (c Contract) SupportsAction(action Action) bool {
	for _, objectType := range c.ObjectTypes {
		if objectType.SupportsAction(action) {
			return true
		}
	}
	return false
}
func (c Contract) ObjectTypeForKind(kind ObjectKind) (ObjectType, bool) {
	for _, objectType := range c.ObjectTypes {
		if objectType.Kind == kind {
			return objectType, true
		}
	}
	return ObjectType{}, false
}
type ObjectType struct {
	Kind          ObjectKind
	DataShape     DataShape
	Actions       []Action
	Views         []View
	SingularLabel string
	PluralLabel   string
}
func (o ObjectType) SupportsAction(action Action) bool {
	for _, candidate := range o.Actions {
		if candidate == action {
			return true
		}
	}
	return false
}
type TypeSpec struct {
	ID               string
	Label            string
	DriverID         string
	Connector        string
	Category         Category
	Traits           TypeTraits
	ConnectionFields []ConnectionField
	Contract         Contract
	DiscoveryPrefill DiscoveryPrefill
	IsAWSManaged     bool
	SSLModes         []SSLModeInfo
}
func (s TypeSpec) ConnectionFieldByKey(key string) (ConnectionField, bool) {
	for _, field := range s.ConnectionFields {
		if strings.EqualFold(field.Key, key) {
			return field, true
		}
	}
	return ConnectionField{}, false
}
type Credentials struct {
	ID          *string           `json:"Id,omitempty"`
	SourceType  string            `json:"SourceType"`
	Values      map[string]string `json:"Values,omitempty"`
	AccessToken *string           `json:"AccessToken,omitempty"`
	IsProfile   bool              `json:"IsProfile,omitempty"`
}
func (c *Credentials) CloneValues() map[string]string {
	if c == nil || c.Values == nil {
		return map[string]string{}
	}
	values := make(map[string]string, len(c.Values))
	for key, value := range c.Values {
		values[key] = value
	}
	return values
}
type ObjectRef struct {
	Kind    ObjectKind
	Locator string
	Path    []string
}
type Object struct {
	Ref         ObjectRef
	Kind        ObjectKind
	Name        string
	Path        []string
	HasChildren bool
	Actions     []Action
	Metadata    []Record
}
type ObjectColumns struct {
	Ref     ObjectRef
	Columns []Column
}
type SessionMetadata struct {
	SourceType      string
	QueryLanguages  []string
	TypeDefinitions []TypeDefinition
	Operators       []string
	AliasMap        map[string]string
}
type Profile struct {
	ID                   string
	DisplayName          string
	SourceType           string
	Values               map[string]string
	IsEnvironmentDefined bool
	Source               string
	SSLConfigured        bool
}