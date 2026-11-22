package postgres
import (
	"github.com/twpayne/go-geom/encoding/ewkb"
	"github.com/twpayne/go-geom/encoding/wkt"
)
func (p *PostgresPlugin) FormatGeometryValue(rawBytes []byte, columnType string) string {
	if len(rawBytes) == 0 {
		return ""
	}
	if geom, err := ewkb.Unmarshal(rawBytes); err == nil {
		if wktStr, err := wkt.Marshal(geom); err == nil {
			return wktStr
		}
	}
	return string(rawBytes)
}