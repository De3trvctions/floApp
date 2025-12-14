package model

import (
	"floapp/standard-library/utility"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type MeterReading struct {
	Id          int64     `orm:"auto"`
	Nmi         string    `orm:"size(10);unique(nmi_ts)"`
	Timestamp   time.Time `orm:"type(datetime);unique(nmi_ts)"`
	Consumption float64   `orm:"digits(20);decimals(6)"`
}

func init() {
	orm.RegisterModel(new(MeterReading))
}

func (m *MeterReading) TableName() string {
	return "meter_readings"
}

func UpsertBatchSQL(rows []MeterReading, tx *utility.TxOrm) error {
	if len(rows) == 0 {
		return nil
	}

	var values []string
	for _, r := range rows {
		values = append(values, fmt.Sprintf(
			"('%s','%s',%f)",
			escape(r.Nmi),
			r.Timestamp.Format("2006-01-02 15:04:05"),
			r.Consumption,
		))
	}

	sql := `
	INSERT INTO meter_readings (nmi, timestamp, consumption)
	VALUES ` + strings.Join(values, ",") + `
	ON DUPLICATE KEY UPDATE
	consumption = VALUES(consumption);
	`

	_, err := tx.Raw(sql).Exec()
	return err
}

func escape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
