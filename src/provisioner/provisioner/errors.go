package provisioner

type TimeoutError struct{}

func (t *TimeoutError) Error() string {
	return "timeout error"
}
