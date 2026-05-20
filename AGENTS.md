# Repository Guidelines

## Project Structure & Module Organization

This repository is currently OpenSpec-first and will evolve into a Vue frontend plus Go backend system. Core files live in:

- `openspec/config.yaml`: repository-level OpenSpec configuration.
- `openspec/changes/<change-name>/`: active proposals with `proposal.md`, `design.md`, `tasks.md`, and capability specs.
- `openspec/specs/`: accepted baseline specs.
- `.codex/skills/`: local Codex workflow skills.
- `doc_*.md`: PRDs and source product documents.

When implementation code is added, keep a clean split such as `frontend/`, `backend/`, `infra/`, and `docker/`. Use kebab-case for change and spec directory names.

## Architecture & Technology Constraints

All contributors should follow these defaults unless a new approved spec changes them:

- database: MySQL
- cache: Redis
- frontend: Vue 3 + Vite + TypeScript
- backend: Golang, with go-zero allowed when it fits the module
- architecture: frontend/backend separation
- i18n: Chinese and English supported, default locale is Simplified Chinese
- deployment: Dockerfile-based

Do not introduce a single-process UI/backend coupling or swap MySQL/Redis ad hoc. If go-zero is used, keep generated structure readable and avoid hiding domain rules inside framework glue.

## Build, Test, and Development Commands

Current workflow is driven by OpenSpec:

- `openspec list --json`: list active changes
- `openspec new change "<name>"`: create a change scaffold
- `openspec status --change "<name>"`: inspect artifact progress
- `openspec instructions <artifact> --change "<name>" --json`: read artifact instructions
- `openspec validate "<name>"`: validate a change before review

Example:

```bash
openspec new change "add-weekly-review-dashboard"
openspec validate "add-weekly-review-dashboard"
```

When app modules exist, document local run commands for frontend, backend, MySQL, and Redis near the module code. Prefer container-friendly commands that match the Dockerfile layout.

## Coding Style & Naming Conventions

Write Markdown and specs in short, direct sections. Keep folders, modules, APIs, and Docker-related paths in kebab-case. Use `camelCase` for TypeScript symbols and idiomatic Go naming for backend packages and exported types. OpenSpec requirements must use:

- `### Requirement: ...`
- `#### Scenario: ...`

For UI copy and product text, keep translation keys stable and make Chinese the default-facing language.

## Testing Guidelines

Validation is mandatory and layered:

- Level 1: unit tests, must pass
- Level 2: integration tests, must pass
- Level 3: end-to-end tests, must pass for cross-component changes

Skipping any required level means the work is not complete. Always run `openspec validate "<change-name>"` for spec changes.

## Commit & Pull Request Guidelines

Repository Git history is not usable in this workspace, so follow concise imperative commits, preferably scoped, for example `docs: add milestone spec`.

PRs should include the problem statement, affected change path, `openspec validate` result, executed test levels, and screenshots or API notes when relevant.
