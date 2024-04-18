package covertest

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/compiler"
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
type ContractInvoker struct {
	*neotest.ContractInvoker
	Methods []Method
}

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
	c.Methods = append(c.Methods, Method{
		Name:         method,
		Instructions: nil,
	})
	tx := c.PrepareInvoke(t, method, args...)
	c.AddNewBlock(t, tx)
	c.CheckHalt(t, tx.Hash(), stackitem.Make(result))
	return tx.Hash()
}

// InvokeFail invokes the method with the args, persists the transaction and checks the error message.
// It returns the transaction hash.
func (c *ContractInvoker) InvokeFail(t testing.TB, message string, method string, args ...interface{}) util.Uint256 {
	c.Methods = append(c.Methods, Method{
		Name:         method,
		Instructions: nil,
	})
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
	return c.SignTx(t, tx, -1, signers...)
}

// SignTx signs a transaction using the provided signers.
func (c *ContractInvoker) SignTx(t testing.TB, tx *transaction.Transaction, sysFee int64, signers ...neotest.Signer) *transaction.Transaction {
	for _, acc := range signers {
		tx.Signers = append(tx.Signers, transaction.Signer{
			Account: acc.ScriptHash(),
			Scopes:  transaction.Global,
		})
	}
	neotest.AddNetworkFee(c.Chain, tx, signers...)
	c.AddSystemFee(c.Chain, tx, sysFee)

	for _, acc := range signers {
		require.NoError(t, acc.SignTx(c.Chain.GetConfig().Magic, tx))
	}
	return tx
}

// AddSystemFee adds system fee to the transaction. If negative value specified,
// then system fee is defined by test invocation.
func (c *ContractInvoker) AddSystemFee(bc *core.Blockchain, tx *transaction.Transaction, sysFee int64) {
	if sysFee >= 0 {
		tx.SystemFee = sysFee
		return
	}
	ops, v, _ := TestInvoke(bc, tx) // ignore error to support failing transactions
	tx.SystemFee = v.GasConsumed()
	c.Methods[len(c.Methods)-1].Instructions = make([]InstrHash, len(ops))
	copy(c.Methods[len(c.Methods)-1].Instructions, ops)
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

type coverline struct {
	Doc       string
	Opcode    int
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	WTFnumber int
	IsCovered bool
}

// MakeCoverage generates an output file with coverage info in correct format
func (c *ContractInvoker) MakeCoverage(t testing.TB, ctrdi *ContractWithDebugInfo, ctrPath string, fileName string) {
	docs := getDocuments(t, ctrdi.DebugInfo.Documents, ctrPath)
	cov := getSeqPoints(t, ctrdi.DebugInfo, docs)
	for _, iMethod := range c.Methods {
		countInstructions(cov, iMethod.Instructions)
	}
	printToFile(t, cov, fileName)
}

func getDocuments(t testing.TB, docs []string, substr string) []int {
	res := make([]int, 0, len(docs))

	for i, cDoc := range docs {
		if strings.Contains(cDoc, substr) {
			res = append(res, i)
		}
	}
	if len(res) == 0 {
		t.Log("Cannot get document\n")
		t.FailNow()
	}
	return res
}

func getSeqPoints(t testing.TB, di *compiler.DebugInfo, docs []int) []coverline {
	res := make([]coverline, 0, 10)

	for _, method := range di.Methods {
		maxLine := method.Range.End
		for _, seqPoint := range method.SeqPoints {
			if isValidDocument(seqPoint.Document, docs) && seqPoint.Opcode < int(maxLine) {
				res = append(res, coverline{
					Doc:       di.Documents[seqPoint.Document],
					Opcode:    seqPoint.Opcode,
					StartLine: seqPoint.StartLine,
					StartCol:  seqPoint.StartCol,
					EndLine:   seqPoint.EndLine,
					EndCol:    seqPoint.EndCol,
					WTFnumber: 1,
					IsCovered: false,
				})
			}
		}
	}
	return res
}

func isValidDocument(iDocToCheck int, docs []int) bool {
	for _, iDoc := range docs {
		if iDoc == iDocToCheck {
			return true
		}
	}
	return false
}

func countInstructions(cov []coverline, codes []InstrHash) {
	for i := 0; i < len(cov); i++ {
		for _, code := range codes {
			if code.Offset == cov[i].Opcode {
				cov[i].IsCovered = true
				//cov[i].WTFnumber++
				//break
			}
		}
	}
}

func printToFile(t testing.TB, cov []coverline, name string) {
	f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	require.NoError(t, err)

	defer f.Close()

	fi, err := os.Stat(name)
	require.NoError(t, err)
	firstToWrite := ""
	if fi.Size() == 0 {
		firstToWrite = "mode: set\n"
	}

	_, err = f.WriteString(firstToWrite)
	require.NoError(t, err)

	for _, info := range cov {
		covered := 0
		if info.IsCovered {
			covered++
		}
		line := fmt.Sprintf("%s:%d.%d,%d.%d %d %d\n", info.Doc, info.StartLine, info.StartCol, info.EndLine, info.EndCol, info.WTFnumber, covered)
		_, err = f.WriteString(line)
		require.NoError(t, err)
	}
}
