CREATE TABLE note_tags (
   id SERIAL PRIMARY KEY,
   tag_name VARCHAR(100) NOT NULL,
   note_id INT NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   CONSTRAINT unique_note_tag UNIQUE(tag_name, note_id),
   CONSTRAINT fk_note_tag_note_id FOREIGN KEY(note_id) REFERENCES notes(id) ON DELETE CASCADE
);