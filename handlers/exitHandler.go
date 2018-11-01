package handlers

// ExitHandler handles exiting the program
type ExitHandler bool

// Close rpc call to set exit status
func (eh *ExitHandler) Close(request interface{}, responce *interface{}) error {
	*eh = true
	return nil
}
