package covertest

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/compiler"
	"github.com/stretchr/testify/require"
)

var mu sync.Mutex

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
	mu.Lock()
	defer mu.Unlock()

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
