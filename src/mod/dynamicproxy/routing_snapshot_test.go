package dynamicproxy

import (
	"fmt"
	"sync"
	"testing"
)

func snapshotTestRouter(t *testing.T) *Router {
	t.Helper()
	router, err := NewDynamicProxy(RouterOption{})
	if err != nil {
		t.Fatal(err)
	}
	return router
}

func TestSwapRoutingSnapshotPublishesCompleteRevision(t *testing.T) {
	router := snapshotTestRouter(t)
	oldRoot := &ProxyEndpoint{RootOrMatchingDomain: "/", DefaultSiteValue: "old"}
	oldHost := &ProxyEndpoint{RootOrMatchingDomain: "old.example"}
	if _, err := router.SwapRoutingSnapshot(RoutingSnapshot{
		Root: oldRoot, Endpoints: map[string]*ProxyEndpoint{"OLD.EXAMPLE": oldHost},
	}); err != nil {
		t.Fatal(err)
	}

	inFlight := router.currentRoutingState()
	newRoot := &ProxyEndpoint{RootOrMatchingDomain: "/", DefaultSiteValue: "new"}
	newHost := &ProxyEndpoint{RootOrMatchingDomain: "new.example"}
	previous, err := router.SwapRoutingSnapshot(RoutingSnapshot{
		Root: newRoot, Endpoints: map[string]*ProxyEndpoint{"NEW.EXAMPLE": newHost},
	})
	if err != nil {
		t.Fatal(err)
	}

	if router.RootEndpoint() != newRoot {
		t.Fatal("new root was not published")
	}
	if got, ok := router.LoadProxyEndpoint("new.example"); !ok || got != newHost {
		t.Fatalf("new endpoint = %v, %v", got, ok)
	}
	if _, ok := router.LoadProxyEndpoint("old.example"); ok {
		t.Fatal("old endpoint remained in active revision")
	}
	if previous.Root != oldRoot || previous.Endpoints["old.example"] != oldHost {
		t.Fatalf("previous snapshot = %+v", previous)
	}
	if got := router.getProxyEndpointFromHostnameState(inFlight, "old.example"); got != oldHost {
		t.Fatal("in-flight request state did not retain its old revision")
	}
}

func TestSwapRoutingSnapshotRejectsInvalidCandidateWithoutChangingRuntime(t *testing.T) {
	router := snapshotTestRouter(t)
	root := &ProxyEndpoint{RootOrMatchingDomain: "/"}
	if _, err := router.SwapRoutingSnapshot(RoutingSnapshot{Root: root}); err != nil {
		t.Fatal(err)
	}
	if _, err := router.SwapRoutingSnapshot(RoutingSnapshot{}); err == nil {
		t.Fatal("snapshot without root was accepted")
	}
	if router.RootEndpoint() != root {
		t.Fatal("invalid candidate changed the active root")
	}
}

func TestRoutingSnapshotConcurrentReadersAndSwaps(t *testing.T) {
	router := snapshotTestRouter(t)
	if _, err := router.SwapRoutingSnapshot(RoutingSnapshot{Root: &ProxyEndpoint{RootOrMatchingDomain: "/"}}); err != nil {
		t.Fatal(err)
	}

	var workers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := 0; i < 500; i++ {
				_ = router.RootEndpoint()
				_, _ = router.LoadProxyEndpoint("host.example")
				router.RangeProxyEndpoints(func(_, _ any) bool { return true })
			}
		}()
	}
	for i := 0; i < 100; i++ {
		host := &ProxyEndpoint{RootOrMatchingDomain: fmt.Sprintf("host-%d.example", i)}
		if _, err := router.SwapRoutingSnapshot(RoutingSnapshot{
			Root:      &ProxyEndpoint{RootOrMatchingDomain: "/"},
			Endpoints: map[string]*ProxyEndpoint{"host.example": host},
		}); err != nil {
			t.Fatal(err)
		}
	}
	workers.Wait()
}
