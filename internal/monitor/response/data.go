package response

type StatisticsResponse struct {
	Name     string    `json:"name"`
	Uploader string    `json:"uploader"`
	History  []float32 `json:"history"`
}
