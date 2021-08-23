package etcdcli

import (
	"context"
	"fmt"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type fakeEtcdClient struct {
	members []*etcdserverpb.Member
	opts    *FakeClientOptions
}

func (f *fakeEtcdClient) Defragment(ctx context.Context, member *etcdserverpb.Member) (*clientv3.DefragmentResponse, error) {
	// dramatic simplification
	f.opts.dbSize = f.opts.dbSizeInUse
	return nil, nil
}

func (f *fakeEtcdClient) Status(ctx context.Context, member *etcdserverpb.Member) (*clientv3.StatusResponse, error) {
	for _, status := range f.opts.status {
		if status.Header.MemberId == member.ID {
			return status, nil
		}
	}
	return nil, fmt.Errorf("status failed no match for member: %v", member)
}

func (f *fakeEtcdClient) MemberAdd(peerURL string) error {
	panic("implement me")
}

func (f *fakeEtcdClient) MemberAddAsLearner(ctx context.Context, peerURL string) error {
	panic("implement me")
}

func (f *fakeEtcdClient) MemberPromote(ctx context.Context, peerURL string) error {
	panic("implement me")
}

func (f *fakeEtcdClient) MemberList() ([]*etcdserverpb.Member, error) {
	return f.members, nil
}

func (f *fakeEtcdClient) MemberRemove(member string) error {
	panic("implement me")
}
func (f *fakeEtcdClient) MemberHealth() (memberHealth, error) {
	var healthy, unhealthy int
	var memberHealth memberHealth
	for _, member := range f.members {
		healthCheck := healthCheck{
			Member: member,
		}
		switch {
		// if WithClusterHealth is not passed we default to all healthy
		case f.opts.healthyMember == 0 && f.opts.unhealthyMember == 0:
			healthCheck.Healthy = true
			break
		case f.opts.healthyMember > 0 && healthy < f.opts.healthyMember:
			healthCheck.Healthy = true
			healthy++
			break
		case f.opts.unhealthyMember > 0 && unhealthy < f.opts.unhealthyMember:
			healthCheck.Healthy = false
			unhealthy++
			break
		}
		memberHealth = append(memberHealth, healthCheck)
	}
	return memberHealth, nil
}

func (f *fakeEtcdClient) UnhealthyMembers() ([]*etcdserverpb.Member, error) {
	if f.opts.unhealthyMember > 0 {
		// unheathy start from beginning
		return f.members[0:f.opts.unhealthyMember], nil
	}
	return []*etcdserverpb.Member{}, nil
}

func (f *fakeEtcdClient) HealthyMembers() ([]*etcdserverpb.Member, error) {
	if f.opts.healthyMember > 0 {
		// heathy start from end
		return f.members[f.opts.unhealthyMember:], nil
	}
	return []*etcdserverpb.Member{}, nil
}

func (f *fakeEtcdClient) MemberStatus(member *etcdserverpb.Member) string {
	panic("implement me")
}

func (f *fakeEtcdClient) GetMember(name string) (*etcdserverpb.Member, error) {
	for _, m := range f.members {
		if m.Name == name {
			return m, nil
		}
	}
	return nil, apierrors.NewNotFound(schema.GroupResource{Group: "etcd.operator.openshift.io", Resource: "etcdmembers"}, name)
}

func (f *fakeEtcdClient) MemberUpdatePeerURL(id uint64, peerURL []string) error {
	panic("implement me")
}

func NewFakeEtcdClient(members []*etcdserverpb.Member, opts ...FakeClientOption) (EtcdClient, error) {
	status := make([]*clientv3.StatusResponse, len(members))
	fakeEtcdClient := &fakeEtcdClient{
		members: members,
		opts: &FakeClientOptions{
			status: status,
		},
	}
	if opts != nil {
		fcOpts := newFakeClientOpts(opts...)
		switch {
		// validate WithClusterHealth
		case fcOpts.healthyMember > 0 || fcOpts.unhealthyMember > 0:
			if fcOpts.healthyMember+fcOpts.unhealthyMember != len(members) {
				return nil, fmt.Errorf("WithClusterHealth count must equal the numer of members: have %d, want %d ", fcOpts.unhealthyMember+fcOpts.healthyMember, len(members))
			}
		}
		fakeEtcdClient.opts = fcOpts
	}

	return fakeEtcdClient, nil
}

type FakeClientOptions struct {
	client          *clientv3.Client
	unhealthyMember int
	healthyMember   int
	status          []*clientv3.StatusResponse
	dbSize          int64
	dbSizeInUse     int64
}

func newFakeClientOpts(opts ...FakeClientOption) *FakeClientOptions {
	fcOpts := &FakeClientOptions{}
	fcOpts.applyFakeOpts(opts)
	fcOpts.validateFakeOpts(opts)
	return fcOpts
}

func (fo *FakeClientOptions) applyFakeOpts(opts []FakeClientOption) {
	for _, opt := range opts {
		opt(fo)
	}
}

func (fo *FakeClientOptions) validateFakeOpts(opts []FakeClientOption) {
	for _, opt := range opts {
		opt(fo)
	}
}

type FakeClientOption func(*FakeClientOptions)

type FakeMemberHealth struct {
	Healthy   int
	Unhealthy int
}

func WithFakeClusterHealth(members *FakeMemberHealth) FakeClientOption {
	return func(fo *FakeClientOptions) {
		fo.unhealthyMember = members.Unhealthy
		fo.healthyMember = members.Healthy
	}
}

func WithFakeStatus(status []*clientv3.StatusResponse) FakeClientOption {
	return func(fo *FakeClientOptions) {
		fo.status = status
	}
}
