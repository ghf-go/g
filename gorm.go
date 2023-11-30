package g

import (
	"context"
	"database/sql/driver"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MysqlPoint struct {
	Lat, Lng float64
}

func (loc MysqlPoint) GormDataType() string {
	return "point"
}

func (p MysqlPoint) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL:  "ST_PointFromText(?)",
		Vars: []interface{}{fmt.Sprintf("POINT(%v %v)", p.Lat, p.Lng)},
	}
}

func (p MysqlPoint) Value() (driver.Value, error) {
	return []byte(fmt.Sprintf("POINT('%v' '%v')", p.Lat, p.Lng)), nil
}

func (p *MysqlPoint) Scan(value interface{}) error {
	var point string
	switch v := value.(type) {
	case []byte:
		point = string(v)
	case string:
		point = v
	default:
		return fmt.Errorf("failed to scan point: %v", value)
	}
	_, err := fmt.Sscanf(point, "POINT(%f %f)", &p.Lat, &p.Lng)
	return err
}
