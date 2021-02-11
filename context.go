package gee

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type H map[string]interface{}

//html 模板缓存
var templateCache = make(map[string]*template.Template)

//context的作用就是把所有需要的参数打个包。并且提供了很多的常用的处理函数，让用户不用写太多重复的代码。
type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	Params map[string]string
	// response info
	StatusCode int
	// middleware
	handlers []HandlerFunc
	index    int

	Log 	*PdLog
	engine *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index: -1,
	}
}

func (c *Context) GetForm(key string) string {
	return c.Req.Form.Get(key)
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	_, err := c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
	if err != nil {
		panic(err)
	}
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	data, err := json.Marshal(obj)
	if err != nil {
		c.Fail(500, "system error")
		panic(err)
	}
	_, err = c.Writer.Write(data)
	if err != nil {
		c.Fail(500, "system error")
		panic(err)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	_, err := c.Writer.Write(data)
	if err != nil {
		c.Log.Error(Wrap(err))
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) HTML(code int, htmlName string, data interface{}) {
	t, ok := templateCache[htmlName]
	if !ok {
		tem, err := template.ParseFiles(c.engine.ViewPath + htmlName + ".html")
		if err != nil {
			c.Log.Error(err)
			http.Error(c.Writer, err.Error(), 404)
			return
		}
		t = tem
	}
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Status(code)
	err := t.Execute(c.Writer, data)
	if err != nil {
		c.Fail(500, "system error")
		panic(err)
	}
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) Template(key string) string {
	value, _ := c.Params[key]
	return value
}

//发送静态文件
func (c *Context) File(absPath string)  {
	f, err := os.Open(absPath)
	if err != nil {
		panic(err)
	}
	//文件打开成功就得关闭
	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()
	finfo, err := f.Stat()
	if err != nil {
		panic(err)
	}
	//获取文件类型
	a := strings.Split(absPath, ".")
	fileType := a[len(a)-1]
	size := finfo.Size()
	data := make([]byte, size)
	n, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	if int64(n) != int64(size) {
		panic("size != n")
	}
	//设置头信息
	switch fileType {
	case "jpg":
		c.SetHeader("Content-Type", "image/jpeg")
	case "png":
		c.SetHeader("Content-Type", "image/png")
	case "mp4":
		c.SetHeader("Content-Type", "video/mp4")
	case "flv":
		c.SetHeader("Content-Type", "video/flv")
	}
	c.SetHeader("Content-Length", strconv.FormatInt(size, 10))
	c.Data(200, data)
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)	//c.Next就直接到达末尾，就不用执行后面的handler了
	c.JSON(code, H{"message": err})
}

func (c *Context) Next()  {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}