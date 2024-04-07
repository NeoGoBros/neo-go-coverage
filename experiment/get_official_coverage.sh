#! /bin/bash

set -e

TEMP_FILE=temp.txt
CONTRACT_FILENAME=contract.go
OUTPUT_COV_FILE=official.cov

echo "line.column,line.column numberOfStatements count" > $OUTPUT_COV_FILE

go test -coverprofile=$TEMP_FILE
cat $TEMP_FILE | grep $CONTRACT_FILENAME | sed 's/^[^:]*://' >> $OUTPUT_COV_FILE

rm $TEMP_FILE
