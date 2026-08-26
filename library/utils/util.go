package utils

import (
	"fmt"
	"hash/crc32"
	"reflect"
	"strings"

	"github.com/hechh/framework/library/safe"
)

func GetCrc32(name string) uint32 {
	return crc32.ChecksumIEEE(safe.StringToBytes(name))
}

func GetTraceId(uid uint64, createTime int64) uint32 {
	return GetCrc32(fmt.Sprintf("%d-%d", uid, createTime))
}

// ParseActorName 解析actor名称
func ParseName(rr reflect.Type) string {
	name := rr.String()
	if index := strings.Index(name, "."); index > -1 {
		name = name[index+1:]
	}
	return name
}
