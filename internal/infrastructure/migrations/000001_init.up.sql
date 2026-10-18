begin;

create type order_status as enum ('pending_payment', 'expired', 'completed');

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
    constraint unique order_items_product (id, product_id)
);

commit;