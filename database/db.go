package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/tianxinbaiyun/mysqlcompare/config"

	_ "github.com/go-sql-driver/mysql" //mysql
)

// DB 数据库定义
var DB = make(map[string]*sql.DB)

// InitDB 初始化连接
func InitDB() {
	GetDB(config.C.Source)
	GetDB(config.C.Destination)
}

// GetDB 获取连接
func GetDB(conn config.Conn) *sql.DB {
	if _, ok := DB[conn.Host]; ok {
		return DB[conn.Host]
	}
	//root:root@tcp(127.0.0.1:3306)/test
	dsn := conn.User + ":" + conn.Pass + "@tcp(" + conn.Host + ":" + conn.Port + ")/" + conn.Database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	DB[conn.Host] = db
	return db
}

// GetCount 获取数据库数量
func GetCount(db *sql.DB, tableName string) (count int64, err error) {
	var rows *sql.Rows
	var sqlStr = "select count(*) as count from " + tableName
	rows, err = db.Query(sqlStr)
	if err != nil {
		return
	}
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return
		}
	}
	err = rows.Close()
	if err != nil {
		return
	}
	return
}

// GetDBName 获取库名
func GetDBName() string {
	return config.C.Source.Database
}

// GetFieldList 获取数据库表字段列表
func GetFieldList(db *sql.DB, tableName string) (fields []string, err error) {
	sqlStr := "SELECT column_name FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s' ORDER BY ordinal_position"
	sqlStr = fmt.Sprintf(sqlStr, GetDBName(), tableName)
	rows, err := db.Query(sqlStr)
	if err != nil {
		return
	}
	for rows.Next() {
		field := ""
		err = rows.Scan(&field)
		if err != nil {
			return
		}
		fields = append(fields, field)
	}
	return
}

// GetRows 查询数据
func GetRows(db *sql.DB, table config.TableInfo, offset int64, size int64) (ret [][]string, newOffset int64, err error) {
	var rows *sql.Rows

	// 条件字符串拼接
	sl := make([]string, 0)
	sl = append(sl, table.Where...)
	andWhere := strings.Join(sl[:], " and ")
	andWhere = strings.Trim(andWhere, "and")
	var sqlStr = fmt.Sprintf("select * from %s", table.Name)
	if andWhere != "" && strings.Trim(andWhere, " ") != "1" {
		sqlStr = fmt.Sprintf("%s where %s", sqlStr, andWhere)
	}
	if size > 0 {
		sqlStr = fmt.Sprintf("%s limit %d,%d", sqlStr, offset, size)
	}

	// 查询数据
	rows, err = db.Query(sqlStr)
	if err != nil {
		return
	}

	columns, err := rows.Columns()
	if err != nil {
		return
	}
	colSize := len(columns)
	pts := make([]interface{}, colSize)
	container := make([]interface{}, colSize)
	for i := range pts {
		pts[i] = &container[i]
	}

	for rows.Next() {
		err = rows.Scan(pts...)
		if err != nil {
			return
		}
		arr := nilToString(container)
		ret = append(ret, arr)
	}
	err = rows.Close()
	if err != nil {
		return nil, 0, err
	}
	if newOffset == 0 {
		newOffset = offset + size
	}
	log.Println("Fetched offset:", offset, " - size:", size)
	return ret, newOffset, nil
}

// Insert 插入数据
func Insert(db *sql.DB, tableName string, row []string) (rowCount int64, err error) {
	// 拼凑values
	values := ""
	for _, s := range row {
		if s != "nil" {
			values = fmt.Sprintf("%s,%s", values, convertString(s))
		} else {
			values = fmt.Sprintf("%s,null", values)
		}
	}
	values = strings.Trim(values, ",")

	// 拼凑成sql字符串
	var sqlStr = fmt.Sprintf("insert into %s values ( %s )", tableName, values)

	// 执行语句
	var ret sql.Result
	ret, err = db.Exec(sqlStr)
	if err != nil {
		return
	}
	rowCount, err = ret.RowsAffected()
	if err != nil {
		return
	}

	return
}

// TruncateTable 清空表数据
func TruncateTable(db *sql.DB, table config.TableInfo) (err error) {
	var sqlStr = "truncate table " + table.Name
	log.Println(sqlStr)
	_, err = db.Exec(sqlStr)
	if err != nil {
		return err
	}
	return nil
}

// convertString 字段内容单引号转换
func convertString(arg string) string {
	var buf strings.Builder
	buf.WriteRune('\'')
	for _, c := range arg {
		if c == '\\' || c == '\'' {
			buf.WriteRune('\\')
		}
		buf.WriteRune(c)
	}
	buf.WriteRune('\'')
	return buf.String()
}

// nilToString 转成字符串
func nilToString(columns []interface{}) []string {
	var strCln []string
	for _, column := range columns {
		if column == nil {
			column = []uint8{'n', 'i', 'l'}
		}
		strCln = append(strCln, string(column.([]uint8)))
	}
	return strCln
}
