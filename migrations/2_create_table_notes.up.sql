CREATE TABLE notes (
   id SERIAL PRIMARY KEY,
   name VARCHAR(200) NOT NULL,
   content TEXT,
   notebook_id INT NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   CONSTRAINT unique_note_name UNIQUE(name),
   CONSTRAINT fk_note_notebook_id FOREIGN KEY(notebook_id) REFERENCES notebooks(id) ON DELETE CASCADE
);