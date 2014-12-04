package nio

type WriterFunc func([]byte) (int, error)

func (w WriterFunc) Write(p []byte) (int, error) {
	return w(p)
}

func (w WriterFunc) Close() error {
	return nil
}
