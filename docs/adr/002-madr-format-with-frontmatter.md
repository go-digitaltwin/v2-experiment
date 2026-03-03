---
status: accepted
date: 2026-03-03
---

# Use MADR 4.0.0 with YAML Frontmatter

## Context and Problem Statement

Having decided to use ADRs
([ADR-001](001-use-architectural-decision-records.md)), we need a consistent
template. The template should be structured enough to guide authors yet flexible
enough to omit sections that add no value for straightforward decisions.

## Decision Drivers

- A widely adopted format reduces the learning curve for contributors.
- YAML frontmatter enables tooling (status dashboards, search, LLM parsing)
  without custom parsers.
- Optional sections keep simple ADRs concise.

## Considered Options

- **MADR 4.0.0 with YAML frontmatter** — the latest version of the Markdown Any
  Decision Records template, extended with machine-readable metadata.
- **MADR 4.0.0 without frontmatter** — same template, status encoded only in the
  heading or body text.
- **Nygard format** — the original short-form ADR template.

## Decision Outcome

Chosen option: "MADR 4.0.0 with YAML frontmatter." The structured frontmatter
(`status`, `date`) makes ADRs queryable by tools while the MADR body provides a
familiar, well-documented structure.

Optional MADR sections (Pros and Cons per option, Confirmation, More
Information) are included only when they add value.

### Consequences

- Every ADR begins with a YAML frontmatter block containing at least `status`
  and `date`.
- The body follows the MADR 4.0.0 template.

## More Information

- [MADR 4.0.0][madr]

[madr]: https://adr.github.io/madr/
