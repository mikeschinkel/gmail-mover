# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for the Gmail Mover project. ADRs document the significant architectural decisions made during the development of this personal information management system.

## What are ADRs?

Architecture Decision Records capture important architectural decisions along with their context and consequences. They help maintain consistency across the project and provide valuable context for new contributors.

## Current ADRs

### Foundational Decisions

- **[ADR-000: Adopting Architecture Decision Records](adr-000-adopting-architecture-decision-records.md)**  
  *The decision to use ADRs for documenting architectural choices*

- **[ADR-001: Go and Related Best Practices](adr-001-go-and-related-best-practices.md)**  
  *Programming language choice, coding standards, architectural patterns, and tooling decisions*

- **[ADR-002: OAuth Configuration and Token Management Strategy](adr-002-oauth-configuration-token-management.md)**  
  *Single OAuth app approach, XDG-compliant config directory structure, and per-account token isolation*

- **[ADR-003: Email Date Preservation Using Gmail internalDate Parameter](adr-003-email-date-preservation-internal-date.md)**  
  *Technical implementation for preserving original email chronology in transferred messages*

### Core Architecture

- **[ADR-004: SQL Schema and Related Concepts](adr-004-sql-schema-and-related-concerns.md)**  
  *Comprehensive database design principles, message types, storage optimizations, and extensibility decisions*

- **[ADR-005: SQLite Sharding Strategy](adr-005-sqlite-sharding-strategy.md)**  
  *Multi-database approach using `ATTACH DATABASE` for scaling email archives*

- **[ADR-006: Gmail Metadata Storage](adr-006-gmail-metadata-storage.md)**  
  *Google Apps Script with PropertiesService for operational metadata and sync state*

- **[ADR-007: Sync Concurrency and Progress Semantics](adr-007-sync-concurrency.md)**  
  *Ordered Frontier Commit (OFC) pattern for reliable Gmail→SQLite synchronization with failure isolation*

## Contributing

When making significant architectural decisions:

1. **Create a new ADR** following the established format
2. **Number it sequentially** (next available ADR-XXX number)
3. **Update this README** to include the new ADR
4. **Cross-reference** related ADRs where appropriate

## ADR Lifecycle

- **Proposed:** Under discussion
- **Accepted:** Decision made and being implemented
- **Deprecated:** No longer recommended but still in use
- **Superseded:** Replaced by a newer ADR

## Questions?

If you have questions about any architectural decision, check the relevant ADR first. If the ADR doesn't answer your question or seems outdated, please raise an issue for discussion.