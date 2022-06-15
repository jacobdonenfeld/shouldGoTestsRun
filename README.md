## shouldGoTestsRun

Tool used to determine if files changed in git branch would affect specified test files. 

Intended to be used by CI resources to determine if tests need to be ran. Unfinished

### Example:
##### Gitlab ci:

```yaml
example_job:
  stage: test
  script:
    - cd $BASE_GIT_DIR
    # This will exit 1 if the tests don't need to be ran
    - if gitChangeGoPath -test-dir ./tests -base-folder-name $REPO_NAME -master-repo-branch-name origin/master; then exit 0; fi
    - go test ./tests 
```

### TODO:
- Have this run on individual test paths
- use golang.org/x/tools/go/packages: https://eli.thegreenplace.net/2020/writing-multi-package-analysis-tools-for-go/
