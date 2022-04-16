build:
	mkdir -p bin
	go build -o bin

test:
	./bin/gitChangeGoPath -test-dir tests -base-folder-name gitChangeGoPath -master-repo-branch-name master