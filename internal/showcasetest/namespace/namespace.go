// Package namespace compile-checks the showcase entry for kubernetes/client-go.
package namespace

import (
	"os"
	"strings"

	"github.com/binaryphile/fluentfp/option"
)

// --- stub for the inClusterClientConfig type ---

type inClusterClientConfig struct{}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

const saPath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// trim converts bytes to a trimmed string.
var trim = func(b []byte) string { return strings.TrimSpace(string(b)) }

// readSANamespace reads the namespace from the service account token file.
var readSANamespace = func() option.String {
	return option.NonErr(os.ReadFile(saPath)).ToString(trim).FlatMap(option.NonEmpty)
}

func (config *inClusterClientConfig) Namespace() (string, bool, error) {
	ns := option.Env("POD_NAMESPACE").OrElse(readSANamespace).Or("default")
	return ns, false, nil
}
