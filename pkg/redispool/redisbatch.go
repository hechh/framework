package redispool

import (
	"github.com/bytedance/sonic"
	"github.com/hechh/framework/define"
	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/pkg/mlog"
)

func Load(args ...*Value) error {
	type data struct {
		*Value
		values []*Value
		args   []string
	}
	datas := map[string]*data{}
	for _, item := range args {
		cliType, dataType, key, field := item.GetGroupId(), item.GetDataType(), item.GetKey(), item.GetField()
		vv, ok := datas[cliType]
		if !ok {
			vv = &data{Value: item}
			datas[cliType] = vv
		}
		vv.values = append(vv.values, item)
		vv.args = append(vv.args, tplutil.Or(dataType == STRING, key, field))
	}
	for _, vv := range datas {
		var results []any
		var err error
		if vv.GetDataType() == STRING {
			results, err = vv.GetClient().MGet(vv.args...)
		} else {
			results, err = vv.GetClient().HMGet(vv.GetKey(), vv.args...)
		}
		if err != nil {
			return err
		}
		for i, val := range vv.values {
			if err = val.Unmarshal(results[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func Save(args ...*Value) error {
	type data struct {
		*Value
		args   []any
		values []*Value
	}
	datas := map[string]*data{}
	for _, item := range args {
		if !item.IsChanged() {
			continue
		}
		buff, err := item.Marshal()
		if err != nil {
			return err
		}
		kk, dataType, key, field := item.GetGroupId(), item.GetDataType(), item.GetKey(), item.GetField()
		vv, ok := datas[kk]
		if !ok {
			vv = &data{Value: item}
			datas[kk] = vv
		}
		vv.args = append(vv.args, tplutil.Or(dataType == STRING, key, field), safe.BytesToString(buff))
		vv.values = append(vv.values, item)
	}
	var err error
	for _, vv := range datas {
		var reterr error
		if vv.GetDataType() == STRING {
			reterr = vv.GetClient().MSet(vv.args...)
		} else {
			reterr = vv.GetClient().HMSet(vv.GetKey(), vv.args...)
		}
		if reterr != nil {
			err = reterr
		}
	}
	if err != nil {
		for _, item := range args {
			if !item.IsChanged() {
				continue
			}
			buff, _ := sonic.Marshal(item.Get())
			mlog.Errorf("保存Redis数据失败 key:%s, field:%s, msg:%s, error:%s", item.GetKey(), item.GetField(), safe.BytesToString(buff), err)
		}
	}
	return err
}

func Remove(args ...*Value) error {
	type data struct {
		*Value
		args []string
	}
	datas := map[string]*data{}
	for _, item := range args {
		kk, dataType, key, field := item.GetGroupId(), item.GetDataType(), item.GetKey(), item.GetField()
		vv, ok := datas[kk]
		if !ok {
			vv = &data{Value: item}
			datas[kk] = vv
		}
		vv.args = append(vv.args, tplutil.Or(dataType == STRING, key, field))
	}
	var err error
	for _, vv := range datas {
		var reterr error
		if vv.GetDataType() == STRING {
			_, reterr = vv.GetClient().Del(vv.args...)
		} else {
			_, reterr = vv.GetClient().HDel(vv.GetKey(), vv.args...)
		}
		if reterr != nil {
			err = reterr
		}
	}
	return err
}

func SaveByCtx(ctx define.IContext) error {
	vals := tplutil.Map3Values[*Value](ctx.GetAllCache())
	if err := Save(vals...); err != nil {
		return err
	}
	ctx.Refresh()
	return nil
}

func SaveDirectlyByCtx(ctx define.IContext) error {
	vals := tplutil.Map3Values[*Value](ctx.GetAllCache())
	if err := SaveDirectly(vals...); err != nil {
		return err
	}
	ctx.Refresh()
	return nil
}

func SaveDirectly(args ...*Value) error {
	type data struct {
		*Value
		args   []any
		values []*Value
	}
	datas := map[string]*data{}
	for _, item := range args {
		buff, err := item.Marshal()
		if err != nil {
			return err
		}
		kk, dataType, key, field := item.GetGroupId(), item.GetDataType(), item.GetKey(), item.GetField()
		vv, ok := datas[kk]
		if !ok {
			vv = &data{Value: item}
			datas[kk] = vv
		}
		vv.args = append(vv.args, tplutil.Or(dataType == STRING, key, field), safe.BytesToString(buff))
		vv.values = append(vv.values, item)
	}
	var err error
	for _, vv := range datas {
		var reterr error
		if vv.GetDataType() == STRING {
			reterr = vv.GetClient().MSet(vv.args...)
		} else {
			reterr = vv.GetClient().HMSet(vv.GetKey(), vv.args...)
		}
		if reterr != nil {
			err = reterr
		}
	}
	if err != nil {
		for _, item := range args {
			if !item.IsChanged() {
				continue
			}
			buff, _ := sonic.Marshal(item.Get())
			mlog.Errorf("保存Redis数据失败 key:%s, field:%s, msg:%s, error:%s", item.GetKey(), item.GetField(), safe.BytesToString(buff), err)
		}
	}
	return err
}
