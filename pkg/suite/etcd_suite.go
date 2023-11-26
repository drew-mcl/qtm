package suite

import (
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type RemoteEtcdSource struct {
	Client *clientv3.Client
	Suite  string
	Prefix string
}

func NewRemoteSuiteSource(client *clientv3.Client, suite, prefix string) *RemoteEtcdSource {
	return &RemoteEtcdSource{
		Client: client,
		Suite:  suite,
		Prefix: prefix,
	}
}

func (rs *RemoteEtcdSource) FetchSuite() (Suite, error) {
	key := fmt.Sprintf("%s/suites/%s", rs.Prefix, rs.Suite)
	resp, err := rs.Client.Get(context.Background(), key)
	if err != nil {
		return Suite{}, err
	}

	//Parse the response etcd resp into a suite
	fmt.Println("resp:", resp)
	return Suite{}, nil
}
