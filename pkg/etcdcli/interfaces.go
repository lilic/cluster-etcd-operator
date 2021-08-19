package etcdcli

import (
	"context"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	EtcdMemberStatusAvailable  = "EtcdMemberAvailable"
	EtcdMemberStatusNotStarted = "EtcdMemberNotStarted"
	EtcdMemberStatusUnhealthy  = "EtcdMemberUnhealthy"
	EtcdMemberStatusUnknown    = "EtcdMemberUnknown"
)

type EtcdClient interface {
	Defragment
	MemberAdder
	MemberHealth
	MemberLister
	MemberRemover
	MemberAdderAsLearner
	UnhealthyMemberLister
	MemberStatusChecker
	Status

	GetMember(name string) (*etcdserverpb.Member, error)
	MemberUpdatePeerURL(id uint64, peerURL []string) error
}

type Defragment interface {
	Defragment(ctx context.Context, member *etcdserverpb.Member) (*clientv3.DefragmentResponse, error)
}

type Status interface {
	Status(ctx context.Context, member *etcdserverpb.Member) (*clientv3.StatusResponse, error)
}

type MemberAdder interface {
	MemberAdd(peerURL string) error
}

type MemberAdderAsLearner interface {
	MemberAddAsLearner(ctx context.Context, peerURL string) (memberHealth, error)
}

type MemberHealth interface {
	MemberHealth() (memberHealth, error)
}

type MemberRemover interface {
	MemberRemove(member string) error
}

type MemberLister interface {
	MemberList() ([]*etcdserverpb.Member, error)
}

type UnhealthyMemberLister interface {
	UnhealthyMembers() ([]*etcdserverpb.Member, error)
}

type MemberStatusChecker interface {
	MemberStatus(member *etcdserverpb.Member) string
}
