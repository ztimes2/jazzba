CREATE TABLE notebooks (
   id SERIAL PRIMARY KEY,
   name VARCHAR(200) NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   CONSTRAINT unique_notebook_name UNIQUE(name)
); 