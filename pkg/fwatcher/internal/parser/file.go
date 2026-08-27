package parser

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
)

type FileInfo struct {
	hash     hash.Hash
	value    string
	filename string
}

func NewFileInfo(filename string, body []byte) *FileInfo {
	hh := md5.New()
	var value string
	if body != nil {
		hh.Write(body)
		value = hex.EncodeToString(hh.Sum(nil))
	}
	return &FileInfo{
		hash:     hh,
		value:    value,
		filename: filename,
	}
}

func (d *FileInfo) GetValue() string {
	return d.value
}

func (d *FileInfo) Update(body []byte) {
	d.hash.Reset()
	d.hash.Write(body)
	value := hex.EncodeToString(d.hash.Sum(nil))
	if value != d.value {
		d.value = value
	}
}

func (d *FileInfo) IsChange(body []byte) bool {
	d.hash.Reset()
	d.hash.Write(body)
	value := hex.EncodeToString(d.hash.Sum(nil))
	return d.value != value
}
