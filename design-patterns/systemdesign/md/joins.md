# SQL Joins

## Tables

**Table A:** `{1, 1, 2, 3, null}`
**Table B:** `{1, 1, 2, 3, null}`

(joining on `A.val = B.val`)

## INNER JOIN

Returns only rows where there's a match in both tables. NULLs never match (NULL = NULL is false in SQL).

```
A.val | B.val
------+------
  1   |   1
  1   |   1
  1   |   1
  1   |   1
  2   |   2
  3   |   3
```

The two 1s in A each match the two 1s in B → 2×2 = 4 rows. NULL is excluded.

## LEFT JOIN (LEFT OUTER JOIN)

All rows from A, matched rows from B (NULL where no match).

```
A.val | B.val
------+------
  1   |   1
  1   |   1
  1   |   1
  1   |   1
  2   |   2
  3   |   3
 null | null   ← A's null, no match in B
```

The null in A appears but B side is NULL because `null = null` is false.

## RIGHT JOIN (RIGHT OUTER JOIN)

All rows from B, matched rows from A (NULL where no match).

```
A.val | B.val
------+------
  1   |   1
  1   |   1
  1   |   1
  1   |   1
  2   |   2
  3   |   3
 null | null   ← B's null, no match in A
```

Mirror image of the left join in this case since the tables are identical.

## FULL OUTER JOIN

All rows from both sides. Unmatched rows get NULL on the other side.

```
A.val | B.val
------+------
  1   |   1
  1   |   1
  1   |   1
  1   |   1
  2   |   2
  3   |   3
 null | null   ← A's null (no match)
 null | null   ← B's null (no match)
```

Both unmatched nulls appear as separate rows.

## CROSS JOIN

Cartesian product — every row in A paired with every row in B. No join condition.

```
5 × 5 = 25 rows total
```

Every combination, including nulls paired with everything.

## Key Takeaway

`NULL = NULL` evaluates to `UNKNOWN` in SQL, not `TRUE`. So NULLs never match in any equality-based join condition. If you need NULLs to match, use `IS NOT DISTINCT FROM` (Postgres) or something like `COALESCE`.


## When to Use Which Join

### Use INNER JOIN when

You only care about rows that have matching data on both sides. If there's no match, you don't want to see it at all.

Example: "Show me all orders with their customer details" — if an order somehow has no customer, you probably don't want it in the result.

```sql
SELECT * FROM orders o
INNER JOIN customers c ON o.customer_id = c.id
```

### Use LEFT JOIN when

You want all rows from the "main" (left) table, even if there's no match in the other table. The unmatched rows get NULLs on the right side.

Example: "Show me all customers and their orders, including customers who haven't ordered anything yet."

```sql
SELECT * FROM customers c
LEFT JOIN orders o ON c.id = o.customer_id
```

Also useful for finding rows with no match (anti-join pattern):

```sql
SELECT * FROM customers c
LEFT JOIN orders o ON c.id = o.customer_id
WHERE o.id IS NULL   -- customers with zero orders
```

### Use RIGHT JOIN when

Same idea as LEFT JOIN but you want all rows from the right table preserved. In practice, most people just swap the table order and use LEFT JOIN instead — it reads more naturally.

```sql
-- These two are equivalent:
SELECT * FROM orders o RIGHT JOIN customers c ON o.customer_id = c.id
SELECT * FROM customers c LEFT JOIN orders o ON c.id = o.customer_id
```

### Use FULL OUTER JOIN when

You want everything from both tables, matched where possible, NULLs where not. This is less common but useful for reconciliation or diffing.

Example: "Compare data between two systems — show matches, things only in system A, and things only in system B."

```sql
SELECT * FROM system_a a
FULL OUTER JOIN system_b b ON a.key = b.key
```

### Quick Decision Guide

```
Do you need unmatched rows?
├── No  → INNER JOIN
└── Yes
    ├── From the left table only?  → LEFT JOIN
    ├── From the right table only? → RIGHT JOIN (or swap + LEFT JOIN)
    └── From both tables?          → FULL OUTER JOIN
```

### Rule of Thumb

- Start with INNER JOIN as the default — it's the most common and the safest (no surprise NULLs).
- Reach for LEFT JOIN when you need to preserve all rows from your "primary" table.
- RIGHT JOIN is rarely used in practice — just reorder your tables and use LEFT JOIN.
- FULL OUTER JOIN is niche — data migration, reconciliation, or when both sides can have unmatched rows.
