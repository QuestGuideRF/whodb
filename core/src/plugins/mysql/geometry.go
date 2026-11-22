package mysql
import (
	"encoding/binary"
	"fmt"
	"github.com/twpayne/go-geom/encoding/wkb"
	"github.com/twpayne/go-geom/encoding/wkt"
)
func (p *MySQLPlugin) FormatGeometryValue(rawBytes []byte, columnType string) string {
	if len(rawBytes) == 0 {
		return ""
	}
	if len(rawBytes) > 4 {
		srid := binary.LittleEndian.Uint32(rawBytes[:4])
		wkbData := rawBytes[4:]
		if geom, err := wkb.Unmarshal(wkbData); err == nil {
			if wktStr, err := wkt.Marshal(geom); err == nil {
				if srid != 0 {
					return fmt.Sprintf("%s (SRID:%d)", wktStr, srid)
				}
				return wktStr
			}
		}
	}
	return string(rawBytes)
}