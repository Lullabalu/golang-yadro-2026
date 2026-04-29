CREATE TABLE comic_words (
  comic_id INT REFERENCES comics(id),
  word_id INT REFERENCES words(id),
  PRIMARY KEY (comic_id, word_id)
);