# README Writing Guide

Patterns for writing Go package documentation. Distilled from README iteration.

## 1. Common Structure

Standard section order for package READMEs:

1. **Title + tagline** — One line: `# package: what it does`
2. **Key concept definition** — Define terminology early (e.g., "A connection is either **open** or **closed**")
3. **Quick example** — Single code block showing primary use case
4. **Links** — pkg.go.dev, related docs (optional for small packages)
5. **Quick Start** — Import + 2-3 examples
6. **Types** — If package has multiple types (optional)
7. **API Reference** — Tables with Function | Signature | Purpose | Example
8. **Patterns** — Common usage patterns (for larger packages)
9. **When NOT to Use** — Explicitly state limitations
10. **See Also** — Links to related packages/docs

## 2. API Table Format

```markdown
| Function | Signature | Purpose | Example |
|----------|-----------|---------|---------|
| `New` | `New(name string) *Client` | Create client | `client = pkg.New("api")` |
```

- Example column shows full chain with assignment
- Use `=` not `:=` (except comma-ok pattern which needs `:=`)
- Show realistic variable names, not `x` or `val`

## 3. Variable Naming in Examples

| Context | Convention | Example |
|---------|------------|---------|
| Wrapper types | Suffix with type name | `httpClient`, `dbConn` |
| Pointers as optionals | Suffix with `Opt` | `configOpt`, `loggerOpt` |
| Collections | Natural plurals | `users`, `records`, `items` |

Results should match variable names:
```go
name := user.Name()  // not: val := user.Name()
```

## 4. Tone

State facts. Show code. Trust the reader.

**Avoid:**
- "It is useful for..." — show the use, don't explain it
- "Unfortunately..." — state the limitation directly
- "Simply..." — if it were simple, you wouldn't need to say so

**Prefer:**
- Direct statements: "Client manages API connections."
- Code examples over prose explanations
- Tables for comparisons and API references

## 5. Scaling by Package Size

| Size | Line Count | Sections |
|------|------------|----------|
| Full | 150-200 | All sections |
| Medium | 100-140 | Skip Patterns, shorter Quick Start |
| Minimal | ~40 | Title, Quick Start, API, When NOT, See Also |

Don't pad small packages. A 40-line README is better than a 100-line README with filler.

## 6. Canonical Examples

When establishing patterns for a project, designate one full-size and one minimal README as references for contributors.

## 7. When NOT to Use

Every README should explicitly state when NOT to use the package. This:
- Sets correct expectations
- Prevents misuse
- Shows the author understands tradeoffs
