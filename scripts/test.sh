go test -coverprofile ./c.out $(go list ./...) || exit $?
echo $?