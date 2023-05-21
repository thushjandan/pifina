package model

type ApiErrorMessage struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type ApiRequestAppRegister struct {
	Name string `json:"name"`
}
