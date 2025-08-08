# Gmail Mover - SQL Schema and Related Concepts

## Overview

This Architecture Decision Record (ADR) documents the key architectural decisions made during the design of a
comprehensive personal messaging archive system, initially focused on Gmail but extensible to other platforms. The
system is designed for space efficiency, historical accuracy, and cross-platform extensibility.

## Core Design Principles

### 1. **Space Efficiency as Primary Concern**

- **Decision**: Design is optimized for space efficiency as a primary concern while maintaining best practices
- **Rationale**: Database will become massive once in production use
- **Implementation**:
    - Normalize email addresses and domains
    - Use CHAR(1) codes for frequently used categorical data
    - Compress raw message content when beneficial

### 2. **INTEGER-Only Foreign Keys**

- **Decision**: All Primary Keys and Foreign Keys must be INTEGER type
- **Rationale**: The "cool" feature you think is great today is often the one you hate later; INTEGER PKs/FKs are
  robust, fast, and avoid painting into corners
- **Rejected**: CHAR(1) foreign keys to messaging_type tables

## Database Schema Architecture

### Message Types and Classification

#### **Core Message Types**

- **Decision**: Five fundamental message types using CHAR(1) codes:
    - `E` - Email Message
    - `D` - Direct Message
    - `P` - Shared Post
    - `G` - Group Chat
    - `L` - Mailing List Message

- **Rationale**: These represent the fundamental communication paradigms that exist across all platforms
- **Note**: "Shared Post" was chosen over "Public Post" to avoid confusion between truly public content versus content
  that is community-visible (e.g., visible within a company Slack workspace)
- **Implementation**: Virtual columns provide human-readable names without storage overhead

#### **Virtual Columns for Readability**

```sql
type_name TEXT GENERATED ALWAYS AS (
  CASE type
    WHEN 'E' THEN 'Email Message'
    WHEN 'D' THEN 'Direct Message'
    WHEN 'P' THEN 'Shared Post'
    WHEN 'G' THEN 'Group Chat'
    WHEN 'L' THEN 'List Message'
    ELSE 'Unspecified'
  END
) VIRTUAL,

type_short_name TEXT GENERATED ALWAYS AS (
  CASE type
    WHEN 'E' THEN 'Email'
    WHEN 'D' THEN 'DM'
    WHEN 'P' THEN 'Post'
    WHEN 'G' THEN 'Chat'
    WHEN 'L' THEN 'List'
    ELSE '???'
  END
) VIRTUAL,

date TEXT GENERATED ALWAYS AS (
  STRFTIME('%Y-%m-%d %H:%M', unix_date, 'unixepoch', 'localtime')
) VIRTUAL
```

### Platform and Identity Management

#### **Platform Architecture**

- **Decision**: Single `platforms` table containing service/brand names
- **Examples**: 'Gmail', 'Outlook.com', 'Twitter.com', 'X.com', 'Slack'
- **Rationale**: Simplified from complex multi-table approach while maintaining historical name tracking

#### **Platform History Tracking**

- **Decision**: `platform_history` table tracks name changes over time
- **Implementation**:
    - Links to conceptual platform via `platform_id`
    - Stores historical names with date ranges
- **Use Case**: Twitter.com → X.com transitions while maintaining ability to query
- **Long-term**: Platform history data may be provided via GitHub repository for community maintenance
- **Question**: Should we use an `is_current` flag or determine current status by the most recent history entry?

#### **Participant Identity System**

- **Decision**: Renamed from "identity" to "participant" for semantic clarity
- **Rationale**: "Participant" better describes entities involved in message exchanges
- **Structure**:
    - `participants` table with `platform_id`, `username`, `authority_id`
    - `authorities` table for email domains and similar qualifiers
    - Supports john.doe@gmail.com as username='john.doe', authority='gmail.com'

### Message Content Strategy

#### **Dual Content Storage**

- **Decision**: Store both processed and raw content
- **Implementation**:
    - `content` (TEXT NOT NULL): Human-readable, searchable version
    - `raw_content` (BLOB): Original compressed message
- **Rationale**:
    - Small messages (tweets): `content` contains full message, `raw_content` may be NULL
    - Large messages (emails): `content` contains cleaned excerpt, `raw_content` stores full original

#### **Message ID Strategy**

- **Decision**: Every message gets a `message_id`, generated if not present
- **Platform-Specific Approach**:
    - Emails: RFC822 Message-ID or generated UUID@domain
    - Social platforms: Native ID (e.g., Twitter's tweet ID: 1951349959200219313)
- **Tracking**: `id_generated` CHAR(1) flag ('Y'/'N') indicates if ID was manufactured

#### **Date Storage**

- **Decision**: Store as `unix_date` INTEGER with virtual `date` column
- **Format**: Virtual column generates 'YYYY-MM-DD HH:MM' format
- **Rationale**: Space efficient storage with human-readable display

### Header Management

#### **Selective Header Extraction**

- **Decision**: Parse essential headers into dedicated message fields, extract useful headers to separate table, leave
  remainder in raw_content
- **Message Fields** (parsed into `messages` table columns):
    - From/To/Subject → parsed into `from_id`, recipient relationships, and `subject` fields
    - Content-Type → normalized to `content_types` table via FK
- **Extracted Headers** (stored in `message_headers` table):
    - Received, Return-Path, X-Priority, Importance, Delivered-To
- **Rationale**: Balance between query performance and storage efficiency

#### **Content Type Normalization**

- **Decision**: `content_types` table with FK from messages
- **Examples**: 'text/plain', 'text/html', 'multipart/alternative'
- **Rationale**: Significant space savings over repeating MIME type strings

### Participant Roles and Relationships

#### **RFC822-Aligned Role System**

- **Decision**: Use RFC822 terminology for participant roles:
    - `F` - From (primary author)
    - `S` - Sender (transmission agent, if different from From)
    - `T` - To (primary recipient)
    - `C` - Cc (carbon copy)
    - `B` - Bcc (blind carbon copy)
    - `R` - Reply-To (reply address)
    - `A` - Agent (system/daemon)
    - `L` - List (mailing list address)

- **Rationale**: Eliminates mental gymnastics by using standard email terminology
- **Benefit**: Clear semantic meaning without translation layer

#### **Fast From Access**

- **Decision**: `from_id` INTEGER NULL in messages table pointing to primary From: participant
- **Rationale**: Avoids JOIN for most common query (getting message sender)
- **Implementation**: This creates intentional redundancy - the From: participant appears both as `from_id` in the
  `messages` table and as a participant with role 'F' in the `message_participants` table
- **Trade-off**: Accepted minimal redundancy for significant performance benefit on primary access pattern

### Tagging and Classification

#### **Hierarchical Tag System**

- **Decision**: Single `tags` table with self-referencing `parent_id`
- **Tag Types**:
    - `G` - Grouping/Category
    - `L` - Leaf
    - `R` - Regular standalone
    - `A` - Alias
    - `S` - Stemming
- **Features**:
    - Supports stemming (retailer ← retailers)
    - Supports aliases (Amazon ← AMZN)
    - Single-level hierarchy constraint to avoid complexity
    - Grouping and Leaf types are paired (e.g., `retailers/amazon`)
        - AI-assisted classification with user feedback
- **Design Philosophy**: Tags are personal to the user and designed for daily interaction with the system rather than
  archival purposes
- **Future Enhancement**: Tag rules for automatic application based on AI-developed, user-guided rules

### Storage Optimizations

#### **Compression Strategy**

- **Decision**: `compression` CHAR(1) NULL field
- **Values**: 'G' (gzip), 'Z' (zstd), '0' (no compression)
- **Rationale**: Explicit compression indication for blob handling

#### **Boolean as CHAR(1) with Explicit Values Over NULL**

- **Decision**: Use explicit values instead of NULL for boolean-like flags (e.g., 'Y'/'N' for Yes/No)
- **Rationale**: Avoid NULL three-valued logic bugs, enable simple WHERE clauses
- **Example**: `WHERE id_generated <> 'Y'` works without NULL handling

#### **Single Underscore Preference**

- **Decision**: Avoid multiple underscores in field names (prefer `platform_history` over `platform_name_version`)
- **Rationale**: Reduces visual noise and improves readability

## Extensibility Decisions

### **Multi-Platform Support**

- **Current**: Gmail API focus
- **Future**: Support for Slack, Teams, Facebook, Twitter, Mastodon, Discord, etc.
- **Architecture**: Platform-agnostic message types with platform-specific `platform_id` and parsing

### **Additional Data Types**

- **Planned Extensions**:
    - Documents (Google Docs, Office files)
    - Browser data (tabs, bookmarks, history)
    - Financial transactions
    - Social media interactions
- **Goal**: 360-degree personal information management

### **Database Backend**

- **Current**: SQLite for development simplicity
- **Future**: PostgreSQL support planned
- **Constraint**: Schema must remain compatible with both systems

## Data Ownership and Privacy

### **User-Controlled Data**

- **Decision**: Users own their SQLite databases
- **Deployment Options**:
    - Hosted SaaS (convenience)
    - Self-hosted (control)
    - Data export capabilities
- **Rationale**: Avoid vendor lock-in, provide data portability

### **Gmail Integration Approach**

- **Decision**: Augment Gmail rather than replace it
- **Method**: Gmail API access with both read and write capabilities
- **Write Operations**: Add labels for categorization, move messages between accounts, delete archived messages
- **Benefit**: No switching friction, maintains user comfort

## Technical Implementation Notes

### **Content Processing Strategy**

- **Decision**: Strip HTML, CSS, quoted text, and signatures from content before storage in `content` field
- **Rationale**: Minimize storage requirements for indexable, searchable, and AI-processable content
- **Implementation**: `content` field contains cleaned, plain-text version while `raw_content` preserves original

### **Date Range Processing**

- **Decision**: Process Gmail archives in monthly chunks
- **Query Pattern**: `after:2020/01/31 before:2020/03/01`
- **Rationale**: Manageable API rate limits and error recovery

### **Deduplication Strategy**

- **Decision**: Use RFC822 Message-ID for cross-account deduplication
- **Fallback**: Generate synthetic IDs for malformed messages
- **Tracking**: `id_generated` flag indicates synthetic vs. original IDs

## Performance Considerations

### **Deduplication Strategy**

- **Decision**: Use RFC822 Message-ID for cross-account deduplication
- **Fallback**: Generate synthetic IDs for malformed messages
- **Tracking**: `id_generated` flag indicates synthetic vs. original IDs

## Quality Assurance

### **Data Integrity**

- **Message ID Uniqueness**: Global uniqueness constraint
- **Participant Uniqueness**: `(platform_id, username, authority_id)` composite key
- **Header Relationship**: Many-to-many between messages and headers (no unique constraint)

### **Error Handling**

- **Malformed Messages**: Generate synthetic IDs, flag as generated
- **Missing Participants**: Allow NULL `from_id` for edge cases

## Decision Rationale Summary

The architecture prioritizes:

1. **Long-term Maintainability**: INTEGER PKs/FKs, clear naming conventions, `platform_history` for data validity over
   long periods
2. **Space Efficiency**: Aggressive normalization, CHAR(1) codes, compression
3. **Query Performance**: Strategic denormalization (from_id), virtual columns
4. **Extensibility**: Platform-agnostic message types, flexible participant system
5. **Data Integrity**: Strong constraints where appropriate, NULL handling for edge cases
6. **User Control**: Data ownership, export capabilities, augmentation vs. replacement

## Future Considerations

### **Community Data Repositories**

- **Platform History**: Maintain current platform name mappings via GitHub repository
- **Tag Definitions**: Provide globally accepted starting point for new users
- **Tag Rules**: Community-contributed rules for automatic message categorization (e.g., "recognize Amazon order
  confirmation and tag appropriately")
- **Implementation**: Users can contribute and adopt community rules with AI assistance

These decisions collectively create a system that can scale to massive personal archives while remaining maintainable
and extensible across multiple communication platforms.