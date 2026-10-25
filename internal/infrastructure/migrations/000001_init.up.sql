begin;

create type order_status as enum ('pending', 'expired', 'completed');

create table products (
    id uuid primary key,
    name text unique not null,
    quantity integer not null
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
    quantity integer not null,
    unique (id, product_id)
);

create table outbox_events (
    id text primary key,
    payload jsonb not null,
    routing_key text not null,
    occurred_at timestamptz not null,
    delivered_at timestamptz
);

commit;
