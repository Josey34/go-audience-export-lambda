CREATE TABLE product (
    id    SERIAL PRIMARY KEY,
    name  VARCHAR NOT NULL,
    price NUMERIC(10,2) NOT NULL
);

CREATE TABLE business (
    id            SERIAL PRIMARY KEY,
    retailer_code VARCHAR NOT NULL,
    retailer_name VARCHAR NOT NULL
);

CREATE TABLE audience (
    id               SERIAL PRIMARY KEY,
    project_id       VARCHAR NOT NULL,
    product_id       INT NOT NULL REFERENCES product(id),
    business_id      INT NOT NULL REFERENCES business(id),
    product_flagging VARCHAR NOT NULL
);

CREATE TABLE export_jobs (
    job_id        UUID PRIMARY KEY,
    project_id    VARCHAR NOT NULL,
    status        VARCHAR NOT NULL,
    s3_key        VARCHAR,
    row_count     INT,
    error_message TEXT,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at   TIMESTAMPTZ
);

CREATE INDEX idx_audience_project ON audience(project_id);

INSERT INTO product (id, name, price) VALUES
    (1, 'Widget Pro',   19.99),
    (2, 'Gadget Plus',  49.99),
    (3, 'Super Thing',  9.99);

INSERT INTO business (id, retailer_code, retailer_name) VALUES
    (1, 'RET001', 'Acme Corp'),
    (2, 'RET002', 'Globex Ltd'),
    (3, 'RET003', 'Initech Inc');

INSERT INTO audience (project_id, product_id, business_id, product_flagging) VALUES
    ('p-123', 1, 1, 'active'),
    ('p-123', 2, 1, 'active'),
    ('p-123', 3, 2, 'inactive'),
    ('p-123', 1, 3, 'active'),
    ('p-123', 2, 2, 'pending'),
    ('p-123', 3, 3, 'active'),
    ('p-456', 1, 1, 'active'),
    ('p-456', 2, 3, 'inactive'),
    ('p-456', 3, 1, 'active'),
    ('p-456', 1, 2, 'pending');
