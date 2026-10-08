CREATE TABLE products (
    id          BIGINT PRIMARY KEY,
    name        TEXT    NOT NULL,
    price       INTEGER NOT NULL CHECK (price >= 0),
    stock       INTEGER NOT NULL CHECK (stock >= 0)
);

INSERT INTO products (id, name, price, stock) VALUES
    (1, 'コーヒー豆 200g', 1200, 10),
    (2, 'ドリップバッグ 10個入り', 980, 0);
