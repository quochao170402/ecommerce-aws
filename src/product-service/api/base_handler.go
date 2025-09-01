package api

type BaseRequest struct {
	PageSize  int32 `json:"pageSize"`
	PageIndex int32 `json:"pageIndex"`
}

type BaseResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
	Success bool   `json:"success"`
}

type PaginationData struct {
	BaseResponse
	Count int32 `json:"count"`
}

