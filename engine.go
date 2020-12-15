package gee

import (
	"fmt"
	"net/http"
	"runtime"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(c *Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	router *Router
	log *PdLog
	handlers []HandlerFunc
	ViewPath string
}

// New is the constructor of gee.Engine
func New(logPath, viewPath string) *Engine {
	log := NewLog(logPath, "", LstdFlags, Debug)
	return &Engine{router: NewRouter(), log: log, handlers: make([]HandlerFunc, 0), ViewPath: viewPath}
}

func Default() *Engine {
	log := NewLog("E:\\log\\video_server\\", "", LstdFlags, Debug)
	engine := &Engine{router: NewRouter(), log: log, handlers: make([]HandlerFunc, 0), ViewPath: "E:\\go_project\\streamingMedia\\video_server\\views\\"}
	engine.Use(catchPanic)	//默认的panic处理
	return engine
}

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	err := engine.router.AddHandler(method, pattern, handler)
	if err != nil {
		panic(err)
	}
}

// GET defines the method to add GET request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// Run defines the method to start a http server
func (engine *Engine) Use(handler HandlerFunc)  {
	engine.handlers = append(engine.handlers, handler)
}

func (engine *Engine) AddViewPath(path string)  {
	engine.ViewPath = path
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	c.index = -1
	c.Log = engine.log
	c.engine = engine
	c.handlers = make([]HandlerFunc, 0)
	//给context添加全局handler
	for _, h := range engine.handlers {
		c.handlers = append(c.handlers, h)
	}
	engine.router.Handle(c)
}

func catchPanic(c *Context)  {
	defer func() {
		r := recover()
		if r != nil {
			errInfo := fmt.Sprintln(r)
			for i := 2; ; i++ {	//从第二个栈开始，前面2个栈多余
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				errInfo += fmt.Sprintln(file, line)
			}
			//记录日志
			c.Log.Error(errInfo)
		}
	}()
	c.Next()
}

