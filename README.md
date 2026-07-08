# Release Clerk

A tool to tag a new version based on the conventional commits history

## TODO
- [x] add global debug flag
- [x] override disable push from cli in `release` command
- [ ] add commit transform hook before parsing them to conventional commits
- [x] add parents field to commit
- [x] add an option to exclude merge commits (parents > 1) (opt-out)
- [ ] add colored output
- [ ] add lint command (should accept git pathspec and utilize commit transform hook before validation) 
      that fails on invalid conventional commit
