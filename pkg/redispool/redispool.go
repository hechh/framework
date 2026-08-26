package redispool

// Config 数据库分片配置
type Config struct {
	DbName   string            `yaml:"dbname,omitempty"`   // 数据库名称
	Db       uint32            `yaml:"db,omitempty"`       // 数据库编号
	User     string            `yaml:"user,omitempty"`     // 数据库用户名
	Password string            `yaml:"password,omitempty"` // 数据库密码
	Ip       string            `yaml:"ip,omitempty"`       // 数据库IP地址
	Port     uint32            `yaml:"port,omitempty"`     // 数据库端口
	Prefix   string            `yaml:"prefix,omitempty"`   // key 前缀
	Slaves   map[int32]*Config `yaml:"slaves,omitempty"`   // 从库配置列表
}

type RedisPool struct {
	newFunc func() IClient     // new函数
	pools   map[string]IClient // 全局数据库连接池
}

func NewRedisPool[T IClient](f func() T) *RedisPool {
	return &RedisPool{
		newFunc: func() IClient { return f() },
		pools:   make(map[string]IClient),
	}
}

func (d *RedisPool) Init(cfgs map[string]*Config) error {
	// 初始化全局数据库
	for dbname, dbCfg := range cfgs {
		cli := d.newFunc()
		if err := cli.Init(dbCfg); err != nil {
			d.Close()
			return err
		}
		d.pools[dbname] = cli
	}
	return nil
}

func (d *RedisPool) Close() {
	for _, cli := range d.pools {
		cli.Close()
	}
}

func (d *RedisPool) Get(name string) IClient {
	return d.pools[name]
}
