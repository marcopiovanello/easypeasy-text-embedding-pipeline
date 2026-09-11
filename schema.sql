create extension if not exists vector;

create table if not exists documents (
    id bigserial primary key,
    document_id uuid not null unique,
    text text,
    embedding vector(1024)
);

create index on documents USING hnsw (embedding vector_cosine_ops);