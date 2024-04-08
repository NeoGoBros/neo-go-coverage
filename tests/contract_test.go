package contract

import (
	"errors"
	"fmt"
	"math/rand"
	"path"
	"testing"

	"git.frostfs.info/TrueCloudLab/contract-coverage-primer/covertest"
	"github.com/nspcc-dev/neo-go/pkg/compiler"
	"github.com/nspcc-dev/neo-go/pkg/neotest"
	"github.com/nspcc-dev/neo-go/pkg/neotest/chain"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/callflag"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/trigger"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/stretchr/testify/require"
)

const ctrPath = "../impulse"

// keys for tests
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
	inv := covertest.CommitteeInvoker(e, ctr.Hash)

	// test get without put
	inv.InvokeFail(t, "Cannot get number", "getNumber", validKey)

	// test put-get with valid key
	inv.Invoke(t, stackitem.Null{}, "putNumber", validKey, 42)
	inv.Invoke(t, 42, "getNumber", validKey)

	// test invalid key
	inv.InvokeFail(t, "Invalid key size", "putNumber", invalidKey, 42)
	inv.InvokeFail(t, "Invalid key size", "getNumber", invalidKey)
	// spew.Dump(inv.Methods)
	inv.MakeCoverage(t, ctrDI, "contract.go", "c.out")
}

func TestRun(t *testing.T) {
    tests := []struct {
        function string
    }{
        {"PutNumber"},
        {"GetNumber"},
    }
    for _, test := range tests {
        test := test
        t.Run(test.function, func(t *testing.T) {
			e := newExecutor(t)
			ctrDI := covertest.CompileFile(t, e.CommitteeHash, ctrPath, path.Join(ctrPath, "config.yml"))
			e.DeployContract(t, ctrDI.Contract, nil)

			startOffsetPutNumber, err := getStartOffset(ctrDI.DebugInfo, test.function)
			require.NoError(t, err)

			hasResult, err := hasResult(ctrDI.DebugInfo, test.function)
			require.NoError(t, err)

			someNum := getNumToPut()

			// set up a VM for covertest.Run()
			covertestRunVM := setUpVMForPut(t, e, ctrDI.Contract, hasResult, startOffsetPutNumber, someNum, invalidKey)
			res, covErr := covertest.Run(covertestRunVM)
			t.Log("Printing collected instructions:")
			dumpCoveredInstructions(res, test.function)
			t.Log("covertest.Run() returned an error: ", covErr)

			// set up a VM for vm.Run()
			origRunVM := setUpVMForPut(t, e, ctrDI.Contract, hasResult, startOffsetPutNumber, someNum, invalidKey)
			runerr := origRunVM.Run()
			t.Log("vm.Run() returned an error: ", covErr)

			// check if errors are the same
			require.Equal(t, runerr.Error(), covErr.Error())

			// check if the number of elements on the stack is the same
			require.Equal(t, origRunVM.Estack().Len(), covertestRunVM.Estack().Len())
        })
    }
}

func dumpCoveredInstructions(instructions []covertest.InstrHash, functionName string) {
	fmt.Printf("=== Covered instructions of the '%s' function\n", functionName)
	println("INDEX\t OPCODE")
	for _, instruction := range instructions {
		println(instruction.Offset, "\t", instruction.Instruction.String())
	}
}

func setUpVMForPut(
	t *testing.T,
	e *neotest.Executor,
	contract *neotest.Contract,
	hasResult bool,
	methodOff int,
	num int,
	key []byte,
) (v *vm.VM) {
	ic, err := e.Chain.GetTestVM(trigger.Application, nil, nil)
	require.NoError(t, err)
	ic.VM.LoadNEFMethod(
		contract.NEF,
		contract.Hash,
		contract.Hash,
		callflag.All,
		hasResult,
		methodOff,
		-1,
		nil,
	)
	ic.VM.Context().Estack().PushVal(num)
	ic.VM.Context().Estack().PushVal(key)
	return ic.VM
}

func getStartOffset(di *compiler.DebugInfo, methodID string) (int, error) {
	for _, method := range di.Methods {
		if method.ID == methodID {
			return int(method.Range.Start), nil
		}
	}
	return 0, errors.New("Method not found")
}

func hasResult(di *compiler.DebugInfo, methodID string) (bool, error) {
	for _, method := range di.Methods {
		if method.ID == methodID {
			if method.ReturnType == "Void" {
				return false, nil
			}
			return true, nil
		}
	}
	return false, errors.New("Method not found")
}

func getNumToPut() int {
	return rand.Intn(100)
}
