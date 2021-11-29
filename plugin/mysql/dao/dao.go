package dao

import (
	"context"
	"database/sql"
	"github.com/cheivin/dio/plugin/mysql/wrapper"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/utils"
	"strings"
)

type Dao struct {
	db  *gorm.DB
	dst interface{}
}

func (Dao) BeanName() string {
	return "mysqlDao"
}

func New(db *gorm.DB) *Dao {
	return &Dao{
		db: db,
	}
}

func (dao *Dao) withDB(db *gorm.DB) *Dao {
	return &Dao{
		db:  db,
		dst: dao.dst,
	}
}

func (dao *Dao) Model(value interface{}) *Dao {
	d := New(dao.db.Model(value))
	d.dst = value
	return d
}

func (dao *Dao) Table(name string, args ...interface{}) *Dao {
	return dao.withDB(dao.db.Table(name, args...))
}

func (dao *Dao) Ctx(ctx context.Context) *Dao {
	return dao.withDB(dao.db.WithContext(ctx))
}

func (dao *Dao) Select(query interface{}, args ...interface{}) *Dao {
	return dao.withDB(dao.db.Select(query, args...))
}

func (dao *Dao) Distinct(args ...interface{}) *Dao {
	return dao.withDB(dao.db.Distinct(args...))
}

func (dao *Dao) Transaction(fc func(tx *Dao) error, opts ...*sql.TxOptions) error {
	return dao.db.Transaction(func(tx *gorm.DB) error {
		return fc(dao.withDB(tx))
	}, opts...)
}

func (dao *Dao) AutoMigrate(dst interface{}, settings ...map[string]interface{}) error {
	db := dao.db
	if len(settings) > 0 {
		for i := range settings {
			for k, v := range settings[i] {
				db = db.Set(k, v)
			}
		}
	}
	return db.AutoMigrate(dst)
}

func (dao *Dao) DB() *gorm.DB {
	return dao.db
}

func (dao *Dao) Where(wrapper *wrapper.Query) *gorm.DB {
	return dao.scopeQueryAndOrder(wrapper)
}

func (dao *Dao) scopeQuery(wrapper *wrapper.Query) *gorm.DB {
	if wrapper == nil {
		return dao.db
	}
	fragments := wrapper.Build()
	query, args, groupBys, _ := fragments[0].(string), fragments[1].([]interface{}), fragments[2].([]string), fragments[3].(string)
	db := dao.db
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(groupBys) > 0 {
		groupBy := clause.GroupBy{Columns: make([]clause.Column, len(groupBys))}
		for i := range groupBys {
			group := groupBys[i]
			fields := strings.FieldsFunc(group, utils.IsValidDBNameChar)
			groupBy.Columns[i] = clause.Column{Name: group, Raw: len(fields) != 1}
		}
		db.Statement.AddClause(groupBy)
	}
	return db
}

func (dao *Dao) scopeQueryAndOrder(wrapper *wrapper.Query) *gorm.DB {
	if wrapper == nil {
		return dao.db
	}
	fragments := wrapper.Build()
	query, args, groupBys, orderBy := fragments[0].(string), fragments[1].([]interface{}), fragments[2].([]string), fragments[3].(string)
	db := dao.db
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(groupBys) > 0 {
		groupBy := clause.GroupBy{Columns: make([]clause.Column, len(groupBys))}
		for i := range groupBys {
			group := groupBys[i]
			fields := strings.FieldsFunc(group, utils.IsValidDBNameChar)
			groupBy.Columns[i] = clause.Column{Name: group, Raw: len(fields) != 1}
		}
		db.Statement.AddClause(groupBy)
	}
	if orderBy != "" {
		db = db.Order(orderBy)
	}
	return db
}

// 查询

func (dao *Dao) FindOne(cause *wrapper.Query, target interface{}) (ok bool, err error) {
	if cause == nil {
		cause = wrapper.Q()
	}
	db := dao.scopeQuery(cause).Limit(1).Find(target)
	return db.RowsAffected > 0, db.Error
}

func (dao *Dao) FindAll(cause *wrapper.Query, target interface{}) error {
	return dao.Where(cause).Find(target).Error
}

func (dao *Dao) List(cause *wrapper.Query, target interface{}, limit ...int) error {
	db := dao.scopeQueryAndOrder(cause)
	switch len(limit) {
	case 2:
		db = db.Offset(limit[0]).Limit(limit[1])
	case 1:
		db = db.Offset(0).Limit(limit[0])
	}
	return db.Find(target).Error
}

func (dao *Dao) Page(cause *wrapper.Query, target interface{}, page, size int) (total int64, err error) {
	err = dao.scopeQuery(cause).Count(&total).Error
	if err != nil {
		return
	}
	err = dao.scopeQueryAndOrder(cause).
		Offset(page * size).Limit(size).
		Find(target).
		Error
	return
}

func (dao *Dao) Delete(db *gorm.DB, cause *wrapper.Query) (int64, error) {
	db = dao.scopeQuery(cause).Delete(dao.dst)
	return db.RowsAffected, db.Error
}

func (dao *Dao) Insert(value interface{}) error {
	return dao.db.Create(value).Error
}

func (dao *Dao) Update(update *wrapper.Update) (int64, error) {
	if update == nil {
		update = wrapper.U()
	}
	db := dao.scopeQuery(update.Query()).Updates(update.Data())
	return db.RowsAffected, db.Error
}

func (dao *Dao) Sum(field string, cause *wrapper.Query, target interface{}) (err error) {
	return dao.Select("COALESCE(SUM(" + field + "), 0)").scopeQuery(cause).Scan(target).Error
}
