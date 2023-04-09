
package discovery_test

import (
	"github.com/golang/mock/gomock"
	"github.com/mu-box/golang-discovery"
	"github.com/mu-box/golang-discovery/mock"
	"github.com/mu-box/golang-discovery/mock-closer"
	"net"
	"testing"
	"time"
)

func TestDiscovery(test *testing.T) {
	ctrl := gomock.NewController(test)
	defer ctrl.Finish()

	generator := mock_golang_discovery.NewMockGenerator(ctrl)
	closer := mock_io.NewMockCloser(ctrl)

	lo := loopBack(test)
	discover, err := discovery.NewDiscovery(lo, "testing", time.Second)
	if err != nil {
		test.Log(err)
		test.FailNow()
	}

	discover.Handle("thing", generator)

	generator.EXPECT().New("127.0.0.2").Return(closer)
	closer.EXPECT().Close().Return(nil)

	go func() {
		<-time.After(time.Second)

		// we can't use the multicast to test on the same machine
		conn, err := net.Dial("udp", "127.0.0.1:5432")
		if err != nil {
			test.Log(err)
			test.FailNow()
		}
		defer conn.Close()

		_, err = conn.Write([]byte("testing,thing@127.0.0.2"))
		if err != nil {
			test.Log(err)
			test.FailNow()
		}

		<-time.After(time.Second / 2)
		_, err = conn.Write([]byte("testing,thing@127.0.0.2"))
		if err != nil {
			test.Log(err)
			test.FailNow()
		}
	}()

	discover.Add("test", "what")

	go discover.Loop(time.Second)
	<-time.After(time.Second * 5)
	discover.Remove("test")
	discover.Unhandle("thing")
	discover.Close()
}

func loopBack(test *testing.T) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		test.Log(err)
		test.FailNow()
	}
	for _, iface := range interfaces {
		addresses, err := iface.Addrs()
		if err != nil {
			test.Log(err)
			test.FailNow()
		}
		for _, addr := range addresses {
			if addr.String() == "127.0.0.1/8" {
				return iface.Name
			}
		}
	}
	test.Log("unable to find loop back interface")
	test.FailNow()
	return ""
}
