package dns

import (
	"context"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	infrav1 "github.com/spectrocloud/cluster-api-provider-maas/api/v1alpha3"
	mockclientset "github.com/spectrocloud/cluster-api-provider-maas/pkg/maas/client/mock"
	"github.com/spectrocloud/cluster-api-provider-maas/pkg/maas/scope"
	"github.com/spectrocloud/maas-client-go/maasclient"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/klogr"
	"net"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	"testing"
)

func TestDNS(t *testing.T) {
	log := klogr.New()
	cluster := &v1alpha3.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: "a",
		},
	}
	maasCluster := &infrav1.MaasCluster{
		Spec: infrav1.MaasClusterSpec{
			DNSDomain: "b.com",
		},
	}

	t.Run("reconcile dns", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ctrl := gomock.NewController(t)
		mockClientSetInterface := mockclientset.NewMockClientSetInterface(ctrl)
		mockDNSResources := mockclientset.NewMockDNSResources(ctrl)
		mockDNSResourceBuilder := mockclientset.NewMockDNSResourceBuilder(ctrl)
		s := &Service{
			scope: &scope.ClusterScope{
				Logger:      log,
				Cluster:     cluster,
				MaasCluster: maasCluster,
			},
			maasClient: mockClientSetInterface,
		}
		mockClientSetInterface.EXPECT().DNSResources().Return(mockDNSResources)
		mockDNSResources.EXPECT().List(context.Background(), gomock.Any()).Return(nil, nil)
		mockClientSetInterface.EXPECT().DNSResources().Return(mockDNSResources)
		mockDNSResources.EXPECT().Builder().Return(mockDNSResourceBuilder)
		mockDNSResourceBuilder.EXPECT().WithFQDN("a.b.com").Return(mockDNSResourceBuilder)
		mockDNSResourceBuilder.EXPECT().WithAddressTTL("10").Return(mockDNSResourceBuilder)
		mockDNSResourceBuilder.EXPECT().WithIPAddresses(nil).Return(mockDNSResourceBuilder)
		mockDNSResourceBuilder.EXPECT().Create(context.Background())
		err := s.ReconcileDNS()

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(s.scope.GetDNSName()).To(BeEquivalentTo("a.b.com"))
	})

	t.Run("update dns attachment", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ctrl := gomock.NewController(t)
		mockClientSetInterface := mockclientset.NewMockClientSetInterface(ctrl)
		mockDNSResources := mockclientset.NewMockDNSResources(ctrl)
		mockDNSResource := mockclientset.NewMockDNSResource(ctrl)
		mockDNSResourceModifier := mockclientset.NewMockDNSResourceModifier(ctrl)
		s := &Service{
			scope: &scope.ClusterScope{
				Logger:      log,
				Cluster:     cluster,
				MaasCluster: maasCluster,
			},
			maasClient: mockClientSetInterface,
		}

		mockClientSetInterface.EXPECT().DNSResources().Return(mockDNSResources)
		mockDNSResources.EXPECT().List(context.Background(), gomock.Any()).Return([]maasclient.DNSResource{mockDNSResource}, nil)
		mockDNSResource.EXPECT().Modifier().Return(mockDNSResourceModifier)
		mockDNSResourceModifier.EXPECT().SetIPAddresses([]string{"1.1.1.1", "8.8.8.8"}).Return(mockDNSResourceModifier)
		mockDNSResourceModifier.EXPECT().Modify(context.Background()).Return(mockDNSResource, nil)

		err := s.UpdateDNSAttachments([]string{"1.1.1.1", "8.8.8.8"})

		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("machine is registered", func(t *testing.T) {
		g := NewGomegaWithT(t)
		ctrl := gomock.NewController(t)
		mockClientSetInterface := mockclientset.NewMockClientSetInterface(ctrl)
		mockDNSResources := mockclientset.NewMockDNSResources(ctrl)
		mockDNSResource := mockclientset.NewMockDNSResource(ctrl)
		mockIPAddress := mockclientset.NewMockIPAddress(ctrl)
		s := &Service{
			scope: &scope.ClusterScope{
				Logger:      log,
				Cluster:     cluster,
				MaasCluster: maasCluster,
			},
			maasClient: mockClientSetInterface,
		}
		mockClientSetInterface.EXPECT().DNSResources().Return(mockDNSResources)
		mockDNSResources.EXPECT().List(context.Background(), gomock.Any()).Return([]maasclient.DNSResource{mockDNSResource}, nil)
		mockDNSResource.EXPECT().IPAddresses().Return([]maasclient.IPAddress{mockIPAddress})
		mockIPAddress.EXPECT().IP().Return(net.ParseIP("1.1.1.1"))
		mockIPAddress.EXPECT().IP().Return(net.ParseIP("8.8.8.8"))

		res, err := s.MachineIsRegisteredWithAPIServerDNS(&infrav1.Machine{
			Addresses: []v1alpha3.MachineAddress{
				{
					Type:    v1alpha3.MachineInternalIP,
					Address: "1.1.1.1",
				},
				{
					Type:    v1alpha3.MachineInternalIP,
					Address: "8.8.8.8",
				},
			},
		})

		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(res).To(BeTrue())
	})
}