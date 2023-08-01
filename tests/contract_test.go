package contract

import (
	"path"
	"testing"

	"git.frostfs.info/TrueCloudLab/contract-coverage-primer/covertest"
	"github.com/nspcc-dev/neo-go/pkg/neotest"
	"github.com/nspcc-dev/neo-go/pkg/neotest/chain"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
)

const ctrPath = "../impulse"

// Key for tests
var (
	validKey   = []byte{1, 2, 3, 4, 5}
	invalidKey = []byte{1, 2, 3}
)

func newExecutor(t *testing.T) *neotest.Executor {
	bc, acc := chain.NewSingle(t)
	return neotest.NewExecutor(t, bc, acc, acc)
}

func TestContract(t *testing.T) {
	e := newExecutor(t)
	ctrDI := covertest.CompileFile(t, e.CommitteeHash, ctrPath, path.Join(ctrPath, "config.yml"))
	ctr := ctrDI.Contract
	e.DeployContract(t, ctr, nil)
	inv := e.CommitteeInvoker(ctr.Hash)

	// test get without put
	inv.InvokeFail(t, "Cannot get number", "getNumber", validKey)

	// test put-get with valid key
	inv.Invoke(t, stackitem.Null{}, "putNumber", validKey, 42)
	inv.Invoke(t, 42, "getNumber", validKey)

	// test invalid key
	inv.InvokeFail(t, "Invalid key size", "putNumber", invalidKey, 42)
	inv.InvokeFail(t, "Invalid key size", "getNumber", invalidKey)
}
