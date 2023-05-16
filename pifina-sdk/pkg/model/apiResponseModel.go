package model

type ApiErrorMessage struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
