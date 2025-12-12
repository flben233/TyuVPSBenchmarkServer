package model

type Host struct {
	Target   string `json:"target"`
	Name     string `json:"name"`
	Id       int64  `json:"id"`
	Uploader string `json:"uploader"`
}
