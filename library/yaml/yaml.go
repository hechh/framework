package yaml

import (
	"os"

	"gopkg.in/yaml.v3"
)

type SlaveConfig struct {
	DbName   string `yaml:"dbname"`
	Db       int    `yaml:"db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
}

type DbConfig struct {
	DbName   string                 `yaml:"dbname"`
	Db       int32                  `yaml:"db"`
	Prefix   string                 `yaml:"prefix"`
	User     string                 `yaml:"user"`
	Password string                 `yaml:"password"`
	Host     string                 `yaml:"host"`
	Slave    map[int32]*SlaveConfig `yaml:"slave"`
}

type EtcdConfig struct {
	Topic     string   `yaml:"topic"`
	Endpoints []string `yaml:"endpoints"`
}

type NatsConfig struct {
	Topic     string `yaml:"topic"`
	Endpoints string `yaml:"endpoints"`
}

// 游戏配置
type TableConfig struct {
	IsRemote  bool     `yaml:"is_remote"`
	Ext       string   `yaml:"ext"`
	Path      string   `yaml:"path"`
	Endpoints []string `yaml:"endpoints"`
}

type NodeConfig struct {
	RouterExpire   int64  `yaml:"router_expire"`
	RegisterExpire int64  `yaml:"register_expire"`
	LogLevel       string `yaml:"log_level"`
	LogPath        string `yaml:"log_path"`
	Ip             string `yaml:"ip"`
	Port           int    `yaml:"port"`
	HttpPort       int    `yaml:"http_port"`
}

func Load(filename string, cfg any) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(content, cfg)
}

/*
type Config struct {
	Mysql   map[int32]*DbConfig   `yaml:"mysql"`
	Redis   map[int32]*DbConfig   `yaml:"redis"`
	Mongodb map[int32]*DbConfig   `yaml:"mongodb"`
	Etcd    *EtcdConfig           `yaml:"etcd"`
	Nats    *NatsConfig           `yaml:"nats"`
	Gate    map[int32]*NodeConfig `yaml:"gate"`
	Client  map[int32]*NodeConfig `yaml:"client"`
	Room    map[int32]*NodeConfig `yaml:"room"`
	Match   map[int32]*NodeConfig `yaml:"match"`
	Db      map[int32]*NodeConfig `yaml:"db"`
	Build   map[int32]*NodeConfig `yaml:"build"`
	Game    map[int32]*NodeConfig `yaml:"game"`
	Gm      map[int32]*NodeConfig `yaml:"gm"`
}
*/
