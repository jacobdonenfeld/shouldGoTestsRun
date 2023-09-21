build:
	mkdir -p bin
	go build -o bin

test:
	./bin/shouldGoTestsRun -test-dir tests -base-folder-name shouldGoTestsRun -comparison-branch-name master