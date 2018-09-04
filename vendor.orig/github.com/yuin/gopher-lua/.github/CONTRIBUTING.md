# How to Contribute
Any kind of contributions are welcome.

## Building GopherLua

GopherLua uses simple inlining tool for generate efficient codes. This tool requires python interpreter. Files name of which starts with `_` genarate files name of which does not starts with `_` . For instance, `_state.go` generate `state.go` . You do not edit generated sources.
To generate sources, some make target is available.

```bash
make build
make glua
make test
```

You have to run `make build` before committing to the repository.

## Pull requests
Our workflow is based on the [github-flow](https://guides.github.com/introduction/flow/>) .

1. Create a new issue.
2. Fork the project.
3. Clone your fork and add the upstream.
    ```bash
    git remote add upstream https://github.com/yuin/gopher-lua.git
    ```

4. Pull new changes from the upstream.
    ```bash
    git checkout master
    git fetch upstream
    git merge upstream/master
    ```

5. Create a feature branch
    ```bash
    git checkout -b <branch-name>
    ```

6. Commit your changes and reference the issue number in your comment.
    ```bash
    git commit -m "Issue #<issue-ref> : <your message>"
    ```

7. Push the feature branch to your remote repository.
    ```bash
    git push origin <branch-name>
    ```

8. Open new pull request.
