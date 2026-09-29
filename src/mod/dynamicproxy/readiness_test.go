package dynamicproxy

import (
	"net"
	"testing"

	"imuslab.com/zoraxy/mod/info/logger"
)

func testReadinessRouter(t *testing.T, port int) *Router {
	t.Helper()
	testLogger, err := logger.NewFmtLogger()
	if err != nil {
		t.Fatal(err)
	}
	router, err := NewDynamicProxy(RouterOption{
		Port:   port,
		UseTls: false,
		Logger: testLogger,
	})
	if err != nil {
		t.Fatal(err)
	}
	router.publishRootEndpoint(&ProxyEndpoint{})
	return router
}

func TestStartProxyServiceIsReadyOnlyAfterBind(t *testing.T) {
	router := testReadinessRouter(t, 0)
	if err := router.StartProxyService(); err != nil {
		t.Fatal(err)
	}
	if !router.IsReady() {
		t.Fatal("router is not ready after binding its primary listener")
	}
	if err := router.StopProxyService(); err != nil {
		t.Fatal(err)
	}
	if router.IsReady() {
		t.Fatal("router remained ready after shutdown")
	}
}

func TestStartProxyServiceDoesNotBecomeReadyWhenBindFails(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	port := occupied.Addr().(*net.TCPAddr).Port

	router := testReadinessRouter(t, port)
	if err := router.StartProxyService(); err == nil {
		t.Fatal("start succeeded despite an occupied primary port")
	}
	if router.IsReady() {
		t.Fatal("router reported ready after its primary bind failed")
	}
}
