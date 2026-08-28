package factory

import (
	"strings"

	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/dbpool"
	"github.com/hechh/framework/pkg/dbpool/adapter/mysql"
	"github.com/hechh/framework/pkg/dbpool/adapter/postgre"
	"github.com/hechh/framework/pkg/dbpool/adapter/starrocks"
)

func NewClient(val string) dbpool.IClient {
	valType := packet.DbType_MYSQL
	if vv, ok := packet.DbType_value[strings.ToUpper(val)]; ok {
		valType = packet.DbType(vv)
	}
	switch valType {
	case packet.DbType_POSTGRESQL:
		return postgre.NewClient(val)
	case packet.DbType_STARROCKS:
		return starrocks.NewClient(val)
	default:
		return mysql.NewClient(val)
	}
}
