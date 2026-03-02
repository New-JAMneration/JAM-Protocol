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
│   ├── INDEX.md           # Documentation index (start here)
│   └── ...                # See INDEX.md for full list
└── <module>/              # Module-level documentation
    └── README.md          # See FOLDER_STRUCTURE.md for list
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

**Entry point:** Start with [INDEX.md](./INDEX.md) for a categorized list of all documents in this directory.

**Key documents:**

| Document | Description |
|----------|-------------|
| `INDEX.md` | **Documentation index** - categorized list of docs in `READMERef/` |
| `DOCUMENTATION_POLICY.md` | Documentation structure and maintenance rules (this file) |
| `FOLDER_STRUCTURE.md` | Project folder structure and module READMEs |

For the full list of documents in `READMERef/`, see [INDEX.md](./INDEX.md).
For module-specific documentation, see [FOLDER_STRUCTURE.md](./FOLDER_STRUCTURE.md).

**Important:** Only cross-module documentation belongs in `READMERef/`. Module-specific docs should stay in their respective folders.

### 2.3 Module README.md

Each module with public APIs or significant complexity should have a `README.md`:

- Located at `<module>/README.md`
- Focuses **only** on that module
- Should include: purpose, file structure, API reference, usage examples
- Should NOT duplicate project-level content

For the list of current module READMEs, see [FOLDER_STRUCTURE.md](./FOLDER_STRUCTURE.md).

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
| **READMERef documentation index** | `READMERef/INDEX.md` |
| **Project folder structure & module READMEs** | `READMERef/FOLDER_STRUCTURE.md` |
| How to set up the project | `README.md` |

### Adding New Documentation

When adding new documentation, update the following:

1. **New module**:
   - Create `<module>/README.md`
   - Update `FOLDER_STRUCTURE.md` (add to structure and module READMEs section)

2. **New doc in `READMERef/`**:
   - Create the file using `UPPER_SNAKE_CASE.md` naming
   - Update `INDEX.md`

3. **New workflow**:
   - Add to existing doc or create new one in `READMERef/`
   - Update `INDEX.md`

