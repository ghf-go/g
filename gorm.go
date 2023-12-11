package g

import (
	"context"
	"database/sql/driver"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type MysqlPoint struct {
	Lat, Lng float64
}

func (loc MysqlPoint) GormDataType() string {
	return "point"
}
func (p MysqlPoint) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql", "sqlite":
		return "POINT"
	}
	return ""
}

func (p MysqlPoint) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL:  "ST_PointFromText(?)",
		Vars: []interface{}{fmt.Sprintf("POINT(%v %v)", p.Lat, p.Lng)},
	}
}

func (p MysqlPoint) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT('%v' '%v')", p.Lat, p.Lng), nil
}

func (p *MysqlPoint) Scan(value interface{}) error {
	var point string
	switch v := value.(type) {
	case []byte:
		point = string(v)
		fmt.Println("byte", point, v)
	case string:
		point = v
		fmt.Println("str", point)
	default:
		return fmt.Errorf("failed to scan point: %v", value)
	}
	var x, y float64
	_, err := fmt.Sscanf(point, "POINT(%f,%f)", &x, &y)
	p.Lat = x
	p.Lng = y
	fmt.Println(err)
	return err
}
