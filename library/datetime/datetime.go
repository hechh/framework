package datetime

import (
	"time"
)

const (
	// DefaultOffset 默认时区偏移量（东八区 UTC+8）
	DefaultOffset = 8 * 60 * 60 // 8小时的秒数
)

// location 当前使用的时区（服务启动时初始化一次，之后不变）
var location = time.FixedZone("CST", DefaultOffset)

// Init 初始化datetime库，设置时区
// offset: 时区偏移量（秒），例如：
//   - 东八区: 8*60*60 = 28800
//   - 东京(东九区): 9*60*60 = 32400
//   - 纽约(西五区): -5*60*60 = -18000
//   - 伦敦(UTC+0): 0
func Init(offset int) {
	location = time.FixedZone("Custom", offset)
}

// InitWithLocation 使用time.Location初始化datetime库
func InitWithLocation(loc *time.Location) {
	if loc == nil {
		return
	}
	location = loc
}

// InitWithZoneName 使用时区名称初始化datetime库
// 例如: "Asia/Shanghai", "America/New_York", "UTC"
func InitWithZoneName(zoneName string) error {
	loc, err := time.LoadLocation(zoneName)
	if err != nil {
		return err
	}
	location = loc
	return nil
}

// GetLocation 获取当前使用的时区
func GetLocation() *time.Location {
	return location
}

// ==================== 直接获取时间戳的便捷函数 ====================

// NowUnix 获取当前时间的Unix时间戳（秒）
func NowUnix() int64 {
	return time.Now().In(location).Unix()
}

// NowUnixMilli 获取当前时间的Unix时间戳（毫秒）
func NowUnixMilli() int64 {
	return time.Now().In(location).UnixMilli()
}

// NowUnixMicro 获取当前时间的Unix时间戳（微秒）
func NowUnixMicro() int64 {
	return time.Now().In(location).UnixMicro()
}

// NowUnixNano 获取当前时间的Unix时间戳（纳秒）
func NowUnixNano() int64 {
	return time.Now().In(location).UnixNano()
}

// Now 获取当前时间的time.Time对象
func Now() time.Time {
	return time.Now().In(location)
}

// ==================== 时间戳转换函数 ====================

// ToLocation 将Unix时间戳（秒）转换为当前时区时间
func ToLocation(unix int64) time.Time {
	return time.Unix(unix, 0).In(location)
}

// ToLocationMilli 将Unix时间戳（毫秒）转换为当前时区时间
func ToLocationMilli(msec int64) time.Time {
	return time.UnixMilli(msec).In(location)
}

// ToLocationMicro 将Unix时间戳（微秒）转换为当前时区时间
func ToLocationMicro(microsec int64) time.Time {
	return time.UnixMicro(microsec).In(location)
}

// ==================== 时间戳精度转换函数 ====================

// NanoToSec 纳秒转秒
func NanoToSec(nano int64) int64 {
	return nano / 1e9
}

// NanoToMilli 纳秒转毫秒
func NanoToMilli(nano int64) int64 {
	return nano / 1e6
}

// MilliToSec 毫秒转秒
func MilliToSec(milli int64) int64 {
	return milli / 1e3
}

// MicroToSec 微秒转秒
func MicroToSec(micro int64) int64 {
	return micro / 1e6
}

// ==================== 时间运算函数（直接使用时间戳） ====================

// AddSeconds 在时间戳基础上增加秒数
func AddSeconds(unix int64, seconds int64) int64 {
	return unix + seconds
}

// AddMinutes 在时间戳基础上增加分钟数
func AddMinutes(unix int64, minutes int64) int64 {
	return unix + minutes*60
}

// AddHours 在时间戳基础上增加小时数
func AddHours(unix int64, hours int64) int64 {
	return unix + hours*3600
}

// AddDays 在时间戳基础上增加天数
func AddDays(unix int64, days int64) int64 {
	return unix + days*86400
}

// ==================== 时间差计算函数（直接使用时间戳） ====================

// DiffSeconds 计算两个时间戳之间的秒数差
func DiffSeconds(start, end int64) int64 {
	return end - start
}

// DiffMinutes 计算两个时间戳之间的分钟数差
func DiffMinutes(start, end int64) int64 {
	return (end - start) / 60
}

// DiffHours 计算两个时间戳之间的小时数差
func DiffHours(start, end int64) int64 {
	return (end - start) / 3600
}

// DiffDays 计算两个时间戳之间的天数差
func DiffDays(start, end int64) int64 {
	return (end - start) / 86400
}

// ==================== 周期计算函数（直接使用时间戳） ====================

// CountMonths 计算从起始时间戳到当前时间戳已经过了多少个月
func CountMonths(startUnix, currentUnix int64) int {
	if startUnix > currentUnix {
		return 0
	}
	start := ToLocation(startUnix)
	current := ToLocation(currentUnix)
	years := current.Year() - start.Year()
	months := int(current.Month() - start.Month())
	return years*12 + months
}

// CountCycles 计算从起始时间戳开始，已经经过了多少个完整周期（周期单位：秒）
func CountCycles(startUnix, currentUnix int64, cycleSeconds int64) int {
	if startUnix > currentUnix || cycleSeconds <= 0 {
		return 0
	}
	return int((currentUnix - startUnix) / cycleSeconds)
}

// CountCyclesRemainder 计算从起始时间戳开始，已经经过了多少个完整周期，以及剩余秒数
func CountCyclesRemainder(startUnix, currentUnix int64, cycleSeconds int64) (int64, int64) {
	if startUnix > currentUnix || cycleSeconds <= 0 {
		return 0, 0
	}
	diff := currentUnix - startUnix
	return diff / cycleSeconds, diff % cycleSeconds
}

// NextCycleTime 计算下一个周期的开始时间戳
func NextCycleTime(startUnix, currentUnix int64, cycleSeconds int64) int64 {
	if cycleSeconds <= 0 {
		return currentUnix
	}
	cycles := int64(CountCycles(startUnix, currentUnix, cycleSeconds))
	return startUnix + (cycles+1)*cycleSeconds
}

// CyclesBetween 计算两个时间戳之间有多少个完整周期
func CyclesBetween(startUnix, endUnix int64, cycleSeconds int64) int64 {
	if startUnix > endUnix || cycleSeconds <= 0 {
		return 0
	}
	return (endUnix - startUnix) / cycleSeconds
}

// ==================== 时间判断函数（直接使用时间戳） ====================

// IsBefore 判断时间戳t1是否在t2之前
func IsBefore(t1, t2 int64) bool {
	return t1 < t2
}

// IsAfter 判断时间戳t1是否在t2之后
func IsAfter(t1, t2 int64) bool {
	return t1 > t2
}

// IsEqual 判断两个时间戳是否相等
func IsEqual(t1, t2 int64) bool {
	return t1 == t2
}

// IsInSameDay 判断两个时间戳是否在同一天
func IsInSameDay(t1, t2 int64) bool {
	time1 := ToLocation(t1)
	time2 := ToLocation(t2)
	return time1.Year() == time2.Year() &&
		time1.Month() == time2.Month() &&
		time1.Day() == time2.Day()
}

// IsInSameMonth 判断两个时间戳是否在同一月
func IsInSameMonth(t1, t2 int64) bool {
	time1 := ToLocation(t1)
	time2 := ToLocation(t2)
	return time1.Year() == time2.Year() && time1.Month() == time2.Month()
}

// IsInSameYear 判断两个时间戳是否在同一年
func IsInSameYear(t1, t2 int64) bool {
	time1 := ToLocation(t1)
	time2 := ToLocation(t2)
	return time1.Year() == time2.Year()
}

// ==================== 时间边界函数 ====================

// StartOfDay 获取某天开始时间戳（00:00:00）
func StartOfDay(unix int64) int64 {
	t := time.Unix(unix, 0).In(location)
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
	return start.Unix()
}

// EndOfDay 获取某天结束时间戳（23:59:59）
func EndOfDay(unix int64) int64 {
	t := time.Unix(unix, 0).In(location)
	end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, location)
	return end.Unix()
}

// StartOfMonth 获取某月开始时间戳（1日 00:00:00）
func StartOfMonth(unix int64) int64 {
	t := time.Unix(unix, 0).In(location)
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, location)
	return start.Unix()
}

// EndOfMonth 获取某月结束时间戳（最后一天 23:59:59）
func EndOfMonth(unix int64) int64 {
	t := time.Unix(unix, 0).In(location)
	// day=0 表示上个月最后一天
	end := time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 0, location)
	return end.Unix()
}

// StartOfYear 获取某年开始时间戳（1月1日 00:00:00）
func StartOfYear(unix int64) int64 {
	t := time.Unix(unix, 0).In(location)
	start := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, location)
	return start.Unix()
}

// EndOfYear 获取某年结束时间戳（12月31日 23:59:59）
func EndOfYear(unix int64) int64 {
	t := time.Unix(unix, 0).In(location)
	end := time.Date(t.Year(), 12, 31, 23, 59, 59, 0, location)
	return end.Unix()
}

// ==================== 时间组件获取函数 ====================

// Year 获取年份
func Year(unix int64) int {
	return ToLocation(unix).Year()
}

// Month 获取月份（1-12）
func Month(unix int64) int {
	return int(ToLocation(unix).Month())
}

// Day 获取日期（1-31）
func Day(unix int64) int {
	return ToLocation(unix).Day()
}

// Hour 获取小时（0-23）
func Hour(unix int64) int {
	return ToLocation(unix).Hour()
}

// Minute 获取分钟（0-59）
func Minute(unix int64) int {
	return ToLocation(unix).Minute()
}

// Second 获取秒（0-59）
func Second(unix int64) int {
	return ToLocation(unix).Second()
}

// Weekday 获取星期几（0=周日, 1=周一, ..., 6=周六）
func Weekday(unix int64) int {
	return int(ToLocation(unix).Weekday())
}

// YearDay 获取一年中的第几天（1-366）
func YearDay(unix int64) int {
	return ToLocation(unix).YearDay()
}

// Date 获取当前日期，格式为YYYYMMDD（如 2026-05-28 → 20260528）
func Date(t time.Time) uint32 {
	return uint32(t.Year()*10000 + int(t.Month())*100 + t.Day())
}

// DaysInMonth 获取当月的天数
func DaysInMonth(unix int64) int {
	t := time.Unix(unix, 0).In(location)
	// day=0 表示上个月的最后一天
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, location).Day()
}

// IsLeapYear 判断是否为闰年
func IsLeapYear(unix int64) bool {
	year := ToLocation(unix).Year()
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

// ==================== 格式化函数 ====================

// Format 格式化时间戳
func Format(unix int64, layout string) string {
	return ToLocation(unix).Format(layout)
}

// FormatDefault 使用默认格式格式化时间戳（RFC3339）
func FormatDefault(unix int64) string {
	return ToLocation(unix).Format(time.RFC3339)
}

// Parse 解析时间字符串为秒级时间戳（使用当前时区）
func Parse(layout, value string) (int64, error) {
	t, err := time.ParseInLocation(layout, value, location)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// ParseMilli 解析时间字符串为毫秒级时间戳（使用当前时区）
func ParseMilli(layout, value string) (int64, error) {
	t, err := time.ParseInLocation(layout, value, location)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}
