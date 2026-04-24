# Contributing to Qaynaq

Thanks for your interest in contributing! This document covers how to propose changes and what we expect from contributors.

## Contributor License Agreement

Before we can accept your contribution, you need to sign our [Contributor License Agreement](./CLA.md). This is a one-time step.

When you open your first pull request, the CLA Assistant bot will comment with a link. Click it, review the CLA, and sign. Your PR cannot be merged until the CLA is signed.

Why: signing the CLA gives us the right to relicense or dual-license your contribution alongside the rest of the project. This flexibility matters because Qaynaq may offer commercial or enterprise editions in the future. You retain the copyright to your contribution.

If you are contributing on behalf of a company, please reach out so we can arrange a Corporate CLA that covers all contributors from your organization.

## Development setup

```bash
git clone https://github.com/qaynaq/qaynaq.git
cd qaynaq
cp .env.example .env          # edit .env and set ROLE, SECRET_KEY
make ui-deps                  # install UI dependencies
make bundle                   # build UI + Go binary
make coordinator              # run locally
```

See the [README](./README.md) for more details.

## Proposing changes

1. **Open an issue first** for anything non-trivial (new feature, architectural change, behavior change). A quick discussion saves everyone time.
2. **Fork and branch.** Use descriptive branch names (`feat/...`, `fix/...`, `docs/...`).
3. **Keep PRs focused.** One logical change per PR. Split unrelated work into separate PRs.
4. **Write clear commit messages.** We squash-merge, so the PR title becomes the commit message - make it descriptive.
5. **Update tests and docs** when your change affects behavior or interfaces.

## Coding standards

- **Go**: follow `gofmt` and `go vet`. Run `go test ./...` before pushing.
- **TypeScript/UI**: follow the existing ESLint and Prettier config in `ui/`.
- **Commits**: single-line subject, imperative mood ("add X", not "added X"). No em dashes.

## Reporting bugs

Use GitHub Issues. Include:
- Qaynaq version (`qaynaq --version`)
- OS / environment (Docker, local binary, etc.)
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs (scrub any secrets first)

## Reporting security issues

Please do **not** open a public issue for security vulnerabilities. Email the maintainers directly or use GitHub's private vulnerability reporting feature on the repository.

## Questions

Open a discussion on GitHub, or reach out on whatever channel is linked from the README.
