package gee

import (
	"runtime"
	"strconv"
)

type WebError struct {
	err error
	StackInfo string	//调用栈信息
}

type WE struct {
	msg string
}

func (we *WE) Error() string{
	return we.msg
}



func (we *WebError) Error() string {
	return we.err.Error()
}


func (we *WebError) GetStackInfo() string {
	return we.StackInfo
}


func NewError(msg string) *WebError {
	err := WE{msg: msg}
	stackInfo := getStack()
	return &WebError{err: &err, StackInfo: stackInfo}
}

//获取位置
func getLine() string {
	_, fileName, line, ok := runtime.Caller(2)
	if !ok {
		panic("can not get file line")
	}
	return "[" + fileName + " : " + strconv.FormatInt(int64(line), 10) + "]"
}

//获取调用栈信息
func getStack() string {
	s := ""
	for i := 0; ;i++ {
		_, fileName, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		s = s + fileName + " " + strconv.FormatInt(int64(line), 10) + "\n"
	}
	return s
}

func Wrap(e error) error {
	//如果是WebError,直接返回;否则把error封装成WebError返回
	if we, ok := e.(*WebError); ok{
		return we
	}else {
		return &WebError{err: e, StackInfo: getStack()}
	}
}

//获取error类型
func Cause(e error) error {
	if we, ok := e.(*WebError); ok{
		//如果是*WebError，就把error提取出来
		return we.err
	}
	return e
}




