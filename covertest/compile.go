package covertest

import (
	"testing"

	"github.com/nspcc-dev/neo-go/cli/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/compiler"
	"github.com/nspcc-dev/neo-go/pkg/config"
	"github.com/nspcc-dev/neo-go/pkg/core/state"
	"github.com/nspcc-dev/neo-go/pkg/neotest"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/manifest"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
)

// ContractWithDebugInfo contains contract info for deployment and debug information for coverage.
type ContractWithDebugInfo struct {
	Contract  *neotest.Contract
	DebugInfo *compiler.DebugInfo
}

// CompileFile compiles a contract from the file and returns its NEF, manifest, hash and debug information.
func CompileFile(t testing.TB, sender util.Uint160, srcPath string, configPath string) *ContractWithDebugInfo {
	// nef.NewFile() cares about version a lot.
	config.Version = "neotest"

	ne, di, err := compiler.CompileWithOptions(srcPath, nil, nil)
	require.NoError(t, err)

	conf, err := smartcontract.ParseContractConfig(configPath)
	require.NoError(t, err)

	o := &compiler.Options{}
	o.Name = conf.Name
	o.ContractEvents = conf.Events
	o.ContractSupportedStandards = conf.SupportedStandards
	o.Permissions = make([]manifest.Permission, len(conf.Permissions))
	for i := range conf.Permissions {
		o.Permissions[i] = manifest.Permission(conf.Permissions[i])
	}
	o.SafeMethods = conf.SafeMethods
	o.Overloads = conf.Overloads
	o.SourceURL = conf.SourceURL
	m, err := compiler.CreateManifest(di, o)
	require.NoError(t, err)

	c := &neotest.Contract{
		Hash:     state.CreateContractHash(sender, ne.Checksum, m.Name),
		NEF:      ne,
		Manifest: m,
	}
	return &ContractWithDebugInfo{
		Contract:  c,
		DebugInfo: di,
	}
}
