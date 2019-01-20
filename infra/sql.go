package infra

import (
	"database/sql"

	"github.com/tadoku/api/interfaces/rdb"

	// Postgres driver that's used to connect to the db
	_ "github.com/lib/pq"
)

// NewSQLHandler creates an interface to run queries on a database
func NewSQLHandler(db *RDB) rdb.SQLHandler {
	return &sqlHandler{db: db}
}

type sqlHandler struct {
	db *RDB
}

func (handler *sqlHandler) Execute(statement string, args ...interface{}) (rdb.Result, error) {
	res := sqlResult{}
	result, err := handler.db.Exec(statement, args...)
	if err != nil {
		return res, err
	}
	res.Result = result

	return res, nil
}

func (handler *sqlHandler) NamedExecute(statement string, arg interface{}) (rdb.Result, error) {
	res := sqlResult{}
	result, err := handler.db.NamedExec(statement, arg)
	if err != nil {
		return res, err
	}
	res.Result = result

	return res, nil
}

func (handler *sqlHandler) Query(statement string, args ...interface{}) (rdb.Row, error) {
	row := new(sqlRow)
	rows, err := handler.db.Query(statement, args...)
	if err != nil {
		return row, err
	}
	row.Rows = rows

	return row, nil
}

type sqlResult struct {
	Result sql.Result
}

func (r sqlResult) LastInsertId() (int64, error) {
	return r.Result.LastInsertId()
}

func (r sqlResult) RowsAffected() (int64, error) {
	return r.Result.RowsAffected()
}

type sqlRow struct {
	Rows *sql.Rows
}

func (r sqlRow) Scan(dest ...interface{}) error {
	return r.Rows.Scan(dest...)
}

func (r sqlRow) Next() bool {
	return r.Rows.Next()
}

func (r sqlRow) Close() error {
	return r.Rows.Close()
}
