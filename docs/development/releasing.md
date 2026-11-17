# Release Procedure

This document describes the steps required to produce a MeiliScope release.

---

## Prerequisites

- Phase 6 approval must be on record. See `docs/development/release-readiness.md`.
- All CI checks must be passing on the `main` branch.
- No unresolved Critical or High security defects.

---

## 1. Select the Release Version

Choose a Semantic Version tag in the form `vMAJOR.MINOR.PATCH`.

Pre-release versions use a suffix: `v1.0.0-rc.1`, `v0.9.0-beta.1`.

The tag is the single authoritative version. It propagates to:
- Archive filenames (`MeiliScope_v1.0.0_windows_amd64.zip`)
- GitHub Release title
- CHANGELOG links

---

## 2. Update CHANGELOG.md

Move items from `[Unreleased]` to a new release section:

```markdown
## [1.0.0] — YYYY-MM-DD
```

Add a comparison link at the bottom of the file.

Commit this change to `main`:
```bash
git add CHANGELOG.md
git commit -m "Release v1.0.0"
```

---

## 3. Run the Full Test Suite

```bash
# Go checks
gofmt -l .
go vet ./...
go test ./internal/...

# Frontend
cd frontend && npm test && cd ..

# Production build
wails3 build
```

All checks must pass before creating a tag.

---

## 4. Create the Release Tag

**Do not create or push a tag without authorization.**

After explicit approval:
```bash
git tag -a v1.0.0 -m "MeiliScope v1.0.0"
git push origin v1.0.0
```

Pushing a tag matching `v[0-9]+.[0-9]+.[0-9]+` triggers the release workflow automatically.

---

## 5. Monitor the Release Workflow

The release workflow runs these jobs in sequence:

```
validate-version → go-checks + frontend-tests → build-windows + build-linux → package → publish (draft)
```

If any job fails:
- Do not proceed.
- Check the job logs.
- Fix the issue.
- If the tag has not been used in a public release, delete and recreate it after the fix:
  ```bash
  git tag -d v1.0.0
  git push origin :refs/tags/v1.0.0
  ```

---

## 6. Verify the Draft Release

After the workflow completes, a **draft** GitHub Release is created.

Verify:
- [ ] Version matches the tag
- [ ] All three artifacts are present: `MeiliScope_vX.Y.Z_windows_amd64.zip`, `MeiliScope_vX.Y.Z_linux_amd64.tar.gz`, `SHA256SUMS.txt`
- [ ] SHA256SUMS.txt checksums match the actual files
- [ ] Release notes are accurate
- [ ] No sensitive files in the archives

Download and test the artifacts locally:
```bash
# Windows (PowerShell)
Get-FileHash MeiliScope_v1.0.0_windows_amd64.zip -Algorithm SHA256

# Linux
sha256sum -c SHA256SUMS.txt --ignore-missing
tar -tvf MeiliScope_v1.0.0_linux_amd64.tar.gz
```

---

## 7. Obtain Publication Approval

Publication is not automatic. The draft release must be reviewed and approved before it is published.

Required:
- Artifact verification complete
- Smoke test on Windows (clean extraction, application launch, basic workflow)
- Smoke test on Linux if possible
- No release blockers

---

## 8. Publish the Release

After explicit authorization, edit the draft release on GitHub and uncheck **Save as draft**. Click **Publish release**.

---

## 9. Verify the Published Release

After publication:
- [ ] Release is visible at `https://github.com/taqu/meiliscope/releases`
- [ ] Download links for all artifacts work
- [ ] SHA256SUMS.txt is accessible and correct
- [ ] Release notes are displayed correctly

---

## Rollback / Withdrawal

If a defective release is published:

1. Immediately add a release notice explaining the issue.
2. Consider marking the release as a pre-release to reduce visibility.
3. Prepare a corrected release with an incremented patch version.
4. Do not delete a published release silently — users who downloaded it should be informed.
5. Do not reuse a version tag for a different set of artifacts.

---

## Local Packaging Verification

To build and verify release packages locally without uploading to GitHub:

```bash
# Build
wails3 build

# Windows package
VERSION="v1.0.0"
WIN_ZIP="MeiliScope_${VERSION}_windows_amd64.zip"
zip "${WIN_ZIP}" bin/meiliscope.exe LICENSE README.md

# Linux package
LINUX_TGZ="MeiliScope_${VERSION}_linux_amd64.tar.gz"
tar -czf "${LINUX_TGZ}" -C bin meiliscope
tar -rf "${LINUX_TGZ}" LICENSE README.md

# Checksums
sha256sum "${WIN_ZIP}" "${LINUX_TGZ}" > SHA256SUMS.txt
sha256sum --check SHA256SUMS.txt

# List contents
unzip -l "${WIN_ZIP}"
tar -tvf "${LINUX_TGZ}"
```
