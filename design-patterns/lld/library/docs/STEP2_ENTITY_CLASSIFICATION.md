# Step 2: Entity Classification - Deep Dive

## What Are Core Entities?

**Core Entities** are the main business objects that:
1. Represent the PRIMARY domain concepts
2. Have independent lifecycle (can exist on their own)
3. Are the "stars" of your system
4. Users directly interact with them
5. Have unique identity (ID/primary key)

### Examples from Library System:
- **Book** - The system is ABOUT books
- **Member** - The system is FOR members
- **Library** - The system IS a library
- **BookItem** - Physical copies that get checked out

### Test: "If I remove this, does the system still make sense?"
- Remove Book? ❌ No library without books
- Remove Member? ❌ No one to borrow books
- Remove Library? ❌ Nothing to manage

---

## What Are Supporting Entities?

**Supporting Entities** help core entities function but:
1. Don't make sense without core entities
2. Provide additional context/details
3. Enable specific features
4. Have dependent lifecycle

### Examples from Library System:
- **Author** - Supports Book (books have authors)
- **Rack** - Supports BookItem (location info)
- **Lending** - Supports checkout transaction
- **Reservation** - Supports booking feature
- **Notification** - Supports communication

### Test: "Can this exist without core entities?"
- Author without Book? 🤔 Technically yes, but not useful in library context
- Rack without BookItem? ❌ Just empty shelves
- Lending without Book/Member? ❌ No transaction possible

---

## What Are Value Objects?

**Value Objects** are:
1. Defined by their VALUE, not identity
2. Immutable (don't change after creation)
3. No unique ID needed
4. Can be replaced entirely
5. Often primitive types or simple structs

### Key Difference: Entity vs Value Object

| Aspect | Entity | Value Object |
|--------|--------|--------------|
| Identity | Has unique ID | No ID, defined by value |
| Equality | Same ID = same object | Same values = same object |
| Mutability | Can change over time | Immutable |
| Lifecycle | Independent | Part of entity |
| Example | Member (ID: M001) | Email ("alice@email.com") |

### Examples from Library System:

**Value Objects:**
- **Barcode** - Just a string value
- **ISBN** - Just a string value
- **Status** - Enum (AVAILABLE, CHECKED_OUT)
- **Fine** - Just a number (amount)
- **Email** - String with validation
- **Phone** - String with format

**Why they're Value Objects:**
```go
// Two members with same email are DIFFERENT people
member1 := Member{ID: "M001", Email: "alice@email.com"}
member2 := Member{ID: "M002", Email: "alice@email.com"}
// member1 != member2 (different IDs)

// Two emails with same value are THE SAME
email1 := "alice@email.com"
email2 := "alice@email.com"
// email1 == email2 (same value)
```

---

## The Secret to Separating Them

### Secret #1: The "Noun Importance Test"

Ask: **"Is this noun a PRIMARY actor or a SUPPORTING detail?"**

Example: "Library members can checkout books"
- **members** → Core (primary actor)
- **books** → Core (primary actor)
- **checkout** → Supporting (action/transaction)

### Secret #2: The "Can It Live Alone Test"

Ask: **"Does this make sense without other entities?"**

```
Book without Member? ✅ Yes (library can have books with no members)
Member without Book? ✅ Yes (member can exist before borrowing)
Lending without Book/Member? ❌ No (transaction needs both)
```

### Secret #3: The "Identity Test"

Ask: **"Do I need to track THIS specific instance over time?"**

```
Member: "Is THIS the same person who borrowed last week?"
→ YES, need ID → Entity

Status: "Is THIS 'AVAILABLE' the same as THAT 'AVAILABLE'?"
→ NO, just a value → Value Object
```

### Secret #4: The "Lifecycle Test"

Ask: **"When does this get created and destroyed?"**

```
Book: Created when added to library, exists independently
→ Core Entity

Lending: Created at checkout, destroyed at return, depends on Book+Member
→ Supporting Entity

Fine: Calculated at return, just a number
→ Value Object
```

### Secret #5: The "Database Table Test"

Ask: **"Would this get its own table with primary key?"**

```
Book → books table with ISBN as PK → Core Entity
Member → members table with member_id as PK → Core Entity
Lending → lendings table with lending_id as PK → Supporting Entity
Status → Just a column (status VARCHAR) → Value Object
```

---

## Practical Classification Framework

### Step-by-Step Process:

#### 1. List ALL nouns from requirements
```
Library, Book, Member, Author, Rack, Copy, Barcode, 
Checkout, Return, Reservation, Fine, Notification, 
Status, ISBN, Title, Date, Limit
```

#### 2. Apply "Primary Actor Test"
**Question: "Is the system ABOUT this?"**

✅ Core: Book, Member, Library, BookItem (copy)
🤔 Maybe: Author, Reservation, Lending
❌ Not Core: Barcode, Status, Fine, Title, Date, Limit

#### 3. Apply "Identity Test"
**Question: "Do I track specific instances?"**

✅ Has Identity: Book (ISBN), Member (ID), BookItem (barcode), Lending (ID)
❌ No Identity: Status, Fine, Barcode (just values)

#### 4. Apply "Dependency Test"
**Question: "Can this exist independently?"**

✅ Independent: Book, Member, Library
❌ Dependent: Lending (needs Book+Member), Reservation (needs Book+Member)

#### 5. Final Classification:

**Core Entities:**
- Library (the system itself)
- Book (main domain object)
- Member (main user)
- BookItem (physical instance)

**Supporting Entities:**
- Author (enriches Book)
- Lending (tracks transaction)
- Reservation (enables booking)
- Rack (provides location)

**Value Objects:**
- Barcode (string)
- ISBN (string)
- Status (enum)
- Fine (float)
- Email (string)
- Phone (string)

---

## Real-World Examples

### Example 1: E-Commerce System

**Core Entities:**
- Customer (who buys)
- Product (what's sold)
- Order (transaction)

**Supporting Entities:**
- OrderItem (line item in order)
- Review (feedback on product)
- ShippingAddress (delivery info)
- Payment (transaction detail)

**Value Objects:**
- Price (amount)
- ProductName (string)
- OrderStatus (enum)
- Email (string)

### Example 2: Hotel Booking System

**Core Entities:**
- Hotel (the property)
- Room (what's booked)
- Guest (who books)
- Booking (reservation)

**Supporting Entities:**
- RoomType (category)
- Amenity (features)
- Payment (transaction)
- Review (feedback)

**Value Objects:**
- RoomNumber (string)
- Price (amount)
- BookingStatus (enum)
- CheckInDate (date)

### Example 3: Hospital Management

**Core Entities:**
- Patient (who gets treated)
- Doctor (who treats)
- Appointment (scheduled visit)

**Supporting Entities:**
- Prescription (medication order)
- MedicalRecord (history)
- Department (organization)
- Bill (payment)

**Value Objects:**
- PatientID (string)
- Diagnosis (string)
- AppointmentStatus (enum)
- Amount (float)

---

## Common Mistakes

### Mistake 1: Making Everything an Entity
```go
// ❌ Wrong - Status doesn't need to be an entity
type Status struct {
    ID   string
    Name string
}

// ✅ Right - Status is just a value
type Status string
const (
    StatusAvailable Status = "AVAILABLE"
)
```

### Mistake 2: Making Core Things Value Objects
```go
// ❌ Wrong - Member needs identity
type Member struct {
    Name  string
    Email string
}
// Two members with same name are different people!

// ✅ Right - Member is an entity
type Member struct {
    ID    string  // Unique identifier
    Name  string
    Email string
}
```

### Mistake 3: Confusing Supporting with Value Objects
```go
// Author is a Supporting Entity (has identity, can exist)
type Author struct {
    ID      string  // Has unique ID
    Name    string
    Country string
}

// AuthorName is a Value Object (just a string)
type AuthorName string
```

---

## Decision Tree

```
Start with a noun
    |
    v
Does it have unique identity? 
    |
    ├─ NO → Value Object (Status, Price, Email)
    |
    └─ YES → Entity
            |
            v
        Is the system ABOUT this?
            |
            ├─ YES → Core Entity (Book, Member)
            |
            └─ NO → Can it exist without other entities?
                    |
                    ├─ YES → Core Entity (might be secondary)
                    |
                    └─ NO → Supporting Entity (Lending, Reservation)
```

---

## Quick Reference Table

| Characteristic | Core Entity | Supporting Entity | Value Object |
|---------------|-------------|-------------------|--------------|
| Has unique ID | ✅ Yes | ✅ Yes | ❌ No |
| Independent lifecycle | ✅ Yes | ❌ No | ❌ No |
| Primary domain concept | ✅ Yes | ❌ No | ❌ No |
| Mutable | ✅ Yes | ✅ Yes | ❌ No |
| Compared by | ID | ID | Value |
| Database table | ✅ Own table | ✅ Own table | ❌ Column |
| Example | Book, Member | Lending, Author | Status, Price |

---

## Key Takeaways

1. **Core Entities** = Main characters of your system
2. **Supporting Entities** = Help core entities do their job
3. **Value Objects** = Simple values without identity
4. Use the 5 tests: Importance, Independence, Identity, Lifecycle, Database
5. When in doubt, start as Value Object, promote to Entity if needed
6. Don't over-engineer - not everything needs to be an entity

---

## Practice Exercise

Classify these for a "Restaurant Ordering System":

Nouns: Restaurant, Menu, MenuItem, Customer, Order, OrderItem, 
       Table, Waiter, Payment, Price, Status, Rating

Try classifying them yourself, then check the answer below!

<details>
<summary>Click to see answer</summary>

**Core Entities:**
- Restaurant (the business)
- Customer (who orders)
- Order (main transaction)
- MenuItem (what's ordered)

**Supporting Entities:**
- OrderItem (line item in order)
- Table (seating assignment)
- Waiter (service provider)
- Payment (transaction detail)
- Menu (collection of items)

**Value Objects:**
- Price (amount)
- Status (enum: PENDING, COMPLETED)
- Rating (number 1-5)

</details>
