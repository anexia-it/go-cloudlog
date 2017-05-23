#!/bin/bash

GO_CLOUDLOG_COVERAGE_TMP=$(mktemp -d)
go test -v -coverprofile=${GO_CLOUDLOG_COVERAGE_TMP}/go_cloudlog.coverage ./
echo "mode: set" > ${GO_CLOUDLOG_COVERAGE_TMP}/combined.coverage
egrep -v '^mode:' ${GO_CLOUDLOG_COVERAGE_TMP}/go_cloudlog.coverage >> ${GO_CLOUDLOG_COVERAGE_TMP}/combined.coverage
go tool cover -html=${GO_CLOUDLOG_COVERAGE_TMP}/combined.coverage -o coverage.html

if [ ! -z "${GO_CLOUDLOG_COVERAGE_TMP}" ]
then
    rm -rf "${GO_CLOUDLOG_COVERAGE_TMP}"
fi
