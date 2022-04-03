package trace

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

func (f *file) SetFileMaxNum(num int) {
	f.maxNum = num
}
func (f *file) SetFileMaxSize(size int) {
	f.maxSize = size
}
func (f *file) SetFilePath(path string) {
	f.path = path
}

type file struct {
	path         string
	name         string
	fType        string
	header       []byte
	user         fmt.Stringer
	file         *os.File
	buf          *bufio.Writer
	size         int
	maxSize      int
	num          int
	maxNum       int
	bufSize      int
	historyFiles []string
	bufWait      bytes.Buffer
	error        error
}

func (f *file) fileNumCheck() error {
	if f.num < f.maxNum {
		return nil
	}
	file:=f.historyFiles[0]
	f.historyFiles = f.historyFiles[1:]
	return errors.WithStack(os.Remove(file))
}

func (f *file)filePathCheck() error {
	return os.MkdirAll(f.path, os.ModePerm)
}

func (f *file) createBufFile() error {
	if err:=f.fileNumCheck();err!=nil{
		return err
	}
	if f.file != nil && f.file.Close() != nil {
		return errors.New("close file error")
	}
	f.file, f.error = os.OpenFile(f.name,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0666,
	)
	if f.error != nil {
		return errors.WithStack(f.error)
	}
	f.buf = bufio.NewWriterSize(f.file, f.bufSize)
	_,err:=f.buf.Write(f.header)
	return errors.WithStack(err)
}

func (f *file) preStore(p []byte) (int,error) {
	if f.size > 1000 {
		return 0,nil
	}
	nn, err := f.bufWait.Write(p)
	if err != nil {
		return  nn,errors.WithStack(err)
	}
	f.size += nn
	return  nn,nil
}

func (f *file) newFile() error {
	if f.name != "" {
		return nil
	}
	f.name = f.path + "/" + f.user.String() + "."+strconv.Itoa(f.num)+"."+f.fType
	if err:=f.filePathCheck();err!=nil{
		return errors.WithStack(err)
	}
	if err := f.createBufFile(); err != nil {
		return  errors.WithStack(err)
	}
	if _, err := f.buf.Write(f.bufWait.Bytes()); err != nil {
		return  errors.WithStack(err)
	}
	f.bufWait.Reset()
	return nil
}

func (f *file) Write(p []byte) (n int, err error) {
	if f.error != nil {
		return 0, errors.WithStack(f.error)
	}
	if n,err:=f.preStore(p);n!=0||err!=nil{
		return n,errors.WithStack(err)
	}
	if err:=f.newFile();err != nil {
		return 0, errors.WithStack(err)
	}
	nn, err := f.buf.Write(p)
	if err != nil {
		return nn, errors.WithStack(err)
	}
	f.size += nn
	if f.size >= f.maxSize {
		f.num++
		f.size = 0
		f.historyFiles=append(f.historyFiles, f.name)
		f.name = ""
		if err := f.buf.Flush(); err != nil {
			return 0, errors.WithStack(err)
		}
	}
	return nn, nil
}

func (f *file) Close() error {
	if len(f.bufWait.Bytes()) > 0 {
		if err := f.newFile(); err != nil {
			return errors.WithStack(err)
		}
		if _, err := f.buf.Write(f.bufWait.Bytes()); err != nil {
			return errors.WithStack(err)
		}
	}
	if f.buf ==nil {
		return nil
	}
	if err := f.buf.Flush(); err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(f.file.Close())
}
