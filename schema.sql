create extension if not exists vector;

create table if not exists documents (
    id bigserial primary key,
    document_id uuid not null unique,
    text text,
    embedding vector(1024)
);