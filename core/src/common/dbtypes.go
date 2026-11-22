package common
import (
	mapset "github.com/deckarep/golang-set/v2"
)
var (
	IntTypes = mapset.NewSet(
		"INTEGER", "SMALLINT", "BIGINT", "INT", "TINYINT", "MEDIUMINT",
		"INT8", "INT16", "INT32", "INT64",
		"SERIAL", "BIGSERIAL", "SMALLSERIAL",
	)
	UintTypes = mapset.NewSet(
		"TINYINT UNSIGNED", "SMALLINT UNSIGNED", "MEDIUMINT UNSIGNED",
		"INT UNSIGNED", "BIGINT UNSIGNED",
		"UINT8", "UINT16", "UINT32", "UINT64",
	)
	BigIntTypes = mapset.NewSet(
		"INT128", "INT256",
		"UINT128", "UINT256",
	)
	DecimalTypes = mapset.NewSet(
		"NUMERIC", "DECIMAL", "NUMBER", "MONEY", "SMALLMONEY",
		"DECIMAL32", "DECIMAL64", "DECIMAL128", "DECIMAL256",
	)
	FloatTypes = mapset.NewSet(
		"REAL", "DOUBLE PRECISION",
		"FLOAT", "DOUBLE",
		"FLOAT32", "FLOAT64",
		"BINARY_FLOAT", "BINARY_DOUBLE",
	)
	BoolTypes = mapset.NewSet(
		"BOOLEAN",
		"BOOL",
		"BIT",
	)
	DateTypes = mapset.NewSet(
		"DATE",
		"DATE32",
	)
	DateTimeTypes = mapset.NewSet(
		"TIMESTAMP", "TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITH LOCAL TIME ZONE",
		"TIME", "TIME WITH TIME ZONE",
		"DATETIME", "YEAR",
		"DATETIME2", "SMALLDATETIME", "DATETIMEOFFSET",
		"INTERVAL", "INTERVAL YEAR TO MONTH", "INTERVAL DAY TO SECOND",
		"DATETIME64",
	)
	UuidTypes = mapset.NewSet(
		"UUID",
		"UNIQUEIDENTIFIER",
	)
	TextTypes = mapset.NewSet(
		"CHARACTER VARYING", "CHARACTER", "TEXT",
		"VARCHAR", "CHAR",
		"TINYTEXT", "MEDIUMTEXT", "LONGTEXT",
		"STRING", "FIXEDSTRING",
		"NVARCHAR", "NCHAR", "NTEXT",
		"VARCHAR2", "NVARCHAR2", "CLOB", "NCLOB", "LONG",
	)
	JsonTypes = mapset.NewSet(
		"JSON",
		"JSONB",
	)
	BinaryTypes = mapset.NewSet(
		"BYTEA",
		"BINARY", "VARBINARY",
		"TINYBLOB", "BLOB", "MEDIUMBLOB", "LONGBLOB",
		"IMAGE",
		"RAW", "LONG RAW", "BFILE",
	)
	GeometryTypes = mapset.NewSet(
		"POINT", "LINE", "LSEG", "BOX", "PATH", "POLYGON", "CIRCLE",
		"GEOMETRY", "GEOGRAPHY",
	)
	NetworkTypes = mapset.NewSet(
		"CIDR", "INET", "MACADDR", "MACADDR8",
		"IPV4", "IPV6",
	)
	XMLTypes = mapset.NewSet(
		"XML",
		"XMLTYPE",
	)
)