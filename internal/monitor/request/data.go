package request

type HostRequest struct {
	Target string `json:"target" binding:"required"`
	Name   string `json:"name" binding:"required"`
}
