# Library Management System - Design Approach

## Step-by-Step Thinking Process

### Phase 1: Identify Core Entities (Nouns)
From requirements, extract nouns:
- **Book** (metadata: title, ISBN, authors, subject, publication date)
- **BookItem** (physical copy with barcode, rack location, status)
- **Member** (person who borrows books)
- **Rack** (physical location)
- **Author** (book creator)
- **Reservation** (booking for unavailable books)
- **Lending** (checkout transaction)
- **Fine** (penalty for late returns)

### Phase 2: Clarify Confusions

#### Q: Should Book have search methods?
**A: NO.** Book is a data entity. Search is a Library/Catalog responsibility.
- Book = "What I am" (data)
- Library/Catalog = "What can be done with books" (operations)

#### Q: Should Book know which Rack it's in?
**A: NO.** Book is metadata (title, ISBN). BookItem (physical copy) knows its rack.
- Book = Abstract concept (1 book, many copies)
- BookItem = Physical copy (has barcode, rack location, availability)

#### Q: Should Rack have Book field or Book have Rack field?
**A: BookItem has Rack reference.** Rack is just a location identifier.
- Rack doesn't "own" books
- BookItem is "located at" a Rack

### Phase 3: Assign Responsibilities

| Responsibility | Owner | Why? |
|---------------|-------|------|
| Search books | Library/Catalog | Knows all books |
| Check availability | BookItem | Knows its status |
| Checkout book | Library | Orchestrates transaction |
| Track borrowed books | Member | Knows personal history |
| Calculate fine | Lending | Knows due date & return date |
| Reserve book | Library | Manages reservations |
| Send notifications | NotificationService | Single responsibility |

### Phase 4: Key Design Decisions

1. **Book vs BookItem Separation**
   - Book = Metadata (1 record)
   - BookItem = Physical copies (N records)
   - Why? Multiple copies of same book

2. **Search Strategy**
   - Use Strategy Pattern for different search criteria
   - Library delegates to SearchService/Catalog

3. **Status Management**
   - BookItem has status: AVAILABLE, CHECKED_OUT, RESERVED, LOST
   - Use enum for type safety

4. **Fine Calculation**
   - Lending knows checkout date, due date
   - Fine = days_overdue × daily_rate
   - Configurator holds fine rate

5. **Reservation Queue**
   - FIFO queue per Book
   - Notify first in queue when available

## Design Patterns Used

- **Strategy Pattern**: Search strategies
- **Observer Pattern**: Notifications
- **Factory Pattern**: Creating lendings, reservations
- **Repository Pattern**: Data access for books, members
- **Service Layer**: Business logic separation

## Class Relationships

```
Library (1) -----> (*) Book
Book (1) -----> (*) BookItem
BookItem (*) -----> (1) Rack
Book (*) -----> (*) Author
Member (1) -----> (*) Lending
BookItem (1) -----> (*) Lending
BookItem (1) -----> (*) Reservation
Member (1) -----> (*) Reservation
```
