package covertest

import (
	"errors"

	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/opcode"
	"github.com/nspcc-dev/neo-go/pkg/vm/vmstate"
)

// InstrHash maps instruction with scripthash of a contract it belongs to.
type InstrHash struct {
	Offset      int
	Instruction opcode.Opcode
	ScriptHash  util.Uint160
}

// Run starts execution of the loaded program and accumulates all seen opcodes
// together with the scripthash of a contract they belong to.
// Original vm.Run(): https://github.com/nspcc-dev/neo-go/blob/v0.101.3/pkg/vm/vm.go#L418
func Run(v *vm.VM) ([]InstrHash, error) {
	if !v.Ready() {
		return nil, errors.New("no program loaded")
	}

	if v.HasFailed() {
		// VM already ran something and failed, in general its state is
		// undefined in this case so we can't run anything.
		return nil, errors.New("VM has failed")
	}

	// vmstate.Halt (the default) or vmstate.Break are safe to continue.
	var ops []InstrHash
	for {
		switch {
		case v.HasFailed():
			// Should be caught and reported already by the v.Step(),
			// but we're checking here anyway just in case.
			return ops, errors.New("VM has failed")
		case v.HasHalted(), v.AtBreakpoint():
			// Normal exit from this loop.
			return ops, nil
		case v.State() == vmstate.None:
			nStr, curInstr := v.Context().NextInstr()
			ops = append(ops, InstrHash{
				Offset:      nStr,
				Instruction: curInstr,
				ScriptHash:  v.Context().ScriptHash(),
			})
			if err := v.Step(); err != nil {
				return ops, err
			}
		default:
			return ops, errors.New("unknown state")
		}
	}
}
