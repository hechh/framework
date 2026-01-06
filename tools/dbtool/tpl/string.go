package tpl

const RedisStringTpl = `
/*
* 本代码由pbtool工具生成，请勿手动修改
*/

package {{.Pkg}}

import (
	"fmt"
	"mypoker/common/pb"
	"mypoker/framework/library/uerror"
	"mypoker/framework/repository/redis"
	
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cast"
)

const (
	DBNAME = {{.DbName}}
)

func GetKey({{GetArgs .Keys}}) string {
	return fmt.Sprintf("{{.Format}}" {{if .Keys}},{{end}} {{GetValues .Keys}})
}

func Get({{GetArgs .Keys}}) (*pb.{{.Name}}, error) {
	// 获取redis连接
	client := redis.Get(DBNAME)
	if client == nil {
		return nil, uerror.New(-1,"redis数据库不存在: %s", DBNAME)
	}
	
	// 加载数据
	str, err := client.Get(GetKey({{GetValues .Keys}}))
	if err != nil {
		return nil, err
	}
	
	// 解析数据
	if len(str) > 0 {
		data := &pb.{{.Name}}{}
		err = proto.Unmarshal([]byte(str), data)
		return data, err
	}
	return nil, nil	
}

func Set({{if .Keys}} {{GetArgs .Keys}}, {{end}} val *pb.{{.Name}}) error {
	// 获取redis连接
	client := redis.Get(DBNAME)
	if client == nil {
		return uerror.New(-1,"redis数据库不存在: %s", DBNAME)
	}

	// 编码数据
	buf, err := proto.Marshal(val)
	if err != nil {
		return err
	}
	
	// 存储数据
	return client.Set(GetKey({{GetValues .Keys}}), buf)
}

func Del({{GetArgs .Keys}}) error {
	// 获取redis连接
	client := redis.Get(DBNAME)
	if client == nil {
		return uerror.New(-1,"redis数据库不存在: %s", DBNAME)
	}

	// 删除数据
	return client.Del(GetKey({{GetValues .Keys}}))
}

func MSet(vals map[string]*pb.{{.Name}}) error {
	// 获取redis连接
	client := redis.Get(DBNAME)
	if client == nil {
		return uerror.New(-1,"redis数据库不存在: %s", DBNAME)
	}

	// 解析数据
	args := []any{}
	for key, val := range vals {
		args = append(args, key)	
		if buf, err := proto.Marshal(val); err != nil {
			return err
		} else {
			args = append(args, buf)
		}
	}

	// 批量储存数据
	return client.MSet(args...)
}

func MGet(keys ...string) (map[string]*pb.{{.Name}}, error) {
	// 获取redis连接
	client := redis.Get(DBNAME)
	if client == nil {
		return nil, uerror.New(-1,"redis数据库不存在: %s", DBNAME)
	}

	// 批量加载数据
	values, err := client.MGet(keys...)
	if err != nil {
		return nil, err	
	}

	// 解析数据
	rets := map[string]*pb.{{.Name}}{}
	for i, key := range keys {
		value := values[i]
		if value == nil {
			continue	
		}
		item := &pb.{{.Name}}{}	 
		err := proto.Unmarshal([]byte(cast.ToString(value)), item)
		if err == nil {
			rets[key] = item
		}
	}
	return rets, nil
}





`
