package response

type HostResponse struct {
	Target string `json:"target"`
	Name   string `json:"name"`
	Id     int64  `json:"id"`
}

type HostIDResponse struct {
	Id int64 `json:"id"`
}
