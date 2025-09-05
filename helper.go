package tbdb

func (i *Instance) validateClient() error {
	if i.client == nil {
		return ErrClientNil
	}
	return nil
}
