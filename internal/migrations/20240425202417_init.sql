-- +goose Up
-- +goose StatementBegin

CREATE TABLE people (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    surname TEXT NOT NULL,
    patronymic TEXT
);

CREATE TABLE cars (
    car_id BIGSERIAL PRIMARY KEY,
    reg_num TEXT NOT NULL,
    mark TEXT NOT NULL,
    model TEXT NOT NULL,
    year INTEGER,
    owner_id BIGINT REFERENCES people(id),
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE cars;
DROP TABLE people;

-- +goose StatementEnd
