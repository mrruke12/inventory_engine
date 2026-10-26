begin;

create type order_status as enum ('pending', 'released', 'expired', 'canceled', 'completed');

create table products (
    id uuid primary key,
    name text unique not null,
    quantity integer not null check (quantity >= 0)
);

create table orders (
    id uuid primary key,
    status order_status not null,
    expires_at timestamptz not null default now() + interval '5 minutes',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table order_items (
    id uuid primary key,
    order_id uuid not null references orders(id),
    product_id uuid not null references products(id),
    quantity integer not null check (quantity > 0),
    unique (order_id, product_id)
);

create table idempotency_keys (
    key text not null,
    scope text not null,
    request_hash bytea not null,
    result jsonb not null,
    created_at timestamptz not null default now(),
    primary key (key, scope)
);

commit;
