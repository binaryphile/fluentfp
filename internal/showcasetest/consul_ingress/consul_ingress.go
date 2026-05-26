// Package consul_ingress compile-checks the showcase entry for hashicorp/consul ingress gateway.
package consul_ingress

import (
	"cmp"

	"github.com/binaryphile/fluentfp/slice"
)

// --- stubs for the consul types ---

const WildcardSpecifier = "*"

type EnterpriseMeta struct{}

type IngressService struct {
	Name           string
	EnterpriseMeta EnterpriseMeta
}

type IngressListener struct {
	servicesField []IngressService
}

// Services is the presumed accessor (per the showcase "We presume" note).
func (l IngressListener) Services() []IngressService { return l.servicesField }

type ServiceID struct {
	ID             string
	EnterpriseMeta EnterpriseMeta
}

func (s ServiceID) Key() string                { return s.ID }
func (s ServiceID) LessThan(o *EnterpriseMeta) bool { return false }

func NewServiceID(name string, meta *EnterpriseMeta) ServiceID {
	return ServiceID{ID: name}
}

type IngressGatewayConfigEntry struct {
	Listeners []IngressListener
}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

// isExplicit returns true if the service is not a wildcard.
var isExplicit = func(s IngressService) bool { return s.Name != WildcardSpecifier }

// toServiceID builds a ServiceID from an IngressService.
var toServiceID = func(s IngressService) ServiceID {
	return NewServiceID(s.Name, &s.EnterpriseMeta)
}

// byEnterpriseThenID sorts by enterprise metadata, then by ID.
var byEnterpriseThenID = func(a, b ServiceID) int {
	switch {
	case a.LessThan(&b.EnterpriseMeta):
		return -1
	case b.LessThan(&a.EnterpriseMeta):
		return 1
	default:
		return cmp.Compare(a.ID, b.ID)
	}
}

func (e *IngressGatewayConfigEntry) ListRelatedServices() []ServiceID {
	var services slice.Mapper[IngressService] = slice.FlatMap(e.Listeners, IngressListener.Services).KeepIf(isExplicit)
	var serviceIDs slice.Mapper[ServiceID] = slice.Map(services, toServiceID)
	return slice.UniqueBy(serviceIDs, ServiceID.Key).Sort(byEnterpriseThenID)
}
