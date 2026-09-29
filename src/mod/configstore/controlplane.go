package configstore

import (
	"context"
	"errors"
	"net/http"
)

const (
	RevisionAPIPath   = "/api/cluster/config"
	NodeStatusAPIPath = "/api/cluster/nodes"
)

// ManagementRouter is implemented by http.ServeMux and Zoraxy's authenticated
// router. Production callers must pass the authenticated management router.
type ManagementRouter interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) error
}

// ControlPlane groups the authoritative revision API, convergence status API
// and data-node follower bootstrap without owning database credentials.
type ControlPlane struct {
	store      Store
	statuses   NodeStatusStore
	revisions  *HTTPHandler
	nodeStatus *NodeStatusHandler
}

func NewControlPlane(store Store, statuses NodeStatusStore, validate Validator, actor ActorResolver) (*ControlPlane, error) {
	if store == nil {
		return nil, errors.New("configuration store is required")
	}
	if statuses == nil {
		return nil, errors.New("node status store is required")
	}
	return &ControlPlane{
		store:      store,
		statuses:   statuses,
		revisions:  NewHTTPHandler(store, validate, actor),
		nodeStatus: NewNodeStatusHandler(statuses),
	}, nil
}

func (c *ControlPlane) RegisterManagementAPI(router ManagementRouter) error {
	if router == nil {
		return errors.New("management router is required")
	}
	if err := router.HandleFunc(RevisionAPIPath, c.revisions.ServeHTTP); err != nil {
		return err
	}
	return router.HandleFunc(NodeStatusAPIPath, c.nodeStatus.ServeHTTP)
}

func (c *ControlPlane) RunDataNode(ctx context.Context, activator Activator, node NodeStatus) error {
	if activator == nil {
		return errors.New("runtime activator is required")
	}
	return FollowNode(ctx, c.store, activator, c.statuses, node)
}
