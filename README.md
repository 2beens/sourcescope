# sourcescope
Small Golang Tool - get list of packages which depend on source from changed files

Useful when having to determine the list of packages that have to be unit tested after a certain change.

1. get source `git clone https://github.com/2beens/sourcescope`

2. `cd sourcescope` and run `go install`

3. `cd` into desired golang project dir, and run `sourcescope`

For more info detailed usage info, use `-h` / `--help` flags

Outputs the list of packages which depend on all files being changed in current `branch` 
