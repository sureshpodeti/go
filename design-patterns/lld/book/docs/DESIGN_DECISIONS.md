# Book System - Low Level Design

## Problem Statement

Design a book reading system where:
- A book has multiple pages
- Users can open, close, and flip through pages
- Users can read page content
- Books have authors and metadata (ISBN, title, publication year)
- Support navigation: next page, previous page, go to specific page

## Class Diagram

```
┌─────────────────────────┐
│        Author           │
│─────────────────────────│
│ - id: string            │
│ - name: string          │
│─────────────────────────│
│ + getId(): string       │
│ + getName(): string     │
└─────────────────────────┘
            ▲
            │
            │ *
            │
            │ 1
┌───────────┴─────────────────────┐
│           Book                  │
│─────────────────────────────────│
│ - isbn: string                  │
│ - title: string                 │
│ - publicationYear: int          │
│ - authors: []Author             │
│ - pages: []Page                 │
│ - currentPageIdx: int           │
│ - isOpen: bool                  │
│─────────────────────────────────│
│ + getISBN(): string             │
│ + getTitle(): string            │
│ + getPublicationYear(): int     │
│ + getAuthors(): []Author        │
│ + addAuthor(Author): error      │
│ + addPage(Page): error          │
│ + getNumberOfPages(): int       │
│ + open(): error                 │
│ + close(): error                │
│ + isBookOpen(): bool            │
│ + nextPage(): error             │
│ + prevPage(): error             │
│ + goToPage(int): error          │
│ + getCurrentPageNumber(): int   │
│ + getCurrentPage(): Page        │
│ + readCurrentPage(): string     │
│ + hasNextPage(): bool           │
│ + hasPrevPage(): bool           │
└───────────┬─────────────────────┘
            │ 1
            │
            │ *
            ▼
┌─────────────────────────┐
│         Page            │
│─────────────────────────│
│ - number: int           │
│ - content: string       │
│─────────────────────────│
│ + getPageNo(): int      │
│ + getContent(): string  │
└─────────────────────────┘
```

## Critical Design Decisions

### 1. Where Should Navigation Logic Live?

#### ❌ Initial Approach (Rejected)
```go
type Page struct {
    number  int
    content string
    prev    *Page  // ❌ Bad idea
    next    *Page  // ❌ Bad idea
}
```

#### Why This is Wrong

**Violates Single Responsibility Principle**
- Page should only know about its own data (number, content)
- Navigation is a book-level concern, not a page-level concern

**Tight Coupling**
```go
// Page becomes aware of other pages
page1.next = page2
page2.prev = page1
page2.next = page3

// What if we want to reorder pages?
// What if we want to remove a page?
// Every page reference needs updating!
```

**Real-World Analogy**
- Does a physical page know what the next page is? No!
- The book (binding) determines page order
- You flip pages by moving through the book, not by asking the page

#### ✅ Correct Approach

```go
type Page struct {
    number  int
    content string
    // No prev/next references!
}

type Book struct {
    pages          []*Page  // Book owns the ordered collection
    currentPageIdx int      // Book tracks reading position
}

func (b *Book) NextPage() error {
    if b.currentPageIdx >= len(b.pages)-1 {
        return errors.New("already at last page")
    }
    b.currentPageIdx++
    return nil
}
```

**Benefits**:
- Page is simple, focused on data
- Book manages structure and navigation
- Easy to reorder, add, or remove pages
- Clear separation of concerns

---

### 2. Book-Author Relationship: Bidirectional vs Unidirectional

#### The Question
Should the relationship be:
- Bidirectional: Book ↔ Author (both know about each other)
- Unidirectional: Book → Author (only Book knows Author)

#### ❌ Bidirectional (Avoid)

```go
type Book struct {
    authors []*Author
}

type Author struct {
    books []*Book  // ❌ Circular dependency
}
```

**Problems**:

1. **Circular Dependency**
```go
// Book imports Author package
// Author imports Book package
// Compilation error or tight coupling
```

2. **Memory Issues**
```go
// Serialization nightmare
book1 -> author1 -> book1 -> author1 -> ... (infinite loop)
```

3. **Maintenance Burden**
```go
// Adding a book to an author requires updating both sides
book.authors = append(book.authors, author)
author.books = append(author.books, book)
// Easy to forget one side and create inconsistency
```

#### ✅ Unidirectional (Recommended)

```go
type Author struct {
    id   string
    name string
    // No reference to books
}

type Book struct {
    authors []*Author  // Book knows its authors
}
```

**Benefits**:
- No circular dependencies
- Author is independent, reusable
- Easy to serialize
- If you need "books by author", query the book collection:

```go
func GetBooksByAuthor(books []*Book, authorID string) []*Book {
    var result []*Book
    for _, book := range books {
        for _, author := range book.authors {
            if author.id == authorID {
                result = append(result, book)
                break
            }
        }
    }
    return result
}
```

**Real-World Analogy**:
- A book lists its authors (on the cover)
- An author doesn't carry a list of all their books in their pocket
- You find an author's books by searching the library

---

### 3. Why Author Needs an ID

#### Scenario 1: Duplicate Names
```go
author1 := Author{name: "Michael Smith"}  // Science fiction writer
author2 := Author{name: "Michael Smith"}  // Romance novelist

// Without ID: How do you distinguish them?
// With ID:
author1 := Author{id: "auth_001", name: "Michael Smith"}
author2 := Author{id: "auth_002", name: "Michael Smith"}
```

#### Scenario 2: Removing an Author
```go
// Without ID
book.removeAuthor("Michael Smith")  // Which one? Ambiguous!

// With ID
book.removeAuthor("auth_001")  // Clear and unambiguous
```

#### Scenario 3: Database Persistence
```sql
CREATE TABLE authors (
    id VARCHAR PRIMARY KEY,     -- Need unique identifier
    name VARCHAR                -- Names can duplicate
);

CREATE TABLE book_authors (
    book_isbn VARCHAR,
    author_id VARCHAR,          -- Foreign key to authors.id
    PRIMARY KEY (book_isbn, author_id)
);
```

**When You Can Skip ID**:
- Guaranteed unique names
- No persistence layer
- Simple in-memory prototype
- Interviewer explicitly says "keep it minimal"

---

### 4. Immutable Author (No Setters)

#### ❌ Mutable Author (Problematic)

```go
type Author struct {
    id   string
    name string
}

func (a *Author) SetName(name string) {
    a.name = name
}
```

**Problem: Shared Reference Side Effects**
```go
author := NewAuthor("auth_001", "J.K. Rowling")
book1.addAuthor(author)
book2.addAuthor(author)  // Same author object

// Later, someone does:
author.SetName("Random Person")

// Both book1 and book2 now show wrong author!
// Unexpected side effects across the system
```

**Problem: Thread Safety**
```go
// Goroutine 1
author.SetName("Name A")

// Goroutine 2 (simultaneously)
author.SetName("Name B")

// Race condition! Need locks, complexity increases
```

#### ✅ Immutable Author (Recommended)

```go
type Author struct {
    id   string
    name string
}

func NewAuthor(id, name string) *Author {
    return &Author{id: id, name: name}
}

// Only getters, no setters
func (a *Author) GetId() string   { return a.id }
func (a *Author) GetName() string { return a.name }
```

**Benefits**:
- Thread-safe by default (no locks needed)
- No unexpected side effects
- Predictable behavior
- Easier to reason about

**If Author Info Must Change**:
```go
// Create new author object
oldAuthor := NewAuthor("auth_001", "Jane Smith")
newAuthor := NewAuthor("auth_001", "Jane Johnson")

book.removeAuthor("auth_001")
book.addAuthor(newAuthor)
```

---

### 5. Navigation Methods: Return Types

#### ❌ Returning Page Objects

```go
func (b *Book) NextPage() Page {
    // What if at last page? Return nil? Zero value?
    // Forces error handling at wrong level
}
```

#### ✅ Separation of Concerns

```go
// Commands: Change state, return errors
func (b *Book) NextPage() error {
    if !b.isOpen {
        return errors.New("book is closed")
    }
    if b.currentPageIdx >= len(b.pages)-1 {
        return errors.New("already at last page")
    }
    b.currentPageIdx++
    return nil
}

// Queries: Read state, return data
func (b *Book) GetCurrentPage() (*Page, error) {
    if !b.isOpen {
        return nil, errors.New("book is closed")
    }
    return b.pages[b.currentPageIdx], nil
}

func (b *Book) ReadCurrentPage() (string, error) {
    if !b.isOpen {
        return "", errors.New("book is closed")
    }
    return b.pages[b.currentPageIdx].content, nil
}
```

**Benefits**:
- Clear error handling
- Follows Command-Query Separation (CQS)
- Navigation and reading are separate operations
- Easier to test and maintain

---

### 6. State Management

#### Why Track `isOpen` and `currentPageIdx`?

```go
type Book struct {
    isOpen         bool  // Is the book currently open?
    currentPageIdx int   // Which page are we on?
}
```

**State Validation**:
```go
func (b *Book) NextPage() error {
    if !b.isOpen {
        return errors.New("book is closed")  // Can't flip if closed!
    }
    // ... rest of logic
}

func (b *Book) Open() error {
    if b.isOpen {
        return errors.New("book already open")
    }
    b.isOpen = true
    b.currentPageIdx = 0  // Start at first page
    return nil
}
```

**State Machine**:
```
[Closed] --open()--> [Open at page 0]
[Open]   --nextPage()--> [Open at page N]
[Open]   --close()--> [Closed]
```

**Real-World Analogy**:
- You can't read a closed book
- Opening a book puts you at the first page
- Closing a book resets your position

---

### 7. Why Not Store `numberOfPages` Field?

#### ❌ Redundant Data

```go
type Book struct {
    pages         []*Page
    numberOfPages int      // ❌ Derived data, can get out of sync
}

func (b *Book) AddPage(page *Page) {
    b.pages = append(b.pages, page)
    b.numberOfPages++  // Easy to forget!
}
```

**Problem: Data Inconsistency**
```go
// Someone forgets to update numberOfPages
b.pages = append(b.pages, newPage)
// Now len(b.pages) != b.numberOfPages
```

#### ✅ Derive from Source of Truth

```go
type Book struct {
    pages []*Page  // Single source of truth
}

func (b *Book) GetNumberOfPages() int {
    return len(b.pages)  // Always correct
}
```

**Exception**: If pages are lazy-loaded from database:
```go
type Book struct {
    pages         []*Page  // Loaded on demand
    numberOfPages int      // Cached from DB for performance
}
```

---

## Design Principles Applied

### 1. Single Responsibility Principle (SRP)
- **Author**: Represents author data
- **Page**: Represents page data
- **Book**: Orchestrates reading experience

### 2. Encapsulation
- Private fields, public methods
- State validation in methods
- Hide implementation details

### 3. Low Coupling
- Page doesn't know about Book
- Author doesn't know about Book
- Unidirectional relationships

### 4. High Cohesion
- Each class has focused, related methods
- Methods grouped by responsibility

### 5. Immutability Where Possible
- Author is immutable (thread-safe, predictable)
- Page content might be mutable (depends on requirements)

### 6. Command-Query Separation
- Commands: `nextPage()`, `open()`, `close()` (change state, return errors)
- Queries: `getCurrentPage()`, `readCurrentPage()` (read state, return data)

---

## Interview Tips

### Always Ask Clarifying Questions

1. "Should we support multiple readers reading the same book at different positions?"
   - If yes: Consider `ReadingSession` class
   
2. "Can pages be edited after creation?"
   - If yes: Add `setContent()` to Page
   
3. "Do we need to handle duplicate author names?"
   - If yes: Add `id` to Author
   
4. "Should we persist reading progress?"
   - If yes: Consider persistence layer

5. "Can a book be read by multiple people simultaneously?"
   - If yes: Separate `Book` (data) from `ReadingSession` (state)

### Show Your Thinking Process

Don't just present a solution. Explain:
- "I'm putting navigation in Book because..."
- "I'm making Author immutable to avoid..."
- "I considered X but chose Y because..."

### Start Simple, Then Extend

1. Start with core classes (Book, Page, Author)
2. Add basic operations (open, read, close)
3. Then add advanced features (bookmarks, annotations)

Don't over-engineer upfront!

### Handle Edge Cases

Always think about:
- What if book is closed?
- What if at first/last page?
- What if page doesn't exist?
- What if no authors?

---

## Common Mistakes to Avoid

### ❌ Putting Navigation in Page
```go
type Page struct {
    prev *Page  // Don't do this!
    next *Page  // Don't do this!
}
```

### ❌ Bidirectional Relationships
```go
type Author struct {
    books []*Book  // Circular dependency!
}
```

### ❌ Returning Data from Commands
```go
func NextPage() Page {  // Should return error, not Page
}
```

### ❌ No State Validation
```go
func (b *Book) NextPage() {
    b.currentPageIdx++  // What if book is closed? At last page?
}
```

### ❌ Mutable Shared Objects
```go
author.SetName("New Name")  // Affects all books sharing this author!
```

---

## Extensions to Consider

### 1. Multiple Readers
```go
type ReadingSession struct {
    book           *Book
    currentPageIdx int
    bookmarks      []int
}
```

### 2. Bookmarks
```go
func (b *Book) AddBookmark(pageNum int) error
func (b *Book) GetBookmarks() []int
func (b *Book) GoToBookmark(index int) error
```

### 3. Annotations
```go
type Annotation struct {
    pageNum int
    text    string
    created time.Time
}

func (b *Book) AddAnnotation(pageNum int, text string) error
```

### 4. Reading History
```go
type ReadingHistory struct {
    pagesRead     []int
    timeSpent     map[int]time.Duration
    lastReadPage  int
    lastReadTime  time.Time
}
```

---

## Running the Code

### In LLD Interviews: Code Must Execute!

It's not enough to just design classes. You must:
1. ✅ Write compilable code
2. ✅ Create a working main function
3. ✅ Demonstrate all features
4. ✅ Show error handling
5. ✅ Test edge cases

### How to Run

```bash
cd lld/book
go run .
```

### What the Demo Shows

The `main.go` demonstrates:
- ✅ Creating books, authors, and pages
- ✅ Opening and closing books
- ✅ Reading pages sequentially
- ✅ Navigation (next, previous, jump to page)
- ✅ Error handling (closed book, boundary conditions)
- ✅ Multiple authors
- ✅ Author immutability

### Expected Output

```
=== Book Reading System Demo ===

✓ Added author: J.K. Rowling
✓ Added 5 pages

--- Book Information ---
Title: Harry Potter and the Sorcerer's Stone
ISBN: 978-0439708180
Publication Year: 1997
Authors: J.K. Rowling
Total Pages: 5

--- Testing Error Handling ---
✓ Cannot read closed book: book is closed

--- Opening the Book ---
✓ Book opened successfully
✓ Current page: 1

--- Reading All Pages ---
Page 1: Mr. and Mrs. Dursley, of number four...
[... all pages ...]
✓ Reached the last page

--- Testing Backward Navigation ---
✓ Moved to previous page

--- Testing Jump to Page ---
✓ Jumped to page 3

--- Testing Boundary Conditions ---
✓ Cannot go before first page: already at first page
✓ Cannot go beyond last page: already at last page

--- Closing the Book ---
✓ Book closed successfully

=== Demo Complete ===
```

### Interview Tips for Code Execution

1. **Write clean, runnable code from the start**
   - Don't write pseudocode
   - Use proper syntax
   - Handle errors properly

2. **Test as you go**
   - Run the code frequently
   - Fix compilation errors immediately
   - Don't wait until the end

3. **Demonstrate edge cases**
   - What happens at boundaries?
   - What if invalid input?
   - What if wrong state?

4. **Keep main() organized**
   - Group related tests
   - Add clear comments/sections
   - Show progression of features

5. **Be ready to extend**
   - Interviewer might ask: "Now add bookmarks"
   - Your code should be easy to extend
   - Show you can add features quickly

---

## Summary

This design demonstrates:
- ✅ Clear separation of concerns
- ✅ Proper encapsulation
- ✅ Unidirectional relationships
- ✅ Immutability where appropriate
- ✅ State management
- ✅ Error handling
- ✅ Extensibility
- ✅ **Working, executable code**

The key insight: **Book orchestrates, Page and Author are simple data containers.**

**Remember**: In LLD interviews, design is 40%, working code is 60%!
