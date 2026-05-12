# Golden Rules for Cracking LLD Problems

## The 7-Step Systematic Approach

### Step 1: Extract Entities (Nouns) and Actions (Verbs)
Read requirements carefully and highlight:
- **Nouns** → Potential classes
- **Verbs** → Potential methods

Example from Library System:
- Nouns: Book, Member, Library, Rack, Author, Copy, Reservation
- Verbs: search, checkout, reserve, return, calculate fine, notify

### Step 2: Identify Core vs Supporting Entities
Not all nouns become classes. Categorize:
- **Core Entities**: Book, Member, BookItem, Library
- **Supporting Entities**: Author, Rack, Reservation, Lending
- **Value Objects**: Barcode, Fine, Status

### Step 3: Define Responsibilities (Single Responsibility Principle)
For each entity, ask: "What should this entity KNOW and DO?"

| Entity | Knows | Does |
|--------|-------|------|
| Book | Metadata (title, ISBN) | Nothing (pure data) |
| BookItem | Status, location | Change status |
| Library | All books, members | Orchestrate operations |
| Member | Personal info | Track history |
| Lending | Dates, fine | Calculate fine |

### Step 4: Establish Relationships
Ask: "How do entities relate?"
- **Has-a**: Library HAS books, Member HAS checkouts
- **Is-a**: Librarian IS-A Member (inheritance)
- **Uses-a**: Library USES NotificationService

### Step 5: Apply Design Patterns
Recognize common patterns:
- Multiple search criteria → **Strategy Pattern**
- Different notifications → **Observer Pattern**
- Creating complex objects → **Factory Pattern**
- Single instance needed → **Singleton Pattern**

### Step 6: Handle Edge Cases
Think about:
- What if member exceeds limit?
- What if book is already checked out?
- What if reservation queue is empty?
- What if fine calculation fails?

### Step 7: Keep It Simple (YAGNI)
Don't over-engineer:
- Start with minimal implementation
- Add complexity only when needed
- Avoid premature optimization

---

## Common Confusions & Solutions

### Confusion 1: Where should search() method go?
**Wrong**: In Book class
```go
// ❌ Wrong - Book shouldn't search itself
func (b *Book) Search(title string) *Book
```

**Right**: In Library/Catalog class
```go
// ✅ Right - Library knows all books
func (lib *Library) SearchBooks(strategy SearchStrategy) []*Book
```

**Rule**: Entity shouldn't search for itself. The container/manager searches.

### Confusion 2: Should Book know its Rack?
**Wrong**: Book has Rack field
```go
// ❌ Wrong - Book is metadata, not physical
type Book struct {
    Title string
    Rack  *Rack
}
```

**Right**: BookItem (physical copy) has Rack
```go
// ✅ Right - Physical copy has location
type BookItem struct {
    Book  *Book
    Rack  *Rack
}
```

**Rule**: Separate abstract concept from physical instance.

### Confusion 3: Who calculates fine?
**Options**:
1. Library calculates → Too much responsibility
2. Member calculates → Member doesn't know due dates
3. Lending calculates → ✅ Lending knows checkout & return dates

**Rule**: Assign responsibility to the entity with most relevant data.

### Confusion 4: Should Rack have Books or Books have Rack?
**Answer**: BookItem has Rack reference (not vice versa)

**Why?**
- Rack is just a location identifier
- BookItem is "located at" a Rack
- Rack doesn't "own" books

---

## Design Principles Checklist

### SOLID Principles
- [ ] **S**ingle Responsibility: Each class has one reason to change
- [ ] **O**pen/Closed: Open for extension, closed for modification
- [ ] **L**iskov Substitution: Subtypes must be substitutable
- [ ] **I**nterface Segregation: Many specific interfaces > one general
- [ ] **D**ependency Inversion: Depend on abstractions, not concretions

### Additional Principles
- [ ] **DRY**: Don't Repeat Yourself
- [ ] **KISS**: Keep It Simple, Stupid
- [ ] **YAGNI**: You Aren't Gonna Need It
- [ ] **Composition over Inheritance**
- [ ] **Program to Interface, not Implementation**

---

## Mental Models for Common Scenarios

### Scenario: Multiple copies of same item
**Pattern**: Separate metadata from instances
- Book (metadata) → BookItem (physical copy)
- Product (catalog) → Inventory (stock)
- Movie (info) → Screening (showtime)

### Scenario: Different search criteria
**Pattern**: Strategy Pattern
```go
interface SearchStrategy {
    Search(items) []Item
}
```

### Scenario: Status tracking
**Pattern**: State Pattern or Enum
```go
type Status string
const (
    Available Status = "AVAILABLE"
    CheckedOut Status = "CHECKED_OUT"
)
```

### Scenario: Notifications
**Pattern**: Observer Pattern or Service
```go
type NotificationService struct {}
func (ns *NotificationService) Notify(user, message)
```

---

## Quick Decision Tree

```
Is it a thing? → Class
Is it a property? → Field
Is it an action? → Method
Is it a rule? → Constant/Config
Is it a variation? → Strategy/Subclass
Is it a relationship? → Association/Composition
```

---

## Common Mistakes to Avoid

1. **God Class**: One class doing everything
   - Solution: Split responsibilities

2. **Anemic Domain Model**: Classes with only getters/setters
   - Solution: Add behavior to entities

3. **Premature Optimization**: Over-engineering from start
   - Solution: Start simple, refactor later

4. **Ignoring Edge Cases**: Not handling errors
   - Solution: Think about failure scenarios

5. **Tight Coupling**: Classes depend on concrete implementations
   - Solution: Use interfaces

---

## Practice Approach

1. **Read requirements 3 times**
   - First: Get overview
   - Second: Extract entities
   - Third: Identify relationships

2. **Draw before coding**
   - Class diagram
   - Sequence diagram for key flows

3. **Start with core flow**
   - Implement happy path first
   - Add edge cases later

4. **Iterate**
   - Code → Review → Refactor
   - Don't aim for perfection in first attempt

---

## Key Takeaway

> "Good design is not about following rules blindly. It's about understanding the problem deeply and making conscious trade-offs."

Start simple. Add complexity only when justified.
