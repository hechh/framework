package parser

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
)

type FileInfo struct {
	hash      hash.Hash
	hashValue string
	filename  string
	body      []byte
}

func NewFileInfo(filename string, body []byte) *FileInfo {
	hh := md5.New()
	var value string
	if body != nil {
		hh.Write(body)
		value = hex.EncodeToString(hh.Sum(nil))
	}
	return &FileInfo{
		hash:      hh,
		hashValue: value,
		filename:  filename,
		body:      body,
	}
}

func (d *FileInfo) GetValue() string {
	return d.hashValue
}

func (d *FileInfo) GetText() []byte {
	return d.body
}

func (d *FileInfo) Update(body []byte) bool {
	d.hash.Reset()
	d.hash.Write(body)
	value := hex.EncodeToString(d.hash.Sum(nil))
	if value != d.hashValue {
		d.hashValue = value
		d.body = body
		return true
	}
	return false
}

func (d *FileInfo) IsChange(body []byte) bool {
	d.hash.Reset()
	d.hash.Write(body)
	value := hex.EncodeToString(d.hash.Sum(nil))
	return d.hashValue != value
}
