package utility

import (
	"context"
	"errors"
	"floapp/standard-library/consts"
	"fmt"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

var ErrorFieldsIllegal = errors.New("DB Illigal field")

type DB struct {
	orm.Ormer
	ctx context.Context
}

func NewDB() *DB {
	return Orm()
}

func Orm(aliasName ...string) *DB {
	name := "default"
	if len(aliasName) != 0 && aliasName[0] != "" {
		name = aliasName[0]
	}
	return &DB{orm.NewOrmUsingDB(name), context.Background()}
}

func (d *DB) Get(m any, cols ...string) (err error) {
	return d.ReadWithCtx(d.getCtx(), m, cols...)
}

// Begin 创建事务
func (d *DB) Begin() (*TxOrm, error) {
	tx, err := d.Ormer.BeginWithCtx(d.getCtx())
	if err != nil {
		return nil, err
	}
	return &TxOrm{TxOrmer: tx}, nil
}

// Count 数量
func (d *DB) Count(i any, fields ...any) (count int64, err error) {
	err = verifyFields(fields)
	if err != nil {
		return
	}
	qs := d.QueryTable(i)
	for i := 0; i < len(fields)/2; i++ {
		qs = qs.Filter(fields[i*2+0].(string), fields[i*2+1])
	}
	count, err = qs.CountWithCtx(d.getCtx())
	return
}

// 校验filter是否符合规定
func verifyFields(fields []any) error {
	if len(fields)%2 != 0 {
		return ErrorFieldsIllegal
	}
	return nil
}

func (d *DB) getCtx() context.Context {
	if d.ctx == nil {
		return context.Background()
	}
	return d.ctx
}

type TxOrm struct {
	orm.TxOrmer
	ctx context.Context
}

func (tx *TxOrm) getCtx() context.Context {
	if tx.ctx == nil {
		return context.Background()
	}
	return tx.ctx
}

func (tx *TxOrm) Get(m any, cols ...string) (err error) {
	return tx.ReadWithCtx(tx.getCtx(), m, cols...)
}

// Count 数量
func (tx *TxOrm) Count(i any, fields ...any) (count int64, err error) {
	err = verifyFields(fields)
	if err != nil {
		return
	}
	qs := tx.QueryTable(i)
	for i := 0; i < len(fields)/2; i++ {
		qs = qs.Filter(fields[i*2+0].(string), fields[i*2+1])
	}
	count, err = qs.CountWithCtx(tx.getCtx())
	return
}

// ToLike generates a LIKE pattern for SQL queries
func ToLike(param string, matchType string) string {
	switch matchType {
	case consts.TO_LIKE_CONTAINS:
		return fmt.Sprintf("%%%s%%", param) // Matches anywhere in the string
	case consts.TO_LIKE_START_WITH:
		return fmt.Sprintf("%s%%", param) // Matches strings that start with value
	case consts.TO_LIKE_END_WITH:
		return fmt.Sprintf("%%%s", param) // Matches strings that end with value
	default:
		return fmt.Sprintf("%%%s%%", param) // Default to "contains"
	}
}

// BuildGroupBy dynamically constructs the GROUP BY clause
func BuildGroupBy(qb orm.QueryBuilder, groupByColumn string, groupByColumnPrefix ...string) orm.QueryBuilder {
	if len(groupByColumnPrefix) > 0 {
		qb.GroupBy(strings.Join(groupByColumnPrefix, ", "))
	} else if groupByColumn != "" {
		columns := strings.Split(groupByColumn, ",")
		qb.GroupBy(strings.Join(columns, ", "))
	}
	return qb
}

// ToIn generates an IN clause with placeholders based on array length
// It returns the modified QueryBuilder and the values to be used in SetArgs
func ToIn(qb orm.QueryBuilder, field string, values []string) orm.QueryBuilder {
	if len(values) == 0 {
		// Handle empty array case
		return qb.And(field + " IN (NULL)")
	}

	// Create placeholders for the IN clause
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	return qb.And(field + " IN (" + strings.Join(placeholders, ", ") + ")")
}
