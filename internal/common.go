package internal

type JsonResult[T interface{}] struct {
	Code    int    `json:"code" form:"code"`
	Error   string `json:"error" form:"error"`
	Message string `json:"message" form:"message"`
	Data    T      `json:"data" form:"data"`
}
