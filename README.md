## shouldGoTestsRun

Tool used to determine if files changed in git branch would affect specified test files. 

Intended to be used by CI resources to determine if tests need to be ran.

Currently your base folder name needs to be the same as your root go.mod module name

### Example:
Want to test this? Clone the repo, and run make build then make test.
Then change some file in a test, run the command again. Then change it in a file the test imports, then run make test again. See how the tool walks the import trees and determines that the tests need to be ran.

##### Gitlab ci:

```yaml
example_job:
  stage: test
  script:
    - cd $BASE_GIT_DIR
    # This will exit 1 if the tests don't need to be ran
    - if shouldGoTestsRun -test-dir ./tests -base-folder-name $REPO_NAME -comparison-branch-name origin/master; then exit 0; fi
    - go test ./tests 
```

### TODO:
- use golang.org/x/tools/go/packages: https://eli.thegreenplace.net/2020/writing-multi-package-analysis-tools-for-go/
