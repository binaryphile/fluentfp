2026-01-03T18:58:37Z | Completion: Named vs Inline Functions Documentation

# Completion: Named vs Inline Functions Documentation

**Completed:** 2026-01-03

## Summary
Documented discriminants for when to name anonymous functions vs use them inline. Updated CLAUDE.md and all READMEs with guidance and consistent examples.

## Deliverables
- CLAUDE.md: New "Named vs Inline Functions" section with preference hierarchy, discriminants table, locality principle
- README.md, slice/README.md: Updated all examples with godoc-style comments

## Key Decisions
- Preference hierarchy: method expressions > named functions > inline anonymous
- Godoc-style comments required on named anonymous functions
- Locality: define near usage, not at package level

## Commit
4a7bd05 - Document named vs inline function guidelines
