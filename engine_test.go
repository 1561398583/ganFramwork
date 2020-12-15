package gee

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestEngine_Run(t *testing.T) {
	engine := Default()
	engine.AddViewPath("E:\\go_project\\streamingMedia\\video_server\\views\\")
	engine.GET("/index/hello", hello)
	engine.GET("/user/*", userController)
	routerNode, err := engine.router.GetNode("GET", "/user")
	if err != nil {
		panic(err)
	}
	routerNode.AddMidHandler(userP)
	engine.Use(catchPanic)
	engine.Use(recordTime)
	engine.GET("/static/pictrue/*", getPictrue)
	engine.Run("localhost:7000")
}

func getPictrue(c *Context)  {
	path := c.Path
	parts := strings.Split(path, "/")
	picName := parts[len(parts)-1]
	absPath := "E:\\pic\\" + picName
	f, err := os.Open(absPath)
	if err != nil {
		panic(err)
	}
	finfo, err := f.Stat()
	if err != nil {
		panic(err)
	}
	size := finfo.Size()
	data := make([]byte, size)
	n, err := f.Read(data)
	if err != nil {
		panic(err)
	}
	if int64(n) != int64(size) {
		panic("size != n")
	}
	c.SetHeader("Content-Type", "image/jpeg")
	c.SetHeader("Content-Length", strconv.FormatInt(size, 10))
	c.Data(200, data)
}

func hello(c *Context)  {
	name := c.GetForm("name")
	m := make(map[string]string)
	m["name"] = name
	c.HTML(200, "hello", m)
}

func userController(c *Context)  {
	path := c.Path
	parts := strings.Split(path, "/")
	m := make(map[string]string)
	m["name"] = parts[len(parts)-1]
	m["work"] = "golang"
	c.JSON(200, m)
}

func userP(c *Context)  {
	fmt.Println("begin user")
	c.Next()
	fmt.Println("finish user")
}

func recordTime(c *Context)  {
	start := time.Now()
	c.Next()
	spend := time.Since(start)
	fmt.Println("spend ", spend)
}


