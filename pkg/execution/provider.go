package execution

import (
	"github.com/masinger/incredible/pkg/provider"
)

type initializedProvider struct {
	provider provider.Provider
	runtime  *provider.Runtime
}
