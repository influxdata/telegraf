package influxdb_v3

type APIError struct {
	Err        error
	StatusCode int
	Retryable  bool
}

func (e APIError) Error() string {
	return e.Err.Error()
}

func (e APIError) Unwrap() error {
	return e.Err
}
