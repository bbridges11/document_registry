# Skills Index

Quick access to all available skills for this project.

## 📁 Skill Files

| Skill | File | Use When |
|-------|------|----------|
| **Project Bootstrap** | [go-project-bootstrap.md](go-project-bootstrap.md) | Starting new service from scratch |
| **Feature Implementation** | [go-feature-implementation.md](go-feature-implementation.md) | Adding feature to existing service |
| **Aggregate Scaffold** | [go-aggregate-scaffold.md](go-aggregate-scaffold.md) | Creating new aggregate/entity |
| **Architecture Refactor** | [go-architecture-refactor.md](go-architecture-refactor.md) | Fixing architecture violations |
| **Generate Collection** | [generate-postman-collection.md](generate-postman-collection.md) | Creating API documentation |

## 📖 Documentation

| Document | File | Purpose |
|----------|------|---------|
| **Complete Guide** | [README.md](README.md) | Full skills documentation |
| **Quick Reference** | [QUICK_REFERENCE.md](QUICK_REFERENCE.md) | One-page cheat sheet |
| **Cursor Rules** | [../.cursorrules](../.cursorrules) | Cursor IDE configuration |
| **Copilot Instructions** | [../.github/copilot-instructions.md](../.github/copilot-instructions.md) | GitHub Copilot configuration |

## 🎯 Quick Start

### 1. Choose Your AI Tool

**Claude.ai:**
```
Upload skill file → Reference in prompt
"Using go-feature-implementation skill, add..."
```

**GitHub Copilot:**
```
@workspace with .claude-skills/go-feature-implementation.md
"Add a comments feature..."
```

**Cursor:**
```
@docs .claude-skills/go-feature-implementation.md
"Implement tagging system..."
```

### 2. Select Right Skill

```
New service?       → go-project-bootstrap.md
New feature?       → go-feature-implementation.md
New aggregate?     → go-aggregate-scaffold.md
Fix violations?    → go-architecture-refactor.md
API docs?          → generate-postman-collection.md
```

### 3. Craft Your Prompt

**Template:**
```
Using [SKILL_NAME] skill, [ACTION] that [DESCRIPTION].

Requirements:
- [REQ_1]
- [REQ_2]
- [REQ_3]
```

**Example:**
```
Using go-feature-implementation skill, add a rating feature 
that allows users to rate documents.

Requirements:
- Ratings: 1-5 stars
- One rating per user per document
- Show average rating
- List top-rated documents
```

## 🔍 Detailed Descriptions

### go-project-bootstrap.md
**Creates:** Complete new Go service with full architecture
**Output:** Repository structure, platform setup, example feature
**Time saved:** 4-6 hours of scaffolding

### go-feature-implementation.md
**Creates:** Feature mapped to all architecture layers
**Output:** Domain, application, adapters, wiring
**Time saved:** 2-3 hours per feature

### go-aggregate-scaffold.md
**Creates:** Complete aggregate with full CRUD
**Output:** Entity, use cases, HTTP endpoints, events
**Time saved:** 3-4 hours per aggregate

### go-architecture-refactor.md
**Creates:** Refactor plan with updated code
**Output:** Violation fixes, layer separation, clean code
**Time saved:** 1-2 hours of cleanup

### generate-postman-collection.md
**Creates:** Postman/Bruno API collection
**Output:** Complete collection with all endpoints
**Time saved:** 1-2 hours of manual documentation

## 💡 Pro Tips

1. **Read the skill first** - Understand what it will do
2. **Be specific** - Clear requirements = better output
3. **Review output** - Always review generated code
4. **Test thoroughly** - Run tests after generation
5. **Combine skills** - Bootstrap → Feature → Refactor

## 🔗 Related Files

```
.
├── .claude-skills/          ← You are here
│   ├── README.md            ← Full documentation
│   ├── QUICK_REFERENCE.md   ← Cheat sheet
│   ├── INDEX.md             ← This file
│   └── *.md                 ← Individual skills
├── .cursorrules             ← Cursor configuration
├── .github/
│   └── copilot-instructions.md  ← Copilot config
├── ARCHITECTURE.md          ← Architecture guide
├── README.md                ← Project overview
└── bruno/                   ← API collection
```

## 📚 Learning Path

**New to project?**
1. Read [ARCHITECTURE.md](../ARCHITECTURE.md)
2. Read [.claude-skills/README.md](README.md)
3. Try [QUICK_REFERENCE.md](QUICK_REFERENCE.md)
4. Practice with a skill

**Ready to build?**
1. Pick the right skill
2. Craft your prompt
3. Review output
4. Test and iterate

**Want to contribute?**
1. Follow existing skill format
2. Test with AI assistant
3. Update this index
4. Submit PR

---

**Questions?** Check [README.md](README.md) for detailed documentation.
