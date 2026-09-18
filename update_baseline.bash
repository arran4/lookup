#!/bin/bash
cd /app
go test -v ./jsonata/... > test_out.txt 2>&1
grep "Unexpected pass! Test" test_out.txt | sed -e 's/.*Test //' -e 's/ is marked.*//' > unexpected_passes.txt
for test in $(cat unexpected_passes.txt); do
    echo "Removing $test from baseline"
    sed -i "s|\"$test\":.*||g" /app/jsonata/baseline.go
done
sed -i '/^[[:space:]]*$/d' /app/jsonata/baseline.go
go fmt /app/jsonata/baseline.go
