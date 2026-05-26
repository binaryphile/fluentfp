// Package annotation compile-checks the showcase entry for k8s ttl_controller.
package annotation

import (
	"strconv"

	"github.com/binaryphile/fluentfp/option"
)

// --- stub for the k8s Node type ---

type Node struct {
	Annotations map[string]string
}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

// tryAtoi parses s as an integer, returning an ok option on success or not-ok on failure.
var tryAtoi = func(s string) option.Int {
	n, err := strconv.Atoi(s)
	return option.New(n, err == nil)
}

func getIntFromAnnotation(node *Node, annotationKey string) (int, bool) {
	annotation := option.Lookup(node.Annotations, annotationKey)
	return option.FlatMap(annotation, tryAtoi).Get()
}
