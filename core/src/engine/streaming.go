package engine
type QueryStreamWriter interface {
	WriteColumns(columns []Column) error
	WriteRow(row []string) error
}
type QueryStreamer interface {
	StreamRawExecute(config *PluginConfig, query string, writer QueryStreamWriter, params ...any) error
}