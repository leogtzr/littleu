.DEFAULT_GOAL := install

# INSTALL_SCRIPT=./install.sh
BIN_FILE=littleu

install:
	go build -o "${BIN_FILE}"
	echo "Build successfull, running binary."
	./"${BIN_FILE}"

clean:
	go clean
	rm --force "cp.out"
	rm --force nohup.out

test:
	go test

check:
	go test

cover:
	go test -coverprofile cp.out
	go tool cover -html=cp.out

run:
	./"${BIN_FILE}"
