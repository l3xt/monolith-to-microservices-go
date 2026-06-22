package apperror

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

func NewRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &RetryableError{Err: err}
}
