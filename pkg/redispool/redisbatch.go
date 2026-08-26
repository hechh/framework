package redispool

import (
	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/library/tplutil"
)

func unmarshal(values []*Value, results []any) error {
	for i, val := range results {
		var err error
		switch vv := val.(type) {
		case string:
			err = values[i].UnmarshalVT(safe.StringToBytes(vv))
		case []byte:
			err = values[i].UnmarshalVT(vv)
		default:
			err = values[i].UnmarshalVT(nil)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func Load(args ...*Value) error {
	type data struct {
		typeData uint32
		key      string
		cli      IClient
		values   []*Value
		args     []string
	}
	datas := map[tplutil.Tuple2[uint32, string]]*data{}
	for _, item := range args {
		cli, typeData := item.Client(), item.Type()
		key, field := item.Key(), item.Field()
		kk := tplutil.T2(cli.UniqueId(), tplutil.Or(typeData == HASH, key, field))
		vv, ok := datas[kk]
		if !ok {
			vv = &data{cli: cli, typeData: typeData, key: key}
			datas[kk] = vv
		}
		vv.values = append(vv.values, item)
		vv.args = append(vv.args, tplutil.Or(typeData == STRING, key, field))
	}
	for _, vv := range datas {
		var results []any
		var err error
		if vv.typeData == STRING {
			results, err = vv.cli.MGet(vv.args...)
		} else {
			results, err = vv.cli.HMGet(vv.key, vv.args...)
		}
		if err != nil {
			return err
		}
		if err := unmarshal(vv.values, results); err != nil {
			return err
		}
	}
	return nil
}

func Save(args ...*Value) error {
	type data struct {
		client   IClient
		typeData uint32
		key      string
		args     []any
	}
	datas := map[tplutil.Tuple2[uint32, string]]*data{}
	for _, item := range args {
		if !item.IsChanged() {
			continue
		}

		buff, err := item.MarshalVT()
		if err != nil {
			return err
		}

		cli, typeData := item.Client(), item.Type()
		key, field := item.Key(), item.Field()

		kk := tplutil.T2(cli.UniqueId(), tplutil.Or(typeData == HASH, key, field))
		vv, ok := datas[kk]
		if !ok {
			vv = &data{typeData: typeData, key: key, client: cli}
			datas[kk] = vv
		}

		kval := tplutil.Or(typeData == STRING, key, field)
		vv.args = append(vv.args, kval, safe.BytesToString(buff))
	}
	for _, vv := range datas {
		var err error
		if vv.typeData == HASH {
			err = vv.client.HMSet(vv.key, vv.args...)
		} else {
			err = vv.client.MSet(vv.args...)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveDirectly(args ...*Value) error {
	type data struct {
		client   IClient
		typeData uint32
		key      string
		args     []any
	}
	datas := map[tplutil.Tuple2[uint32, string]]*data{}
	for _, item := range args {
		buff, err := item.MarshalVT()
		if err != nil {
			return err
		}

		cli, typeData := item.Client(), item.Type()
		key, field := item.Key(), item.Field()

		kk := tplutil.T2(cli.UniqueId(), tplutil.Or(typeData == HASH, key, field))
		vv, ok := datas[kk]
		if !ok {
			vv = &data{typeData: typeData, key: key, client: cli}
			datas[kk] = vv
		}

		kval := tplutil.Or(typeData == STRING, key, field)
		vv.args = append(vv.args, kval, safe.BytesToString(buff))
	}
	for _, vv := range datas {
		var err error
		if vv.typeData == HASH {
			err = vv.client.HMSet(vv.key, vv.args...)
		} else {
			err = vv.client.MSet(vv.args...)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
