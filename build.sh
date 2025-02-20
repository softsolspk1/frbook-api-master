# GOOS=linux GOARCH=amd64 go build --o frBook-api
# scp frBook-api  root@134.209.152.216:~/api/

#!/bin/bash

#!/bin/bash

# Set GO environment variables only if needed
export GOOS=linux
export GOARCH=amd64

# Build the binary
go build -o frBook-api
