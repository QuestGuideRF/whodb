package mysql
func (p *MySQLPlugin) GetPrimaryKeyColQuery() string {
	return `
		SELECT k.COLUMN_NAME
		FROM information_schema.TABLE_CONSTRAINTS t
		JOIN information_schema.KEY_COLUMN_USAGE k
		ON k.CONSTRAINT_NAME = t.CONSTRAINT_NAME
		AND k.TABLE_SCHEMA = t.TABLE_SCHEMA
		AND k.TABLE_NAME = t.TABLE_NAME
		WHERE t.CONSTRAINT_TYPE = 'PRIMARY KEY'
		AND t.TABLE_SCHEMA = ?
		AND t.TABLE_NAME = ?
		ORDER BY k.ORDINAL_POSITION;
	`
}