package integration

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	// id
	Data T `json:"data"`
}
