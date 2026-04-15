# Consistency in ACID

Consistency means: a transaction takes the database from one valid state to another valid state. "Valid" means all defined rules (constraints, triggers, cascades) are satisfied. If any rule would be violated, the transaction is rejected entirely.

Think of it as a bouncer at the door — every write has to pass all the rules before it's allowed in.

---

## How Unique Constraints Work Internally

You're on the right track thinking about a "map," but it's more nuanced. Databases don't use a simple hash map in memory. They use a B+ Tree index.

When you declare:

```sql
CREATE TABLE authors (
    id INT PRIMARY KEY,
    email VARCHAR(255) UNIQUE
);
```

The DB creates a B+ Tree index on the `email` column. Here's what happens on insert:

```
INSERT INTO authors (id, email) VALUES (1, 'alice@example.com');
-- DB traverses the B+ Tree index on `email`
-- Key not found → insert succeeds
-- New entry added to the B+ Tree

INSERT INTO authors (id, email) VALUES (2, 'alice@example.com');
-- DB traverses the same B+ Tree index
-- Key FOUND → constraint violation → transaction rejected
```

Why a B+ Tree and not a hash map?

```
Hash Map:
  - O(1) lookup, but...
  - Can't do range queries
  - Doesn't survive restarts without extra work
  - Hard to keep sorted on disk

B+ Tree:
  - O(log n) lookup — still very fast
  - Sorted on disk — great for range scans
  - Naturally page-aligned — fits how disks work
  - Survives restarts (it IS the on-disk structure)

Visually, the B+ Tree index on `email` looks like:

              [jane@...]
             /          \
    [alice@..., bob@...]  [jane@..., zara@...]
         |        |            |          |
       leaf     leaf         leaf       leaf
      (row ptr) (row ptr)  (row ptr)  (row ptr)
```

Each leaf node points back to the actual row on disk. So the "uniqueness check" is really just a B+ Tree lookup — if the key already exists in the tree, reject.

Some databases (like PostgreSQL) also support hash indexes, but B+ Trees are the default because they're more versatile.

---

## How Cascading Deletes Work Internally

You're right to think about a tree of relationships, but the DB doesn't maintain a single global "relationship tree." Instead, it uses the system catalog (metadata tables) + foreign key indexes.

Here's the setup:

```sql
CREATE TABLE authors (
    id INT PRIMARY KEY,
    name VARCHAR(100)
);

CREATE TABLE books (
    id INT PRIMARY KEY,
    title VARCHAR(200),
    author_id INT REFERENCES authors(id) ON DELETE CASCADE
);

CREATE TABLE reviews (
    id INT PRIMARY KEY,
    content TEXT,
    book_id INT REFERENCES books(id) ON DELETE CASCADE
);
```

The relationship chain is: `authors → books → reviews`

### What the DB stores internally

The system catalog (think: internal metadata tables) records:

```
Foreign Key Catalog (simplified):
┌─────────────┬──────────────┬───────────────┬────────────┐
│ child_table  │ child_column │ parent_table  │ on_delete  │
├─────────────┼──────────────┼───────────────┼────────────┤
│ books        │ author_id    │ authors       │ CASCADE    │
│ reviews      │ book_id      │ books         │ CASCADE    │
└─────────────┴──────────────┴───────────────┴────────────┘
```

This IS effectively a dependency graph (not a tree, because a table can have multiple foreign keys to different parents):

```
    authors
       |
       ▼
     books
       |
       ▼
    reviews
```

### What happens on DELETE

```sql
DELETE FROM authors WHERE id = 42;
```

Step by step:

```
1. DB looks up system catalog: "who references authors.id?"
   → books.author_id (ON DELETE CASCADE)

2. DB executes: DELETE FROM books WHERE author_id = 42
   But BEFORE deleting, it checks the catalog again:
   "who references books.id?"
   → reviews.book_id (ON DELETE CASCADE)

3. DB executes: DELETE FROM reviews WHERE book_id IN (
       SELECT id FROM books WHERE author_id = 42
   )

4. No more dependents → reviews rows deleted
5. books rows deleted
6. authors row deleted
7. Transaction commits — all or nothing
```

The cascade is essentially a recursive/DFS walk through the dependency graph stored in the system catalog.

### How the child lookups are fast

The foreign key column (`author_id` in books) typically has a B+ Tree index too:

```
B+ Tree index on books.author_id:

         [50]
        /    \
   [10, 42]   [50, 73]
     |    |      |    |
   rows  rows  rows  rows
   with  with  with  with
   a=10  a=42  a=50  a=73
```

So finding "all books where author_id = 42" is a fast O(log n) index lookup, not a full table scan.

---

## Putting It All Together — A Concrete Example

```
State BEFORE:
  authors: [(id=42, name='Tolkien')]
  books:   [(id=1, title='The Hobbit', author_id=42),
            (id=2, title='LOTR', author_id=42)]
  reviews: [(id=10, content='Amazing', book_id=1),
            (id=11, content='Epic', book_id=2)]

Execute: DELETE FROM authors WHERE id = 42;

Cascade walk:
  authors(42) → books(1), books(2) → reviews(10), reviews(11)

State AFTER (if committed):
  authors: []
  books:   []
  reviews: []

If ANY step fails → entire transaction rolls back → state unchanged
```

---

## Summary

| Mechanism | Internal Structure | How It Enforces Consistency |
|---|---|---|
| Unique constraint | B+ Tree index on the column | Lookup before insert; reject if key exists |
| Primary key | B+ Tree index (clustered) | Same as unique + not null |
| Foreign key check | B+ Tree index on FK column + system catalog | Lookup parent before child insert; lookup children before parent delete |
| Cascade delete | System catalog (dependency graph) + recursive walk | DFS through FK graph, deleting dependents bottom-up |

The key insight: consistency isn't magic. It's enforced by well-chosen data structures (B+ Trees for fast lookups) and metadata (system catalog for relationship tracking), all wrapped in a transaction so it's atomic — the state is either fully valid or unchanged.
