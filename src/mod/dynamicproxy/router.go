package dynamicproxy

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"imuslab.com/zoraxy/mod/dynamicproxy/dpcore"
	"imuslab.com/zoraxy/mod/dynamicproxy/exploits"
	"imuslab.com/zoraxy/mod/utils"
)

func (router *Router) currentRoutingState() *routingState {
	state := router.routingState.Load()
	if state != nil {
		return state
	}
	empty := &routingState{endpoints: &sync.Map{}}
	if router.routingState.CompareAndSwap(nil, empty) {
		return empty
	}
	return router.routingState.Load()
}

func (router *Router) RootEndpoint() *ProxyEndpoint {
	return router.currentRoutingState().root
}

func (router *Router) publishRootEndpoint(endpoint *ProxyEndpoint) {
	endpoint.parent = router
	for {
		current := router.currentRoutingState()
		next := &routingState{root: endpoint, endpoints: current.endpoints}
		if router.routingState.CompareAndSwap(current, next) {
			return
		}
	}
}

func (router *Router) RangeProxyEndpoints(fn func(key, value any) bool) {
	router.currentRoutingState().endpoints.Range(fn)
}

func (router *Router) LoadProxyEndpoint(key string) (*ProxyEndpoint, bool) {
	value, ok := router.currentRoutingState().endpoints.Load(key)
	if !ok {
		return nil, false
	}
	endpoint, ok := value.(*ProxyEndpoint)
	return endpoint, ok
}

// CurrentRoutingSnapshot returns root and host pointers from one atomically
// loaded routing state. Callers that serialize state therefore cannot combine
// the root from one revision with hosts from another.
func (router *Router) CurrentRoutingSnapshot() RoutingSnapshot {
	state := router.currentRoutingState()
	snapshot := RoutingSnapshot{
		Root:      state.root,
		Endpoints: make(map[string]*ProxyEndpoint),
	}
	state.endpoints.Range(func(key, value any) bool {
		lookupKey, keyOK := key.(string)
		endpoint, endpointOK := value.(*ProxyEndpoint)
		if keyOK && endpointOK {
			snapshot.Endpoints[lookupKey] = endpoint
		}
		return true
	})
	return snapshot
}

// SwapRoutingSnapshot atomically publishes a complete prepared routing
// revision. Requests already in progress retain their previously loaded state.
func (router *Router) SwapRoutingSnapshot(snapshot RoutingSnapshot) (RoutingSnapshot, error) {
	if snapshot.Root == nil {
		return RoutingSnapshot{}, errors.New("routing snapshot root is required")
	}

	endpointMap := &sync.Map{}
	for key, endpoint := range snapshot.Endpoints {
		if endpoint == nil {
			return RoutingSnapshot{}, errors.New("routing snapshot contains a nil endpoint")
		}
		lookupKey := strings.ToLower(strings.TrimSpace(key))
		if lookupKey == "" {
			return RoutingSnapshot{}, errors.New("routing snapshot contains an empty endpoint key")
		}
		endpoint.parent = router
		endpointMap.Store(lookupKey, endpoint)
	}
	snapshot.Root.parent = router

	previousState := router.routingState.Swap(&routingState{root: snapshot.Root, endpoints: endpointMap})
	previous := RoutingSnapshot{Endpoints: make(map[string]*ProxyEndpoint)}
	if previousState == nil {
		return previous, nil
	}
	previous.Root = previousState.root
	previousState.endpoints.Range(func(key, value any) bool {
		lookupKey, keyOK := key.(string)
		endpoint, endpointOK := value.(*ProxyEndpoint)
		if keyOK && endpointOK {
			previous.Endpoints[lookupKey] = endpoint
		}
		return true
	})
	return previous, nil
}

/*
	Dynamic Proxy Router Functions

	This script handle the proxy rules router spawning
	and preparation
*/

// Prepare proxy route generate a proxy handler service object for your endpoint
func (router *Router) PrepareProxyRoute(endpoint *ProxyEndpoint) (*ProxyEndpoint, error) {
	for _, thisOrigin := range endpoint.ActiveOrigins {
		//Create the proxy routing handler
		err := thisOrigin.StartProxy(endpoint.upstreamTLSServerName())
		if err != nil {
			log.Println("Unable to setup upstream " + thisOrigin.OriginIpOrDomain + ": " + err.Error())
			continue
		}
	}

	endpoint.parent = router

	//Prepare proxy routing handler for each of the virtual directories
	for _, vdir := range endpoint.VirtualDirectories {
		domain := vdir.Domain
		if len(domain) == 0 {
			//invalid vdir
			continue
		}
		if domain[len(domain)-1:] == "/" {
			domain = domain[:len(domain)-1]
		}

		//Parse the web proxy endpoint
		webProxyEndpoint := domain
		if !strings.HasPrefix("http://", domain) && !strings.HasPrefix("https://", domain) {
			//TLS is not hardcoded in proxy target domain
			if vdir.RequireTLS {
				webProxyEndpoint = "https://" + webProxyEndpoint
			} else {
				webProxyEndpoint = "http://" + webProxyEndpoint
			}
		}

		path, err := url.Parse(webProxyEndpoint)
		if err != nil {
			return nil, err
		}

		proxy := dpcore.NewDynamicProxyCore(path, vdir.MatchingPath, &dpcore.DpcoreOptions{
			IgnoreTLSVerification: vdir.SkipCertValidations,
			FlushInterval:         500 * time.Millisecond,
			UpstreamTLSServerName: endpoint.upstreamTLSServerName(),
		})
		vdir.proxy = proxy
		vdir.parent = endpoint
	}

	// Initialize the exploit detector for this endpoint
	endpoint.InitializeExploitDetector()

	return endpoint, nil
}

// Add Proxy Route to current runtime. Call to PrepareProxyRoute before adding to runtime
func (router *Router) AddProxyRouteToRuntime(endpoint *ProxyEndpoint) error {
	lookupHostname := strings.ToLower(endpoint.RootOrMatchingDomain)
	if len(endpoint.ActiveOrigins) == 0 {
		//There are no active origins. No need to check for ready
		router.currentRoutingState().endpoints.Store(lookupHostname, endpoint)
		return nil
	}
	if !router.loadBalancer.UpstreamsReady(endpoint.ActiveOrigins) {
		//This endpoint is not prepared
		return errors.New("proxy endpoint not ready. Use PrepareProxyRoute before adding to runtime")
	}
	// Push record into running subdomain endpoints
	router.currentRoutingState().endpoints.Store(lookupHostname, endpoint)
	return nil
}

// Set given Proxy Route as Root. Call to PrepareProxyRoute before adding to runtime
func (router *Router) SetProxyRouteAsRoot(endpoint *ProxyEndpoint) error {
	if !router.loadBalancer.UpstreamsReady(endpoint.ActiveOrigins) {
		//This endpoint is not prepared
		return errors.New("proxy endpoint not ready. Use PrepareProxyRoute before adding to runtime")
	}
	// Push record into running root endpoints
	router.publishRootEndpoint(endpoint)
	return nil
}

// ProxyEndpoint remove provide global access by key
func (router *Router) RemoveProxyEndpointByRootname(rootnameOrMatchingDomain string) error {
	targetEpt, err := router.LoadProxy(rootnameOrMatchingDomain)
	if err != nil {
		return err
	}

	return targetEpt.Remove()
}

// GetProxyEndpointById retrieves a proxy endpoint by its ID from the Router's ProxyEndpoints map.
// It returns the ProxyEndpoint if found, or an error if not found.
func (h *Router) GetProxyEndpointById(searchingDomain string, includeAlias bool) (*ProxyEndpoint, error) {
	var found *ProxyEndpoint
	h.RangeProxyEndpoints(func(key, value interface{}) bool {
		proxy, ok := value.(*ProxyEndpoint)
		if ok && (proxy.RootOrMatchingDomain == searchingDomain || (includeAlias && utils.StringInArray(proxy.MatchingDomainAlias, searchingDomain))) {
			found = proxy
			return false // stop iteration
		}
		return true // continue iteration
	})
	if found != nil {
		return found, nil
	}
	return nil, errors.New("proxy rule with given id not found")
}

func (h *Router) GetProxyEndpointByAlias(alias string) (*ProxyEndpoint, error) {
	var found *ProxyEndpoint
	h.RangeProxyEndpoints(func(key, value interface{}) bool {
		proxy, ok := value.(*ProxyEndpoint)
		if !ok {
			return true
		}
		//Also check for wildcard aliases that matches the alias
		for _, thisAlias := range proxy.MatchingDomainAlias {
			if ok && thisAlias == alias {
				found = proxy
				return false // stop iteration
			} else if ok && strings.HasPrefix(thisAlias, "*") {
				//Check if the alias matches a wildcard alias
				if strings.HasSuffix(alias, thisAlias[1:]) {
					found = proxy
					return false // stop iteration
				}
			}
		}
		return true // continue iteration
	})
	if found != nil {
		return found, nil
	}
	return nil, errors.New("proxy rule with given alias not found")
}

// InitializeExploitDetector initializes or updates the exploit detector for this proxy endpoint
func (pe *ProxyEndpoint) InitializeExploitDetector() {
	if pe.BlockCommonExploits || pe.BlockAICrawlers {
		exploitRespType := exploits.ExploitsRequestResponseType(pe.MitigationAction)
		pe.detector = exploits.NewExploitDetector(pe.BlockCommonExploits, pe.BlockAICrawlers, exploitRespType)
	} else {
		pe.detector = nil
	}
}
