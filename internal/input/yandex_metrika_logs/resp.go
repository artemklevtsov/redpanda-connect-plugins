package logs

type apiEvalLogRequestResponse struct {
	Result evalLogRequest `json:"log_request_evaluation"`
}

type evalLogRequest struct {
	IsPossible bool `json:"possible"`
	MaxDays    int  `json:"max_possible_day_quantity"`
}

type apiLogRequestResponse struct {
	Request apiLogRequest `json:"log_request"`
}

type apiLogRequest struct {
	apiQuery
	RequestID uint64 `json:"request_id"`
	CounterID uint64 `json:"counter_id"`
	Status    string `json:"status"`
	Size      uint64 `json:"size"`
	Parts     []struct {
		Num  int    `json:"part_number"`
		Size uint64 `json:"size"`
	}
}
