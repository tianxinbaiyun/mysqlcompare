package service

import (
	"database/sql"
	"github.com/tianxinbaiyun/mysqlcompare/config"
	"github.com/tianxinbaiyun/mysqlcompare/database"
	"log"
)

// TableField 字段结构体
type TableField struct {
	Name   string // 字段名称
	Omit   bool   // 是否忽略
	Unique bool   // 是否唯一
}

// Compare 同步函数
func Compare() {

	// 变量定义
	var (
		err error
		//affectID             int64
		srcOffset, dstOffset int64
		fistFlag             bool
		srcRows              = make([][]string, 0)
		dstRows              = make([][]string, 0)
		//srcMaps              = map[string][]string{}
		dstMaps = map[string][]string{}
	)

	// 读取配置文件到struct,初始化变量
	config.InitConfig()

	//连接数据库
	dstDB := database.GetDB(config.C.Destination)
	srcDB := database.GetDB(config.C.Source)

	//同步数据
	for _, table := range config.C.Table {

		fistFlag = true
		syncCount := 0
		uniqueI := 0
		fields := map[int]*TableField{}

		uniqueI, fields, err = GetFields(srcDB, table)
		if err != nil {
			log.Println(err)
			return
		}

		for fistFlag || (table.Paging && len(srcRows) > 0) {

			// 如果不分页，设置offset，Batch 为0，查询所有数据
			if !table.Paging {
				srcOffset = 0
				dstOffset = 0
				table.Batch = 0
			}

			// 获取src数据
			srcRows, srcOffset, err = database.GetRows(srcDB, table, srcOffset, table.Batch)
			if err != nil {
				log.Println("err:", err)
				return
			}

			// 获取dst数据
			dstRows, dstOffset, err = database.GetRows(dstDB, table, dstOffset, table.Batch)
			if err != nil {
				log.Println("err:", err)
				return
			}
			// 转成map列表,以唯一值为key
			for _, row := range dstRows {
				dstMaps[row[uniqueI]] = row
			}

			rowLen := len(srcRows)

			if rowLen <= 0 {
				break
			}
			fistFlag = false

			// 循环对比数据
			for _, row := range srcRows {
				// 对比目标的数据
				dst := dstMaps[row[uniqueI]]

				// 数据定义
				data := &Data{
					Unique:   row[uniqueI],
					Row:      row,
					fields:   make([]string, 0),
					TryTimes: 0,
				}

				// 循环对比字段
				for i := range row {
					if fields[i].Omit {
						continue
					}
					if len(row) != len(dst) {
						data.fields = append(data.fields, "all")
						break
					}
					if row[i] != dst[i] {
						data.fields = append(data.fields, fields[i].Name)
					}
				}
				// 比对完成，删除对应的map
				delete(dstMaps, row[uniqueI])

				// 如果不一样的字段存在，提交到处理的队列
				if len(data.fields) > 0 {
					data.Put()
				}
			}

			if len(dstMaps) > 0 {
				for s, row := range dstMaps {
					data := &Data{
						Unique:   s,
						Row:      row,
						fields:   []string{"-all"},
						TryTimes: 0,
					}
					data.Put()
				}
			}

			// 统计同步数量
			syncCount = syncCount + rowLen

			// 如果返回数量小于size，结束循环
			if int64(rowLen) < table.Batch {
				break
			}
		}
		log.Printf("compare done Table %s ， count %d", table.Name, syncCount)
	}
	return
}

// GetFields 获取数据库表字段
func GetFields(db *sql.DB, table config.TableInfo) (uniqueI int, fields map[int]*TableField, err error) {
	fieldList := make([]string, 0)

	// 获取字段
	fieldList, err = database.GetFieldList(db, table.Name)
	if err != nil {
		log.Println(err)
		return
	}
	// 数据库表忽略的字段，转成map
	omitFields := map[string]int{}
	for i, s := range table.Omit {
		omitFields[s] = i
	}

	fields = map[int]*TableField{}
	// 获取唯一值字段对应的下标
	for i, s := range fieldList {
		fields[i] = &TableField{
			Name:   s,
			Omit:   false,
			Unique: false,
		}
		if s == table.Unique {
			uniqueI = i
			fields[i].Unique = true
		}
		if _, ok := omitFields[s]; ok {
			fields[i].Omit = true
		}
	}
	return
}
