package gee

import (
	"errors"
	"net/http"
	"strings"
)

//只有刚启动服务器的时候会写入，后面就算是高并发也都是读，所以并发安全
type Router struct {
	//目前只支持GET，POST
	GetRoot *Node
	PostRoot *Node
}

func NewRouter()  *Router{
	router := Router{}
	router.GetRoot = NewNode("/")
	router.PostRoot = NewNode("/")
	return &router
}

type Node struct {
	Name string	//节点名
	Handler HandlerFunc		//处理方法
	MidHandler	[]HandlerFunc	//附加处理，在Handler外面包裹一层处理
	Children	[]*Node		//子节点
	IsLeaf bool	//是否是叶子节点
}

func (n *Node) AddChild(child *Node)  {
	n.Children = append(n.Children, child)
}

func (n *Node) GetChild(name string) *Node {
	for _, c := range n.Children {
		if c.Name == name || c.Name == "*"{
			return c
		}
	}
	return nil
}


func (n *Node) AddMidHandler(hanlder HandlerFunc)  {
	n.MidHandler = append(n.MidHandler, hanlder)
}

func (n *Node) Handle(hanlder HandlerFunc) error{
	if !n.IsLeaf {	//只有叶子节点才能有handler
		return errors.New("node is not a leaf node")
	}
	n.Handler = hanlder
	return  nil
}


func NewNode(name string)  *Node{
	return &Node{Name: name, MidHandler: make([]HandlerFunc, 0), Children: make([]*Node, 0)}
}


func (r *Router) AddNode(method string, pattern string) (*Node, error){
	var currentNode *Node
	switch method {
	case "GET":
		currentNode = r.GetRoot
	case "POST":
		currentNode = r.PostRoot
	default:
		return nil, errors.New("method error")
	}
	parts := strings.Split(pattern, "/")
	for _,part := range parts {
		if part == "" {
			continue
		}
		c := currentNode.GetChild(part)
		if c == nil {
			newNode := NewNode(part)
			currentNode.AddChild(newNode)
			c = newNode
		}
		currentNode = c
	}
	return currentNode, nil
}

func (r *Router) GetNode(method string, pattern string) (*Node, error){
	var currentNode *Node
	switch method {
	case "GET":
		currentNode = r.GetRoot
	case "POST":
		currentNode = r.PostRoot
	default:
		return nil, errors.New("method error")
	}
	parts := strings.Split(pattern, "/")
	for _,part := range parts {
		if part == "" {		// 去掉”/xx/xx/“前后的空白
			continue
		}
		c := currentNode.GetChild(part)
		if c == nil {
			return nil, errors.New("not found")
		}
		currentNode = c
	}
	return currentNode, nil
}

func (r *Router) AddHandler(method string, pattern string, handler HandlerFunc) error{
	node, err := r.AddNode(method, pattern)
	if err != nil {
		return err
	}
	node.Handler = handler
	node.IsLeaf = true
	return nil
}



func (r *Router) Handle(c *Context) {
	var currentNode *Node
	switch c.Method {
	case "GET":
		currentNode = r.GetRoot
	case "POST":
		currentNode = r.PostRoot
	default:
		c.String(http.StatusNotFound, "Method error: %s\n", c.Method)
		return
	}

	c.Log.Debug("Router:Handler=> get request :"+c.Path)
	//添加所有的midhandler
	parts := strings.Split(c.Path, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}
		child := currentNode.GetChild(part)
		if child == nil {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
			return
		}
		for _, h := range child.MidHandler {
			c.handlers = append(c.handlers, h)
		}
		currentNode = child
		if currentNode.Name == "*" {	//如果是”*“，就不用往下找了
			break
		}
	}
	//添加叶子节点的handler
	if currentNode == nil {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		return
	}
	c.handlers = append(c.handlers, currentNode.Handler)
	err := c.Req.ParseForm()
	if err != nil {
		c.Log.Error(err)
	}
	//开始执行
	c.Next()
}

