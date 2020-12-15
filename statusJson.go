package gee

type RespJson struct {
	Status string	//"ok" or "error"
	Code int
	Data string
}

const (
	OK = "ok"
	ERROR = "error"
)

//code
const (
	NoError = 1 << iota
	ParamError 	//参数错误
)
