---
status: accepted
date: 2026-03-03
---

# Use Architectural Decision Records

## Context and Problem Statement

Design decisions in this project will accumulate over time. Without a structured
record, the rationale behind past decisions becomes tribal knowledge that is
lost when context shifts or contributors change.

## Decision Drivers

- Decisions should be discoverable alongside the code they affect.
- The format should be lightweight enough that writing an ADR is not a burden.
- Both human readers and LLMs working with the codebase benefit from a
  consistent, structured format.

## Considered Options

- **Architectural Decision Records in the repository** — structured Markdown
  files committed with the code.
- **Wiki or external documentation** — decisions recorded outside the
  repository.
- **No formal record** — rely on commit messages and pull request descriptions.

## Decision Outcome

Chosen option: "Architectural Decision Records in the repository." Keeping ADRs
in the repo ensures they are versioned with the code, reviewable in pull
requests, and available offline.

### Consequences

- Every significant architectural decision gets a dedicated file in `docs/adr/`.
- ADRs are immutable once accepted; superseding decisions reference the
  original.

## More Information

- [ADR GitHub organization](https://adr.github.io/)
