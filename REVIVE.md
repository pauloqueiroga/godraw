# Godraw Revival Plan (Fall 2026)

This plan describes my intentions to revive this library after so many years of being dormant. It addresses the findings described below.

### Security
- Style injection: `NewImage` puts the URL straight into the style string. A URL containing `;` injects extra style keys (I confirmed `;rounded=1` becomes its own key). This matters if callers pass untrusted URLs.
- HTML labels: draw.io renders `Value` as HTML when the style has `html=1`. The library should document that it escapes XML only and that callers must sanitize labels.
- XML parsing: Go's `encoding/xml` does not resolve external entities, so there is no XXE risk. There is no input size limit; that is the caller's job, but it is worth documenting.
- CI supply chain: the workflow uses `SonarSource/sonarcloud-github-action@master`, an unpinned branch that receives your `SONAR_TOKEN`. That action is also deprecated.

### Hygiene and Distribution
- Go version: `go.mod` says `go 1.17` and CI pins Go 1.17, which has been unsupported for years.
- CI actions: `actions/checkout@v2` and `actions/setup-go@v2` are obsolete, and the workflow has no `permissions:` block.
- Line endings: `godraw.go` and `README.md` have Windows (CRLF) line endings, so `gofmt -l` flags the file. There is no `.gitattributes`.
- License: LGPL-2.1 is unusual for Go. Go links statically, which makes LGPL obligations awkward for users, so many companies avoid such libraries. The file header also says "All rights reserved", which contradicts an open license.
- Versions: v0.1.0 to v0.1.2 are already published on the Go module proxy. Published versions are immutable, so fixes ship as new tags.
- Naming and API:
  - `layerId` should be `layerID`.
  - The `Style` receiver is named `a`.
  - `Add` takes a `*Cell` but stores a copy, so later edits to the cell are silently lost.
  - `Vertex`, `Edge` and `Relative` are `"1"` strings rather than booleans.
  - `Width` and `Height` are strings while `X` and `Y` are ints.
  - There is no constructor for edges.
- Scope: real `.drawio` files wrap the model in `<mxfile><diagram>`, often compressed. The library only reads a bare `<mxGraphModel>`.

## The Plan

### Phase 1: Repository hygiene (no behavior change)
1. (done) Add `.gitattributes` with `* text=auto eol=lf` and renormalize, so `gofmt` is clean.
1. (done) Raise the `go` line in `go.mod` to a supported release. Go supports only the two latest releases; I'd set it to the older of those two.
1. (done) Decide on the license. I recommend MIT, BSD-3-Clause or Apache-2.0. You are the sole author, so you can relicense freely. Fix the "All rights reserved" header either way.

### Phase 2: Fix the bugs, tests first
1. Write table-driven tests that assert the actual XML output and round-trip equality. They should fail on bugs 1 to 6.
1. Fix the `NewImageXY` X/Y assignment.
1. Rewrite the style codec:
   - split with `strings.Cut` on the first `=` only;
   - skip empty segments;
   - write keys in sorted order with `strings.Builder`;
   - return an empty attribute name for an empty style, so the attribute is omitted.
1. Reject or escape `;` in style values, which closes the injection hole.
1. Add a fuzz test for the style round trip (`func FuzzStyle`, built into Go since 1.18). It would have caught bugs 2 and 3.
1. Add a runnable `Example` function with an `// Output:` block. It appears as documentation on pkg.go.dev and is verified by `go test`. Move the tests to package `godraw_test` so they exercise only the public API.
1. Release this as **v0.1.3**: fixes only, no API break.

### Phase 3: API cleanup (breaking, release as v0.2.0)
Breaking changes are acceptable before v1, but batch them into one release.

1. Change the geometry fields to `float64` (X, Y, Width, Height) and give vertices a default size.
1. Rename `layerId` to `layerID` and the receiver to `s`.
1. Make `Add` consistent: either take `Cell` by value or store `[]*Cell`.
1. Add `NewEdge(id, layerID, sourceID, targetID)`. Consider boolean fields with custom marshalling for `Vertex` and `Edge`.
1. Add a `Validate()` method that checks for duplicate IDs, dangling parent/source/target references and a missing root cell.
1. Optionally add `mxfile`/`diagram` wrapper support, with a size limit on decompression to prevent zip-bomb inputs.

### Phase 4: Static analysis and CI
1. Rewrite the workflow:
    - use current major versions of `actions/checkout` and `actions/setup-go`, with `go-version-file: go.mod` or a matrix of `stable` and `oldstable`;
    - add `permissions: contents: read` at the top;
    - pin third-party actions to a full commit SHA.
1. CI steps: `go mod verify`, `go vet ./...`, `go test -race -shuffle=on -coverprofile=...`, and `govulncheck ./...`. `govulncheck` is the official Go vulnerability scanner and reports only vulnerabilities your code actually reaches.
1. Add `golangci-lint` v2 with a `.golangci.yml`. It is the standard lint aggregator. A sensible set is `staticcheck`, `govet`, `errcheck`, `revive`, `gosec`, `misspell` and `gocritic`, with `gofumpt` and `goimports` as formatters.
1. Decide on SonarCloud: either drop it or migrate to `SonarSource/sonarqube-scan-action` pinned by SHA. I'd drop it for a library this size, because `golangci-lint` plus `govulncheck` covers more Go-specific ground. If you drop it, remove `sonar-project.properties` and the seven README badges. Note that the Sonar step currently fails on pull requests from forks, since forks don't receive secrets.
1. Add `.github/dependabot.yml` for the `github-actions` and `gomod` ecosystems.
1. In GitHub settings:
    - enable branch protection on `main` with required checks;
    - enable secret scanning and push protection;
    - enable CodeQL default setup;
    - enable private vulnerability reporting.
1. Optionally add the OpenSSF Scorecard action, which audits most of the above and gives you a badge.

### Phase 5: Distribution and documentation
1. A Go library needs no build artifacts. Distribution is a git tag: push `vX.Y.Z` and the module proxy and pkg.go.dev pick it up. Create a GitHub Release for each tag with notes.
1. Expand the README: install command, a short usage example, pkg.go.dev and Go Report Card badges, and the supported Go versions.
1. Add `CHANGELOG.md` and `SECURITY.md`. A `CONTRIBUTING.md` is optional.
1. Go to **v1.0.0** only once the Phase 3 API has settled. After v1, a breaking change requires a `/v2` module path, so there is no need to hurry.
