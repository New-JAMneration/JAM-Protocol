# Documentation Policy

This document defines the documentation structure, maintenance rules, and
guidelines for the **JAM-Protocol** project. Its goal is to keep documentation
organized, up-to-date, and easy to navigate.

---

## 1. Documentation Structure

The project uses a **layered documentation** approach:

```
JAM-Protocol/
├── README.md              # Project entry point
├── READMERef/             # Project-level documentation
│   ├── CODE_FORMATTING.md
│   ├── DEVELOPMENT_DOC.md
│   ├── DOCUMENTATION_POLICY.md  (this file)
│   ├── ENCODER_AND_DECODER.md
│   ├── FOLDER_STRUCTURE.md
│   ├── GOLANG_TEST.md
│   ├── LOGGER_USAGE.md
│   ├── RECORD_TERMINAL.md
│   ├── RELEASE_AND_PUBLISH.md
│   ├── RUST_VRF_COMPILE_GUIDE.md
│   └── SEMANTIC_COMMIT_MESSAGES.md
└── <module>/              # Module-level documentation
    └── README.md
```

### 1.1 Documentation Layers

| Layer | Location | Content | Audience |
|-------|----------|---------|----------|
| **Project Entry** | `README.md` | Quick start, setup, key links | New users, external visitors |
| **Project-level** | `READMERef/` | Cross-module concepts, policies, workflows | Developers, maintainers |
| **Module-level** | `<module>/README.md` | Module API, usage, internal design | Developers working on that module |

---

## 2. Documentation Types and Responsibilities

### 2.1 Root README.md

The root `README.md` should contain:

- Project name and one-line description
- Badges (CI status, etc.)
- Quick start / setup instructions
- Links to documentation in `READMERef/`
- **Should NOT contain**: detailed API docs, lengthy tutorials

### 2.2 READMERef/ Directory

Project-level documentation that spans multiple modules.

**Current documents:**

| Document | Description |
|----------|-------------|
| `CODE_FORMATTING.md` | Code formatting guidelines using `gofmt` |
| `DEVELOPMENT_DOC.md` | Development documentation and guidelines |
| `ENCODER_AND_DECODER.md` | SCALE codec encoder/decoder usage |
| `FOLDER_STRUCTURE.md` | Project folder structure overview |
| `GOLANG_TEST.md` | Go testing guidelines |
| `LOGGER_USAGE.md` | Logger usage instructions |
| `RECORD_TERMINAL.md` | Terminal recording guidelines |
| `RELEASE_AND_PUBLISH.md` | Release and publish workflow |
| `RUST_VRF_COMPILE_GUIDE.md` | Rust VRF library compilation guide |
| `SEMANTIC_COMMIT_MESSAGES.md` | Commit message conventions |

**Important:** Only cross-module documentation belongs in `READMERef/`. Module-specific docs should stay in their respective folders.

### 2.3 Module README.md

Each module with public APIs or significant complexity should have a `README.md`:

- Located at `<module>/README.md`
- Focuses **only** on that module
- Should include: purpose, file structure, API reference, usage examples
- Should NOT duplicate project-level content

**Current module READMEs:**

- `cmd/fuzz/README.md` - Fuzzing tool usage and test data
- `pkg/erasure_coding/reed-solomon-ffi/README.md` - Reed-Solomon FFI library
- `pkg/test_data/README.md` - Test data information

---

## 3. When to Update Documentation

### 3.1 Required Updates

Documentation **must** be updated when:

| Change | Required Doc Update |
|--------|---------------------|
| New public API | Module README or `FOLDER_STRUCTURE.md` |
| API behavior change | Module README |
| New module added | Create module README, update `FOLDER_STRUCTURE.md` |
| Setup/install change | Root `README.md` |
| Workflow change | Relevant doc in `READMERef/` |
| New design decision | Relevant doc in `READMERef/` |

### 3.2 PR Checklist

Before submitting a PR, ask yourself:

- [ ] Did I add/change any public API? → Update module README
- [ ] Did I add a new file/module? → Update `FOLDER_STRUCTURE.md`
- [ ] Did I change setup steps? → Update root `README.md`
- [ ] Did I change any workflow? → Update relevant doc in `READMERef/`
- [ ] Is this a large change? → Consider adding or updating a doc

---

## 4. Maintenance Rules

### 4.1 Keep Documentation Close to Code

- Module documentation lives **next to** the module code
- When you modify code, check if the nearby README needs updating

### 4.2 Single Source of Truth (DRY)

- Do NOT duplicate content across multiple docs
- Use **links** to reference other docs instead of copying
- Example: Root README links to `READMERef/` rather than listing all details

### 4.3 Staleness Indicators

For documents that may become outdated, add a note at the top:

```markdown
> **Note:** This layout can change as we go; we try to keep the doc
> in sync when we touch the code.
```

### 4.4 Review in PRs

- Reviewers should check if documentation updates are needed
- Documentation changes should be part of the same PR as code changes
  (not a separate follow-up PR)

---

## 5. Naming Conventions

### 5.1 File Names

| Type | Convention | Example |
|------|------------|---------|
| Policy docs in `READMERef/` | `UPPER_SNAKE_CASE.md` | `DOCUMENTATION_POLICY.md`, `CODE_FORMATTING.md` |
| Module README | `README.md` | `cmd/fuzz/README.md` |

### 5.2 Section Headers

- Use title case: `## When to Update Documentation`
- Use numbered sections for long documents: `## 1. Documentation Structure`

---

## 6. Quick Reference

### Where to Find What

| Looking for... | Go to... |
|----------------|----------|
| How to set up the project | `README.md` |
| Code formatting | `READMERef/CODE_FORMATTING.md` |
| Commit message format | `READMERef/SEMANTIC_COMMIT_MESSAGES.md` |
| Project folder structure | `READMERef/FOLDER_STRUCTURE.md` |
| Encoder/Decoder usage | `READMERef/ENCODER_AND_DECODER.md` |
| Go testing guidelines | `READMERef/GOLANG_TEST.md` |
| Logger usage | `READMERef/LOGGER_USAGE.md` |
| Release workflow | `READMERef/RELEASE_AND_PUBLISH.md` |
| Rust VRF compilation | `READMERef/RUST_VRF_COMPILE_GUIDE.md` |
| Fuzzing tool | `cmd/fuzz/README.md` |
| Erasure coding FFI | `pkg/erasure_coding/reed-solomon-ffi/README.md` |

### Adding New Documentation

1. **New module**: Create `<module>/README.md`, update `FOLDER_STRUCTURE.md`
2. **New doc in `READMERef/`**: Create the file, update this policy if needed
3. **New workflow**: Add to existing doc or create new one in `READMERef/`

---

## 7. Future Improvements (Optional)

These are not required but could be added later:

- [ ] Add `markdown-link-check` to CI to detect broken links
- [ ] Add GoDoc comments to public APIs for auto-generated documentation
- [ ] Create an `INDEX.md` in `READMERef/` for easier navigation
