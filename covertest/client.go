package covertest

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/core"
	"github.com/nspcc-dev/neo-go/pkg/core/block"
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/neotest"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/callflag"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/trigger"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/stretchr/testify/require"
)

// ContractInvoker is a client for a specific contract.
// It also accumulates all executed instructions for evry method invokation.
// Original ContractInvoker: https://github.com/nspcc-dev/neo-go/blob/master/pkg/neotest/client.go
type ContractInvoker struct {
	*neotest.ContractInvoker
	Methods []Method
}

// Method maps method name with executed instructions.
type Method struct {
	Name         string
	Instructions []InstrHash
}

// CommitteeInvoker creates a new ContractInvoker for the contract with hash h and a committee multisignature signer.
func CommitteeInvoker(e *neotest.Executor, h util.Uint160) *ContractInvoker {
	return &ContractInvoker{
		ContractInvoker: &neotest.ContractInvoker{
			Executor: e,
			Hash:     h,
			Signers:  []neotest.Signer{e.Committee},
		},
		Methods: nil,
	}
}

// Invoke invokes the method with the args, persists the transaction and checks the result.
// Returns transaction hash.
func (c *ContractInvoker) Invoke(t testing.TB, result interface{}, method string, args ...interface{}) util.Uint256 {
	tx := c.PrepareInvoke(t, method, args...)
	c.AddNewBlock(t, tx)
	c.CheckHalt(t, tx.Hash(), stackitem.Make(result))
	return tx.Hash()
}

// InvokeFail invokes the method with the args, persists the transaction and checks the error message.
// It returns the transaction hash.
func (c *ContractInvoker) InvokeFail(t testing.TB, message string, method string, args ...interface{}) util.Uint256 {
	tx := c.PrepareInvoke(t, method, args...)
	c.AddNewBlock(t, tx)
	c.CheckFault(t, tx.Hash(), message)
	return tx.Hash()
}

// PrepareInvoke creates a new invocation transaction.
func (c *ContractInvoker) PrepareInvoke(t testing.TB, method string, args ...interface{}) *transaction.Transaction {
	return c.NewTx(t, c.Signers, c.Hash, method, args...)
}

// NewTx creates a new transaction which invokes the contract method.
// The transaction is signed by the signers.
func (c *ContractInvoker) NewTx(t testing.TB, signers []neotest.Signer,
	hash util.Uint160, method string, args ...interface{}) *transaction.Transaction {
	tx := c.NewUnsignedTx(t, hash, method, args...)
	return c.SignTx(t, tx, -1, method, signers...)
}

// SignTx signs a transaction using the provided signers.
func (c *ContractInvoker) SignTx(t testing.TB, tx *transaction.Transaction, sysFee int64, method string, signers ...neotest.Signer) *transaction.Transaction {
	for _, acc := range signers {
		tx.Signers = append(tx.Signers, transaction.Signer{
			Account: acc.ScriptHash(),
			Scopes:  transaction.Global,
		})
	}
	neotest.AddNetworkFee(c.Chain, tx, signers...)
	c.AddSystemFee(c.Chain, tx, sysFee, method)

	for _, acc := range signers {
		require.NoError(t, acc.SignTx(c.Chain.GetConfig().Magic, tx))
	}
	return tx
}

// AddSystemFee adds system fee to the transaction. If negative value specified,
// then system fee is defined by test invocation.
func (c *ContractInvoker) AddSystemFee(bc *core.Blockchain, tx *transaction.Transaction, sysFee int64, method string) {
	if sysFee >= 0 {
		tx.SystemFee = sysFee
		return
	}
	ops, v, _ := TestInvoke(bc, tx) // ignore error to support failing transactions
	tx.SystemFee = v.GasConsumed()
	c.Methods = append(c.Methods, Method{
		Name:         method,
		Instructions: ops,
	})
}

// TestInvoke creates a test VM with a dummy block and executes a transaction in it.
func TestInvoke(bc *core.Blockchain, tx *transaction.Transaction) ([]InstrHash, *vm.VM, error) {
	lastBlock, err := bc.GetBlock(bc.GetHeaderHash(bc.BlockHeight()))
	if err != nil {
		return nil, nil, err
	}
	b := &block.Block{
		Header: block.Header{
			Index:     bc.BlockHeight() + 1,
			Timestamp: lastBlock.Timestamp + 1,
		},
	}

	// `GetTestVM` as well as `Run` can use a transaction hash which will set a cached value.
	// This is unwanted behavior, so we explicitly copy the transaction to perform execution.
	ttx := *tx
	ic, _ := bc.GetTestVM(trigger.Application, &ttx, b)

	defer ic.Finalize()

	ic.VM.LoadWithFlags(tx.Script, callflag.All)
	ops, err := Run(ic.VM)
	return ops, ic.VM, err
}
