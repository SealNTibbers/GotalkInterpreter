set -e -u

go get -d -v ./...

export MODE=TESTS

go test  $(go list ./... | grep -v docs ) \
   -race -coverprofile cover.out -covermode atomic

perc=`go tool cover -func=cover.out | tail -n 1 | sed -Ee 's!^[^[:digit:]]+([[:digit:]]+(\.[[:digit:]]+)?)%$!\1!'`
echo "Total coverage: $perc %"
res=`echo "$perc >= 50.0" | bc`
test "$res" -eq 1 && exit 0
echo "Insufficient coverage: $perc" >&2
exit 1

