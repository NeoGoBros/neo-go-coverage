#! /bin/bash

set -e

CONTRACT_FILENAME=contract.go
OUTPUT_COV_FILE=kate.cov

cd ../tests; go test; cd ../experiment

echo "line.column,line.column numberOfStatements count" > $OUTPUT_COV_FILE
cat ../tests/c.out | grep $CONTRACT_FILENAME | sed 's/^[^:]*://' >> $OUTPUT_COV_FILE
