package fileutil

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/imports"
	"gopkg.in/yaml.v3"
)

// fileLocks 按目标文件路径维护互斥锁，保证对同一文件的并发写是串行的。
// 在 Windows 上 os.Rename 覆盖已存在目标时若被其他写句柄占用会失败，
// 因此同一文件的原子写必须互斥。
var (
	fileLocks   = make(map[string]*sync.Mutex)
	fileLocksMu sync.Mutex
)

// lockFile 获取指定路径对应的互斥锁。
func lockFile(path string) *sync.Mutex {
	fileLocksMu.Lock()
	defer fileLocksMu.Unlock()
	mu, ok := fileLocks[path]
	if !ok {
		mu = &sync.Mutex{}
		fileLocks[path] = mu
	}
	return mu
}

// EnsureDir 判断目录是否存在，如果不存在则创建目录
func EnsureDir(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(dir, os.FileMode(0o755))
}

// CreateFile 创建文件
func CreateFile(fileName string, flag int) (fb *os.File, err error) {
	// 判断路径是否存在
	pp := filepath.Dir(fileName)
	if err := os.MkdirAll(pp, os.FileMode(0o755)); err != nil {
		return nil, err
	}
	// 创建文件
	if fb, err = os.OpenFile(fileName, flag, os.FileMode(0o644)); err != nil {
		return nil, err
	}
	return
}

// IsSameFile 文件是否相同
func IsSameFile(fb *os.File, name string) bool {
	st2, _ := os.Stat(name)
	st1, _ := fb.Stat()
	return os.SameFile(st1, st2)
}

// Save 保存文件（自动创建目录，.go 文件自动格式化 import）
func Save(fileName string, buf []byte) error {
	dir := filepath.Dir(fileName)
	if err := os.MkdirAll(dir, os.FileMode(0o755)); err != nil {
		return err
	}
	if ext := filepath.Ext(fileName); ext == ".go" {
		var err error
		if buf, err = imports.Process(fileName, buf, nil); err != nil {
			return err
		}
	}
	return os.WriteFile(fileName, buf, os.FileMode(0o644))
}

// AtomicSave 原子保存文件（先写临时文件再 rename）。
// 避免文件监听方在写盘过程中（截断后/写完前）读到半写内容。
//
// 并发安全：对同一目标文件的并发写加互斥锁串行化，并使用唯一临时文件名
// （os.CreateTemp），避免多个 goroutine 同时写同一个固定 .tmp 文件导致
// 互相覆盖 / rename 失败（Windows 上尤其明显）。
func AtomicSave(fileName string, buf []byte) error {
	mu := lockFile(fileName)
	mu.Lock()
	defer mu.Unlock()

	dir := filepath.Dir(fileName)
	if err := os.MkdirAll(dir, os.FileMode(0o755)); err != nil {
		return err
	}

	// 在目标同目录创建唯一临时文件，保证 rename 原子替换
	tmp, err := os.CreateTemp(dir, filepath.Base(fileName)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // 失败时清理残留临时文件

	if _, err := tmp.Write(buf); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(os.FileMode(0o644)); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, fileName)
}

// ParseFiles 解析go文件
func ParseFiles(v ast.Visitor, files ...string) error {
	fset := token.NewFileSet()
	for _, filename := range files {
		// 解析语法树
		fs, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		// 遍历语法树
		ast.Walk(v, fs)
		/*
			buf := bytes.NewBuffer(nil)
			ast.Fprint(buf, fset, fs, nil)
			os.WriteFile(fmt.Sprintf("%s.ini", filename), buf.Bytes(), 0644)
		*/
	}
	return nil
}

func Glob(pattern string, isRecursive bool) ([]string, error) {
	if !isRecursive {
		return filepath.Glob(pattern)
	}

	// 递归模式：
	// 1. 提取目录和文件名模式
	dir, filePattern := filepath.Split(pattern)
	if dir == "" {
		dir = "."
	}
	// 如果 dir 为空或 "."，则从当前目录开始

	// 2. 遍历目录树
	var matches []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// 遇到权限错误等，可选择跳过或返回错误
			return nil // 跳过该文件/目录继续
		}
		if d.IsDir() {
			return nil // 继续遍历子目录
		}
		// 检查文件名是否匹配模式（仅匹配文件名，不包含路径）
		ok, err := filepath.Match(filePattern, d.Name())
		if err != nil {
			return err // 模式语法错误
		}
		if ok {
			matches = append(matches, path)
		}
		return nil
	})
	return matches, err
}

func SearchFile(filename string, depth int) string {
	if _, err := os.Stat(filename); err == nil {
		abs, _ := filepath.Abs(filename)
		return abs
	}
	if depth <= 0 {
		return ""
	}
	return SearchFile(filepath.Join("..", filename), depth-1)
}

func LoadYaml(filename string, val any) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(content, val)
}

func Map2Yaml(data any, cfg any, names ...string) error {
	var val any = data
	for i, name := range names {
		vv, ok := val.(map[string]any)
		if !ok {
			return fmt.Errorf("字段类型错误, path:%s", strings.Join(names[:i+1], "/"))
		}
		if val, ok = vv[name]; !ok {
			return fmt.Errorf("字段不存在, path:%s", strings.Join(names[:i+1], "/"))
		}
	}
	bytes, err := yaml.Marshal(val)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, cfg)
}
