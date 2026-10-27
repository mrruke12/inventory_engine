# PG locks
There're 2 main entities:
1. orders
2. products

Thus, we must always lock them **in the following order** to avoid deadlocks:
1. Lock on orders
2. Lock on products
