package suite

type EtcdDataSource struct {
	SuiteName string
	Endpoint  string
}

func (e EtcdDataSource) ReadData() (Suite, error) {
	// Implement reading from etcd
	return Suite{}, nil
}
