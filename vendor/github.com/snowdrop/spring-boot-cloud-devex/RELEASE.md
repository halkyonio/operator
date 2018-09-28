# Release the project and generate go release

## The easy way

Simply create a release from the GitHub UI named `release-${SOME_VERSION}`
where `SOME_VERSION` could be for example `0.3.0`

This will cause CircleCI to perform a release and will create a GitHub release named `v0.3.0`
that will contain the built binaries for both MacOS and Linux  

**Note**: This assumes that the `GITHUB_API_TOKEN` has been set in the CircleCI UI for the job

## Manually using make

Execute this command within the root of the project where you pass as parameters your `GITHUB_API_TOKEN` and `VERSION` which corresponds to the tag to be created

```yaml
make upload GITHUB_API_TOKEN=YOURTOKEN VERSION=0.3.0
```

**Remark** : You can next edit the release to add a `changelog` using this command `git log --oneline --decorate v0.2.0..v0.3.0`

**Note**: This command assumes that you have all necessary command line dependencies installed

## Using goreleaser

Tag the release and push it to the github repo

```bash
git tag -a v0.2.0 -m "Release fixing access to packages files"
git push origin v0.2.0
```

Next, use the [`goreleaser`](https://github.com/goreleaser/goreleaser) tool to build cross platform the project and publish it on github

Create the following `.goreleaser.yml` file
```yaml
builds:
- binary: sb
  env:
    - CGO_ENABLED=0
archive:
  replacements:
    darwin: Darwin
    linux: Linux
    windows: Windows
    386: i386
    amd64: x86_64
checksum:
  name_template: 'checksums.txt'
snapshot:
  name_template: "{{ .Tag }}-next"
changelog:
  sort: asc
  filters:
    exclude:
    - '^docs:'
    - '^test:'
```

Export your `GITHUB_TOKEN` and then execute this command to release

`goreleaser`