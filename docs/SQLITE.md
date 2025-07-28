# Email Tagging and Search Design for SQLite-Based Archive

This document outlines a schema and design for tagging, normalizing, and searching email content in a SQLite-based email archiving system. It is based on a conversation about implementing both standard and AI-assisted tagging in a portable and performant way.

---

## Goals

* Store emails in full-fidelity with raw MIME content
* Enable full-text search on simplified/plain content
* Support intersectional tag-based filtering
* Normalize tags (e.g., stemming, singular/plural)
* Allow for compound and alias-based tagging
* Support graph-based relationships between tags
* Support multiple email sources, including Gmail, PST, `.emlx`, and `.eml`
* Use simplified body extraction for efficient indexing
* Avoid duplication of email body content to minimize storage bloat

---

## Supported Email Sources

The tool is designed to handle ingestion from multiple formats:

* **Gmail API**: Fetches raw MIME format using `users.messages.get?format=raw`
* **Outlook PST files**:

    * Primary recommendation: [`mooijtech/go-pst`](https://github.com/mooijtech/go-pst) (pure Go)
    * Alternative: `readpst` CLI from `libpst` (external process)
* **Apple Mail `.emlx` files**:

    * Found under `~/Library/Mail/V*/...`
    * Each `.emlx` contains full MIME + Apple metadata trailer
* **Generic `.eml` files**:

    * Already in MIME format

In each case, full MIME is extracted, parsed, and retained.

---

## MIME Storage and Simplification

### `emails` Table

```sql
CREATE TABLE emails (
    id INTEGER PRIMARY KEY,
    message_id TEXT UNIQUE,
    thread_id TEXT,
    subject TEXT,
    date TEXT,
    folder TEXT,
    source TEXT,
    raw_mime BLOB NOT NULL
);
```

* `raw_mime` is the canonical source of truth
* No duplication of full body content

### Simplified Body

Plain text bodies are extracted and simplified before being indexed:

* Strip quoted text (`>` lines, "On \[date], X wrote:")
* Remove signatures (`--\n` blocks or footers)
* Collapse whitespace
* Optionally truncate (e.g. 5000 chars)

This simplified version is stored in a **separate FTS5 table**:

### `emails_fts` Table

```sql
CREATE VIRTUAL TABLE emails_fts USING fts5(
    email_id UNINDEXED,
    body,
    content=''
);
```

This preserves raw MIME while enabling lightweight full-text search.

---

## Tagging System

### `tags` Table

```sql
CREATE TABLE tags (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE,
    normalized_name TEXT
);
```

### `email_tags` Table

```sql
CREATE TABLE email_tags (
    email_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    FOREIGN KEY(email_id) REFERENCES emails(id) ON DELETE CASCADE,
    FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (email_id, tag_id)
);
```

---

## Tag Querying and Indexing

### Intersecting Tags

```sql
SELECT email_id
FROM email_tags et
JOIN tags t ON et.tag_id = t.id
WHERE t.name IN ('amazon', 'retailer', 'online')
GROUP BY et.email_id
HAVING COUNT(DISTINCT t.name) = 3;
```

### Combining Tag + FTS Search

```sql
SELECT e.*
FROM emails e
JOIN emails_fts fts ON fts.email_id = e.id
JOIN email_tags et1 ON e.id = et1.email_id
JOIN tags t1 ON et1.tag_id = t1.id AND t1.name = 'vendor'
JOIN email_tags et2 ON e.id = et2.email_id
JOIN tags t2 ON et2.tag_id = t2.id AND t2.name = 'unpaid'
WHERE fts.body MATCH 'invoice';
```

---

## Tag Normalization

Tags are normalized via:

* Lowercasing
* Optional stemming (e.g., Porter2)
* Singular/plural collapsing

### `normalized_name` is used internally for matching, but `name` is displayed in UI.

---

## Compound Tags, Aliases, and Graph

### `tag_aliases` Table

```sql
CREATE TABLE tag_aliases (
    id INTEGER PRIMARY KEY,
    tag_id INTEGER NOT NULL,
    alias TEXT UNIQUE,
    FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
```

### `tag_relationships` Table

```sql
CREATE TABLE tag_relationships (
    id INTEGER PRIMARY KEY,
    parent_tag_id INTEGER NOT NULL,
    child_tag_id INTEGER NOT NULL,
    kind TEXT CHECK(kind IN ('is_a', 'related', 'alias')),
    FOREIGN KEY(parent_tag_id) REFERENCES tags(id) ON DELETE CASCADE,
    FOREIGN KEY(child_tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
```

---

## Header Extraction

All MIME headers are extracted and stored as individual records:

### `headers` Table

```sql
CREATE TABLE headers (
    id INTEGER PRIMARY KEY,
    email_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY(email_id) REFERENCES emails(id) ON DELETE CASCADE
);
```

This enables querying on arbitrary headers, including `List-ID`, `X-Mailer`, etc.

---

## Additional Considerations

* Tags may be user-specific or shared (`user_id` column)
* AI can generate tags automatically based on content and sender
* Tag quality can be managed via confidence scores
* Periodic tag review can be implemented for curation
* FTS input text may be pre-processed to remove boilerplate or noise
* Plaintext simplification is not stored — only indexed

---

## Summary

This system enables a high-fidelity, extensible, AI-friendly email archive:

* MIME remains canonical
* Full-text search is optimized and decluttered
* Tags are normalized, relational, and queryable
* Compound and semantic tagging supported
* All of this runs in portable, embeddable SQLite with FTS5

Future extensions may include:

* Contact normalization
* Fuzzy search
* Email threading
* Scheduled AI tag reviews

This design is suitable for personal and developer-focused archiving, or as a foundation for broader tools.
