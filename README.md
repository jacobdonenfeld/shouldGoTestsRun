## gitChangeGoPath
Name to be changed

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
    - gitChangeGoPath -test-dir ./tests -base-folder-name $REPO_NAME -master-repo-branch-name origin/master
    - go test ./tests 
```