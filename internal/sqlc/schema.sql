CREATE TABLE messages
(
   id              INTEGER PRIMARY KEY,
   type            CHAR(1) NOT NULL,              -- E(mail), D(irect message), P(ost), G(roup chat), L(ist message)
   type_name       TEXT GENERATED ALWAYS AS (
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
   platform_id     INTEGER NOT NULL,             -- FK to platforms.id used for sending
   from_id         INTEGER NULL,                 -- FK to participants.id
   unix_date       INTEGER NOT NULL,
   date            TEXT GENERATED ALWAYS AS (STRFTIME('%Y-%m-%dT%H:%M', unix_date, 'localtime')) VIRTUAL,
   subject         TEXT    NULL,
   content         TEXT    NOT NULL,             -- Space optimized version of the message
   message_id      TEXT    NOT NULL,             -- rfc822 Message-ID; might need to be auto-generated
   content_type_id INTEGER NOT NULL,             -- FK to content_types.id
   id_generated    CHAR(1) NOT NULL DEFAULT 'N', -- If 'Y' the message had no message id and we had to generate one
   compression     CHAR(1) NOT NULL DEFAULT '0', -- G(zip), Z(std), or '0' meaning raw_content is empty
   raw_content     BLOB    NULL,                 -- The original compressed rfc822 message
   UNIQUE (message_id, platform_id),
   FOREIGN KEY (platform_id) REFERENCES platforms (id),
   FOREIGN KEY (from_id) REFERENCES participants (id),
   FOREIGN KEY (content_type_id) REFERENCES content_types (id)
)
;

CREATE TABLE message_tags
(
   id         INTEGER PRIMARY KEY,
   message_id INTEGER NOT NULL, -- FK to messages.id
   tag_id     INTEGER NOT NULL, -- FK to tags.id
   UNIQUE (message_id, tag_id), -- A message typically has only one of each tag type
   FOREIGN KEY (message_id) REFERENCES messages (id),
   FOREIGN KEY (tag_id) REFERENCES tags (id)
)
;

CREATE TABLE tags
(
   id        INTEGER PRIMARY KEY,
   tag       TEXT    NOT NULL,
   type      CHAR(1) NOT NULL,      -- G(rouping), L(eaf), R(egular standlone), A(lias), S(temming)
   parent_id INTEGER NULL, -- FK to tags.id
   UNIQUE (tag, parent_id),
   FOREIGN KEY (parent_id) REFERENCES tags (id)
)
;

CREATE TABLE content_types
(
   id   INTEGER PRIMARY KEY,
   type TEXT NOT NULL UNIQUE
)
;

CREATE TABLE headers
(
   id   INTEGER PRIMARY KEY,
   name TEXT NOT NULL UNIQUE
)
;

CREATE TABLE message_headers
(
   id         INTEGER PRIMARY KEY,
   message_id INTEGER NOT NULL, -- FK to messages.id
   header_id  INTEGER NOT NULL, -- FK to headers.id
   value      TEXT    NOT NULL,
   FOREIGN KEY (message_id) REFERENCES messages (id),
   FOREIGN KEY (header_id) REFERENCES headers (id)
)
;

CREATE TABLE message_participants
(
   id               INTEGER PRIMARY KEY,
   message_id       INTEGER NOT NULL,                     -- FK to messages.id
   participant_id   INTEGER NOT NULL,                     -- FK to participants.id
   participant_type CHAR(1) NOT NULL,                     -- F(rom), T(o), S(ender), C(c), B(cc), A(gent), L(ist), R(eply-to)
   UNIQUE (message_id, participant_id, participant_type), -- A specific participant can only play a role once per message (e.g., not two 'To's)
   FOREIGN KEY (message_id) REFERENCES messages (id),
   FOREIGN KEY (participant_id) REFERENCES participants (id)
)
;

CREATE TABLE participants
(
   id           INTEGER PRIMARY KEY,
   platform_id  INTEGER NOT NULL,                -- FK to platforms.id
   username     TEXT    NOT NULL,                -- e.g., 'john.doe', 'U12345678', 'johndoe' (Twitter)
   authority_id INTEGER,                         -- FK to participant_authorities.id (NULLable if no authority)

   UNIQUE (platform_id, username, authority_id), -- Ensures uniqueness of the full participant
   FOREIGN KEY (platform_id) REFERENCES platforms (id),
   FOREIGN KEY (authority_id) REFERENCES participant_authorities (id)
)
;

CREATE TABLE participant_authorities
(
   id          INTEGER PRIMARY KEY,
   platform_id INTEGER NOT NULL, -- Which platform this authority belongs to (e.g., email_domains)
   name        TEXT    NOT NULL, -- The actual string (e.g., 'example.com', '0001' for Discord)

   UNIQUE (platform_id, name),   -- Ensure unique authority per platform
   FOREIGN KEY (platform_id) REFERENCES platforms (id)
)
;

CREATE TABLE authorities
(
   id   INTEGER PRIMARY KEY,
   name TEXT NOT NULL UNIQUE -- e.g., 'gmail.com', 'google.com', 'microsoft.com'
)
;

CREATE TABLE platforms
(
   id   INTEGER PRIMARY KEY,
   name TEXT NOT NULL UNIQUE -- e.g., 'Twitter/X', 'Meta Platforms (Facebook)', 'Instagram', 'LinkedIn'
)
;

CREATE TABLE platform_history
(
   id          INTEGER PRIMARY KEY,
   platform_id INTEGER NOT NULL,        -- FK to platforms.id
   name        TEXT    NOT NULL UNIQUE, -- e.g., 'Twitter.com', 'X.com', 'Facebook.com', 'Instagram.com'
   is_current  BOOLEAN NOT NULL DEFAULT 0,
   start_date  INTEGER NOT NULL,
   end_date    INTEGER,
   UNIQUE (platform_id, name),
   FOREIGN KEY (platform_id) REFERENCES platforms (id)
)
;
