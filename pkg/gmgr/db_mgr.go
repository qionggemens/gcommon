/**
 * Create Time:2022/12/8
 * User: luchao
 * Email: lcmusic1994@gmail.com
 */

package gmgr

import (
	"fmt"
	"github.com/qionggemens/gcommon/pkg/glog"
	"github.com/qionggemens/gcommon/pkg/nacos"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

func NewDB(dbName string) (*gorm.DB, error) {
	mysqlUrl := nacos.GetString(fmt.Sprintf("db.%s.mysql_url", dbName), "")
	// 最大和最小尽量相差很小，要不然会引起无限新开连接，直至端口号用完
	maxOpenConns := nacos.GetInt(fmt.Sprintf("db.%s.max_open_conns", dbName), 5)
	minIdleConns := nacos.GetInt(fmt.Sprintf("db.%s.min_idle_conns", dbName), 5)
	maxLifeTime := nacos.GetInt64(fmt.Sprintf("db.%s.max_life_time", dbName), 60)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       mysqlUrl, // DSN data source name
		DefaultStringSize:         64,       // string 类型字段的默认长度
		DisableDatetimePrecision:  true,     // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,     // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,     // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,    // 根据当前 MySQL 版本自动配置

	}), &gorm.Config{
		PrepareStmt: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 设置禁用表名复数形式属性为 true，`User` 的表名将是 `user`
		}})
	if nil != err {
		glog.Errorf("InitDB fail - msg:%v, url:%s", err, mysqlUrl)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		glog.Errorf("InitDB fail - get sql.DB: %v", err)
		return nil, err
	}
	// SetMaxIdleConns 空闲连接上限；SetMaxOpenConns 打开连接上限（原先两者参数写反）
	sqlDB.SetMaxIdleConns(minIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifeTime) * time.Second)
	glog.Infoln("InitDB db success")
	return db, nil
}
