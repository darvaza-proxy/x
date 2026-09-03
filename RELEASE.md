# Release Process for darvaza.org/x

This document describes the release process for the darvaza.org/x mono-repo,
including dependency order and procedures to ensure consistent releases.

## Quick Reference

### Release Order

1. **Tier 1** (independent): cmp, config, sync, fs, container, text, time
2. **Tier 2** (dependent): net (→fs, sync), web (→fs), tls (→container, sync)

### Essential Commands

```bash
# Check current versions
git tag --list '<pkg>/v*' | sort -V

# Create signed annotated tag (requires a configured GPG/SSH signing key)
git tag -s <pkg>/vX.Y.Z -F .tmp/tag-<pkg>-vX.Y.Z.txt

# Push specific tags
git push origin <pkg>/vX.Y.Z

# Update a single internal dependency (avoid `make up`, which fans out
# to every external dep)
go -C <pkg> get darvaza.org/x/<dep>@vX.Y.Z
go -C <pkg> mod tidy
```

## Package Dependencies

The following diagram shows the internal dependencies between packages:

```text
        Tier 1 - Independent packages:
┌─────────┐ ┌─────────┐ ┌───────────┐ ┌─────────┐
│   cmp   │ │ config  │ │   text    │ │  time   │
└─────────┘ └─────────┘ └───────────┘ └─────────┘

┌─────────┐ ┌─────────┐ ┌───────────┐
│   fs    │ │  sync   │ │ container │
└────┬────┘ └────┬────┘ └─────┬─────┘
     │           │            │
     ├───────┐   ├────────┐   │
     │       │   │        │   │
     ▼       ▼   ▼        ▼   ▼
┌───────┐ ┌─────────┐  ┌───────────┐
│  web  │ │   net   │  │    tls    │
└───────┘ └─────────┘  └───────────┘
        Tier 2 - Dependent packages
```

## Release Tiers

Packages must be released in the following order to maintain dependency
consistency:

### Tier 1 - Independent Packages

These packages have no internal dependencies within darvaza.org/x and can
be released in any order or simultaneously:

- **darvaza.org/x/cmp**
- **darvaza.org/x/config**
- **darvaza.org/x/sync**
- **darvaza.org/x/fs**
- **darvaza.org/x/container**
- **darvaza.org/x/text**
- **darvaza.org/x/time**

### Tier 2 - Dependent Packages

These packages depend on Tier 1 packages and must be released after their
dependencies:

- **darvaza.org/x/net** (depends on fs and sync)
- **darvaza.org/x/web** (depends on fs)
- **darvaza.org/x/tls** (depends on container and sync)

## Release Process

### 1. Pre-release Checklist

Before starting the release process:

- [ ] Ensure a full build is clean: `make`
- [ ] Tier 2 internal-dep bumps are targeted (`go -C <pkg> get
      darvaza.org/x/<dep>@vX.Y.Z`), not blanket `make up`
- [ ] Review and update CHANGELOG.md for each package (when present)
- [ ] Ensure all documentation is up to date
- [ ] Check current versions:

  ```bash
  git tag --list '*/v*' | sort -V
  ```

- [ ] Verify no uncommitted changes: `git status`

### 2. Tier 1 Release

1. Check the latest tags for each package to determine new version numbers:

   ```bash
   # List current tags
   git tag --list '<pkg>/v*' | sort -V
   ```

2. Create signed annotated tags. The tag message describes the release,
   following the structure below:

   ```bash
   # Compose the message in a scratch file (.tmp/ is gitignored)
   $EDITOR .tmp/tag-<pkg>-vX.Y.Z.txt

   # Create the signed tag from the file
   git tag -s <pkg>/vX.Y.Z -F .tmp/tag-<pkg>-vX.Y.Z.txt
   ```

   Each message file should follow the structure:

   ```text
   darvaza.org/x/<pkg> vX.Y.Z

   Brief description of the release

   Changes since vA.B.C:
   - List of changes
   - Breaking changes should be clearly marked
   - New features
   - Bug fixes

   Dependencies:
   - darvaza.org/core vX.Y.Z
   - Go 1.25 or later
   ```

3. Push all tags at once, one argument per tag:

   ```bash
   git push origin <pkg>/vX.Y.Z
   ```

4. Wait for pkg.go.dev to index the new versions (usually 5-10 minutes).

5. Document the release (e.g., PR comment, release notes):

   ```bash
   # Example PR comment
   gh pr comment PR_NUMBER --body "## Tier 1 Packages Released

   The following packages have been released:

   \`\`\`bash
   go get darvaza.org/x/<pkg>@vX.Y.Z
   \`\`\`"
   ```

### 3. Update Tier 2 Dependencies

1. Update go.mod files in Tier 2 packages to use the new versions.
   Use targeted `go get` commands — avoid `make up`, which blanket-bumps
   every external dependency. Repeat for each internal dependency of
   each Tier 2 package, as listed under [Release Tiers](#release-tiers):

   ```bash
   go -C <pkg> get darvaza.org/x/<dep>@vX.Y.Z
   go -C <pkg> mod tidy
   ```

2. Run a clean build to confirm compatibility:

   ```bash
   make
   ```

3. Commit the dependency updates with explicit paths (no `git add -A`),
   reading the message from a scratch file as the tags do:

   ```bash
   $EDITOR .tmp/commit-release-deps.txt
   git commit -s -F .tmp/commit-release-deps.txt <pkg>/go.mod <pkg>/go.sum
   ```

   The message lists each bump:

   ```text
   build: update internal dependencies for release

   - <pkg>: update <dep> to vX.Y.Z
   ```

### 4. Tier 2 Release

1. Check current Tier 2 versions:

   ```bash
   git tag --list '<pkg>/v*' | sort -V
   ```

2. Create signed annotated tags for Tier 2 packages following the same
   pattern as Tier 1:

   ```bash
   git tag -s <pkg>/vX.Y.Z -F .tmp/tag-<pkg>-vX.Y.Z.txt
   ```

   Each message file should follow the structure:

   ```text
   darvaza.org/x/<pkg> vX.Y.Z

   Release with updated dependencies

   Changes since vA.B.C:
   - Update darvaza.org/x/<dep> to vX.Y.Z
   - Other changes...

   Dependencies:
   - darvaza.org/core vX.Y.Z
   - darvaza.org/x/<dep> vX.Y.Z
   - Go 1.25 or later
   ```

3. Push all Tier 2 tags at once, one argument per tag:

   ```bash
   git push origin <pkg>/vX.Y.Z
   ```

4. Document the complete release:

   ```bash
   gh pr comment PR_NUMBER --body "## All Packages Released

   Tier 2 packages have been released:

   \`\`\`bash
   go get darvaza.org/x/<pkg>@vX.Y.Z
   \`\`\`

   All packages now require Go 1.25 or later."
   ```

## Version Numbering

Each package maintains its own semantic version. When releasing:

- **Patch version** (v0.2.x): Bug fixes, documentation updates
- **Minor version** (v0.x.0): New features, backwards-compatible changes
- **Major version** (vx.0.0): Breaking changes

### Common Release Scenarios

1. **Updating Go version requirement** (e.g., Go 1.24 → 1.25):
   - This is a breaking change requiring minor version bump
   - Update all packages even if no code changes
   - Document clearly in release notes

2. **Updating core dependencies** (e.g., darvaza.org/core):
   - If API changes require code updates, bump minor version
   - Document any behaviour changes in release notes

3. **Adding new features**:
   - Minor version bump for the affected package only
   - Other packages remain at current versions

4. **Security fixes**:
   - Patch version for affected packages
   - Consider releasing all packages if the fix is in a shared dependency

## Automation Considerations

For future automation, consider:

1. A script that checks internal dependencies and enforces release order
2. Automated version bumping based on commit messages
3. GitHub Actions workflow for coordinated releases
4. Dependency update PRs when Tier 1 packages are released

## Troubleshooting

### Common Issues

1. **Dependency version conflicts**: Ensure all internal dependencies use
   compatible versions before releasing.

2. **Missing tags**: If a package is not found after tagging, ensure the tag
   follows the format `packagename/vX.Y.Z`.

3. **Build failures in dependent packages**: Update and test Tier 2 packages
   with new Tier 1 versions before tagging.

### Rollback Procedure

If issues are discovered after release:

1. Do not delete tags (they may already be cached)
2. Release a new patch version with the fix
3. Update dependent packages if necessary

## See also

- [README.md](README.md): General repository information.
- [AGENTS.md](AGENTS.md): Development guidelines for AI agents.
