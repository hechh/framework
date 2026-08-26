package registry

var (
	tables = make(map[string][]any)
)

func Register(dbname string, tabs ...any) {
	tables[dbname] = append(tables[dbname], tabs...)
}

func GetTables(dbname string) []any {
	return tables[dbname]
}
