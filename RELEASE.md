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

# Create annotated tag
git tag -a package/vX.Y.Z -m "Release message"

# Push specific tags
git push origin package/v1.0.0 package2/v2.0.0

# Update dependency
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

- [ ] Ensure all tests pass: `make test`
- [ ] Run linting: `make lint`
- [ ] Update dependencies: `make up && make tidy`
- [ ] Review and update CHANGELOG.md for each package
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

2. Create annotated tags with comprehensive release notes:

   ```bash
   # Create annotated tag with release message
   git tag -a cmp/v0.3.0 -m "darvaza.org/x/cmp v0.3.0

   Brief description of the release

   Changes since vX.Y.Z:
   - List of changes
   - Breaking changes should be clearly marked
   - New features
   - Bug fixes

   Dependencies:
   - darvaza.org/core vX.Y.Z
   - Go 1.23 or later"
   ```

   For multiple packages, consider using message files:

   ```bash
   # Create message file
   cat > tag-cmp.txt <<EOF
   darvaza.org/x/cmp v0.3.0

   Release description...
   EOF

   # Create tag using message file
   git tag -a cmp/v0.3.0 -F tag-cmp.txt
   ```

3. Push all tags at once:

   ```bash
   git push origin cmp/v0.3.0 config/v0.5.0 sync/v0.3.0 \
     fs/v0.5.1 container/v0.3.0
   ```

4. Wait for pkg.go.dev to index the new versions (usually 5-10 minutes).

5. Document the release (e.g., PR comment, release notes):

   ```bash
   # Example PR comment
   gh pr comment PR_NUMBER --body "## Tier 1 Packages Released

   The following packages have been released:

   \`\`\`bash
   go get darvaza.org/x/cmp@v0.3.0
   go get darvaza.org/x/config@v0.5.0
   go get darvaza.org/x/sync@v0.3.0
   go get darvaza.org/x/fs@v0.5.1
   go get darvaza.org/x/container@v0.3.0
   \`\`\`"
   ```

### 3. Update Tier 2 Dependencies

1. Update go.mod files in Tier 2 packages to use the new versions:

   ```bash
   # Update without changing directories (avoiding confusion)
   go -C net get darvaza.org/x/fs@v0.5.1
   go -C net mod tidy

   go -C web get darvaza.org/x/fs@v0.5.1
   go -C web mod tidy

   go -C tls get darvaza.org/x/container@v0.3.0
   go -C tls mod tidy
   ```

   Alternative approach using make:

   ```bash
   # If the Makefile supports it
   make -C net up
   make -C web up
   make -C tls up
   ```

2. Run tests to ensure compatibility:

   ```bash
   make test
   ```

3. Commit the dependency updates:

   ```bash
   git add -A
   git commit -m "build: update internal dependencies for release

   - net: update fs to v0.5.1
   - web: update fs to v0.5.1
   - tls: update container to v0.3.0"
   ```

### 4. Tier 2 Release

1. Check current Tier 2 versions:

   ```bash
   git tag --list | grep -E "^(net|web|tls)/" | sort -V
   ```

2. Create annotated tags for Tier 2 packages following the same pattern as
   Tier 1:

   ```bash
   git tag -a net/v0.3.0 -m "darvaza.org/x/net v0.3.0

   Release with updated dependencies

   Changes since vX.Y.Z:
   - Update darvaza.org/x/fs to v0.5.1
   - Other changes...

   Dependencies:
   - darvaza.org/core vX.Y.Z
   - darvaza.org/x/fs v0.5.1
   - Go 1.23 or later"
   ```

3. Push all Tier 2 tags:

   ```bash
   git push origin net/v0.3.0 web/v0.3.0 tls/v0.3.0
   ```

4. Document the complete release:

   ```bash
   gh pr comment PR_NUMBER --body "## All Packages Released

   Tier 2 packages have been released:

   \`\`\`bash
   go get darvaza.org/x/net@v0.3.0
   go get darvaza.org/x/web@v0.3.0
   go get darvaza.org/x/tls@v0.3.0
   \`\`\`

   All packages now require Go 1.23 or later."
   ```

## Version Numbering

Each package maintains its own semantic version. When releasing:

- **Patch version** (v0.2.x): Bug fixes, documentation updates
- **Minor version** (v0.x.0): New features, backwards-compatible changes
- **Major version** (vx.0.0): Breaking changes

### Common Release Scenarios

1. **Updating Go version requirement** (e.g., Go 1.22 → 1.23):
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
