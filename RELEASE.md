# Release Process for darvaza.org/x

This document describes the release process for the darvaza.org/x mono-repo,
including dependency order and procedures to ensure consistent releases.

## Quick Reference

### Release Order

1. **Tier 1** (independent): cmp, config, sync, fs, container
2. **Tier 2** (dependent): net (→fs), web (→fs), tls (→container)

### Essential Commands

```bash
# Check current versions
git tag --list | grep -E "^(package)/" | sort -V

# Create signed annotated tag (requires a configured GPG/SSH signing key)
git tag -s package/vX.Y.Z -m "Release message"

# Push specific tags
git push origin package/v1.0.0 package2/v2.0.0

# Update a single internal dependency (avoid `make up`, which fans out
# to every external dep)
go -C package get darvaza.org/x/dependency@vX.Y.Z
go -C package mod tidy
```

## Package Dependencies

The following diagram shows the internal dependencies between packages:

```text
                Tier 1 - Independent packages:
┌─────────────┐ ┌──────────────┐ ┌─────────────┐
│     cmp     │ │    config    │ │     sync    │
│  (no deps)  │ │  (no deps)   │ │  (no deps)  │
└─────────────┘ └──────────────┘ └─────────────┘

┌─────────────┐                  ┌─────────────┐
│     fs      │                  │  container  │
│  (no deps)  │                  │  (no deps)  │
└──────┬──────┘                  └──────┬──────┘
       │                                │
       ├─────────────┐                  │
       ▼             ▼                  ▼
┌─────────┐   ┌─────────┐       ┌─────────────┐
│   net   │   │   web   │       │     tls     │
└─────────┘   └─────────┘       └─────────────┘

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

### Tier 2 - Dependent Packages

These packages depend on Tier 1 packages and must be released after their
dependencies:

- **darvaza.org/x/net** (depends on fs)
- **darvaza.org/x/web** (depends on fs)
- **darvaza.org/x/tls** (depends on container)

## Release Process

### 1. Pre-release Checklist

Before starting the release process:

- [ ] Ensure a full build is clean: `make`
- [ ] Tier 2 internal-dep bumps are targeted (`go -C <pkg> get
      darvaza.org/x/<dep>@vX.Y.Z`), not blanket `make up`
- [ ] Review and update CHANGELOG.md for each package (when present)
- [ ] Ensure all documentation is up to date
- [ ] Check current versions:
  `git tag --list | grep -E "^(cmp|config|sync|fs|container)/" | sort -V`
- [ ] Verify no uncommitted changes: `git status`

### 2. Tier 1 Release

1. Check the latest tags for each package to determine new version numbers:

   ```bash
   # List current tags
   git tag --list | grep -E "^(cmp|config|sync|fs|container)/" | sort -V
   ```

2. Create signed annotated tags with comprehensive release notes:

   ```bash
   # Create signed tag with release message inline
   git tag -s cmp/v0.2.2 -m "darvaza.org/x/cmp v0.2.2

   Brief description of the release

   Changes since vX.Y.Z:
   - List of changes
   - Breaking changes should be clearly marked
   - New features
   - Bug fixes

   Dependencies:
   - darvaza.org/core vX.Y.Z
   - Go 1.24 or later"
   ```

   For multiple packages, prefer message files in `.tmp/` (gitignored):

   ```bash
   # Compose the message in a scratch file
   $EDITOR .tmp/tag-cmp-v0.2.2.txt

   # Create the signed tag from the file
   git tag -s cmp/v0.2.2 -F .tmp/tag-cmp-v0.2.2.txt
   ```

3. Push all tags at once:

   ```bash
   git push origin cmp/v0.2.2 config/v0.5.1 sync/v0.3.1 fs/v0.5.3 container/v0.3.2
   ```

4. Wait for pkg.go.dev to index the new versions (usually 5-10 minutes).

5. Document the release (e.g., PR comment, release notes):

   ```bash
   # Example PR comment
   gh pr comment PR_NUMBER --body "## Tier 1 Packages Released

   The following packages have been released:

   \`\`\`bash
   go get darvaza.org/x/cmp@v0.2.2
   go get darvaza.org/x/config@v0.5.1
   go get darvaza.org/x/sync@v0.3.1
   go get darvaza.org/x/fs@v0.5.3
   go get darvaza.org/x/container@v0.3.2
   \`\`\`"
   ```

### 3. Update Tier 2 Dependencies

1. Update go.mod files in Tier 2 packages to use the new versions.
   Use targeted `go get` commands — avoid `make up`, which blanket-bumps
   every external dependency:

   ```bash
   go -C net get darvaza.org/x/fs@v0.5.3
   go -C net mod tidy

   go -C web get darvaza.org/x/fs@v0.5.3
   go -C web mod tidy

   go -C tls get darvaza.org/x/container@v0.3.2
   go -C tls mod tidy
   ```

2. Run a clean build to confirm compatibility:

   ```bash
   make
   ```

3. Commit the dependency updates with explicit paths (no `git add -A`):

   ```bash
   git commit -s -m "build: update internal dependencies for release

   - net: update fs to v0.5.3
   - web: update fs to v0.5.3
   - tls: update container to v0.3.2" \
     net/go.mod net/go.sum web/go.mod web/go.sum tls/go.mod tls/go.sum
   ```

### 4. Tier 2 Release

1. Check current Tier 2 versions:

   ```bash
   git tag --list | grep -E "^(net|web|tls)/" | sort -V
   ```

2. Create signed annotated tags for Tier 2 packages following the same
   pattern as Tier 1:

   ```bash
   git tag -s net/v0.6.3 -F .tmp/tag-net-v0.6.3.txt
   git tag -s web/v0.13.0 -F .tmp/tag-web-v0.13.0.txt
   git tag -s tls/v0.6.1 -F .tmp/tag-tls-v0.6.1.txt
   ```

   Each message file should follow the structure:

   ```text
   darvaza.org/x/net v0.6.3

   Release with updated dependencies

   Changes since vX.Y.Z:
   - Update darvaza.org/x/fs to v0.5.3
   - Other changes...

   Dependencies:
   - darvaza.org/core vX.Y.Z
   - darvaza.org/x/fs v0.5.3
   - Go 1.24 or later
   ```

3. Push all Tier 2 tags:

   ```bash
   git push origin net/v0.6.3 web/v0.13.0 tls/v0.6.1
   ```

4. Document the complete release:

   ```bash
   gh pr comment PR_NUMBER --body "## All Packages Released

   Tier 2 packages have been released:

   \`\`\`bash
   go get darvaza.org/x/net@v0.6.3
   go get darvaza.org/x/web@v0.13.0
   go get darvaza.org/x/tls@v0.6.1
   \`\`\`

   All packages now require Go 1.24 or later."
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
