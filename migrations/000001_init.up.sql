
CREATE TABLE IF NOT EXISTS expenses (
  id SERIAL PRIMARY KEY,
  title TEXT,
  amount FLOAT,
  note TEXT,
  tags TEXT[]
);

INSERT INTO "expenses" ("id","title", "note","amount", "tags") VALUES ( 10,'test-title', 'test-note', 15, '{test-tags}');
