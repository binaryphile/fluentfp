//go:build ignore

// Package snippet is the verification harness for the consul_ingress
// showcase entry in docs/showcase.md (hashicorp/consul
// agent/structs/config_entry_gateways.ListRelatedServices rewrite).
//
// The snippet declares three package-level helpers and one method on
// *IngressGatewayConfigEntry, so the substitution is at package scope
// (no function wrapper). The types below stub the consul ingress
// vocabulary — IngressService, IngressListener (with the presumed
// `Services()` accessor noted in the showcase prose), ServiceID,
// EnterpriseMeta, NewServiceID, WildcardSpecifier — without pulling in
// the full hashicorp/consul module.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"cmp"

	"github.com/binaryphile/fluentfp/slice"
)

const WildcardSpecifier = "*"

type EnterpriseMeta struct{}

type IngressService struct {
	Name           string
	EnterpriseMeta EnterpriseMeta
}

// IngressListener carries the presumed Services() accessor noted in
// the showcase: a one-line method replacing what is a public field in
// the real consul source.
type IngressListener struct {
	servicesField []IngressService
}

func (l IngressListener) Services() []IngressService { return l.servicesField }

type ServiceID struct {
	ID             string
	EnterpriseMeta EnterpriseMeta
}

func (s ServiceID) Key() string                     { return s.ID }
func (s ServiceID) LessThan(o *EnterpriseMeta) bool { return false }

func NewServiceID(name string, meta *EnterpriseMeta) ServiceID {
	return ServiceID{ID: name}
}

type IngressGatewayConfigEntry struct {
	Listeners []IngressListener
}

// __SNIPPET__
