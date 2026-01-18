# KindredCard v0.2.0 - Relationships Update

## 🎉 New Feature: Contact Relationships

KindredCard now supports linking contacts together through relationships! You can model family trees, professional networks, and social connections.

### What's New

#### Database Changes
- **New Tables**:
  - `relationship_types`: Stores available relationship types (33 pre-configured)
  - `relationships`: Links contacts together with relationship types

#### API Endpoints
- `GET /api/v1/relationship-types` - Get all relationship types
- `POST /api/v1/relationship-types` - Create custom relationship type
- `POST /api/v1/contacts/{id}/relationships` - Add relationship between contacts
- `DELETE /api/v1/relationships/{rel_id}` - Remove a relationship

#### CardDAV Support
- Relationships are exported using Apple's X-ABRELATEDNAMES format
- Compatible with Apple Contacts, iCloud, and most CardDAV clients
- Format: `item1.X-ABRELATEDNAMES:Name` + `item1.X-ABLabel:_$!<Type>!$_`

#### Web Interface
- Contact cards now display relationships
- Relationships shown as colored badges on contact cards
- Visual indication of how contacts are connected

### Relationship Types Included

**Family (25 types):**
Mother, Father, Son, Daughter, Brother, Sister, Husband, Wife, Spouse, Partner, Ex-Husband, Ex-Wife, Step-Mother, Step-Father, Step-Son, Step-Daughter, Grandfather, Grandmother, Grandson, Granddaughter, Uncle, Aunt, Nephew, Niece, Cousin, Parent, Child, Sibling

**Professional (5 types):**
Manager, Assistant, Colleague, Mentor, Direct Report

**Social (1 type):**
Friend

**Plus:** Create unlimited custom relationship types!

### Migration Guide

If you have an existing KindredCard installation:

```bash
# Run the migration
psql -d KindredCard < migrations/001_add_relationships.sql

# Restart KindredCard
make run
```

### Examples

#### Create a Family Tree
```bash
# John's mother is Jane
curl -X POST http://localhost:8080/api/v1/contacts/123/relationships \
  -H "Content-Type: application/json" \
  -d '{"related_contact_id": 456, "relationship_type_id": 1}'

# John's father is Bob
curl -X POST http://localhost:8080/api/v1/contacts/123/relationships \
  -H "Content-Type: application/json" \
  -d '{"related_contact_id": 789, "relationship_type_id": 2}'
```

#### Run the Demo
```bash
make demo
# or
bash examples/relationships_demo.sh
```

### Technical Details

- **Cascade Delete**: When a contact is deleted, all relationships are automatically removed
- **Duplicate Prevention**: The same relationship between two contacts cannot be created twice
- **Bidirectional Support**: Some relationships (like Friend) should be created in both directions
- **Extensible**: Create custom relationship types for your specific needs

### Documentation

- 📖 **[RELATIONSHIPS.md](RELATIONSHIPS.md)** - Complete relationship feature documentation
- 📋 **[RELATIONSHIP_TYPES.md](RELATIONSHIP_TYPES.md)** - Quick reference of all relationship types
- 🧪 **[relationships_test.go](internal/db/relationships_test.go)** - Test suite for relationships

### Breaking Changes

None! This is a backwards-compatible addition.

### Database Schema Changes

```sql
-- New tables
relationship_types (id, name, reverse_name, is_system)
relationships (id, contact_id, related_contact_id, relationship_type_id, created_at)

-- New indexes
idx_relationships_contact_id
idx_relationships_related_contact_id  
idx_relationships_type
```

### Model Changes

The `Contact` model now includes a `Relationships` field:
```go
type Contact struct {
    // ... existing fields ...
    Relationships []Relationship `json:"relationships,omitempty"`
}
```

### Future Enhancements

Planned improvements:
- Automatic bidirectional relationship creation
- Visual family tree generator
- Relationship history tracking
- Relationship validation (prevent conflicts)
- Smart relationship suggestions
- Bulk relationship import/export

### Credits

Developed with ❤️ for personal CRM enthusiasts who want to maintain rich relationship data about their contacts.

### Support

- 🐛 **Issues**: [Report bugs or request features](https://github.com/steveredden/KindredCard/issues)
- 💬 **Discussions**: Share use cases and ideas
- 📚 **Documentation**: See README.md and RELATIONSHIPS.md

---

## Installation

### New Installation
```bash
# Clone and setup
git clone https://github.com/steveredden/KindredCard.git
cd KindredCard
make docker-up
make run
```

### Existing Installation
```bash
# Update code
git pull

# Run migration
psql -d KindredCard < migrations/001_add_relationships.sql

# Restart
make run
```

Enjoy building your connected contact network! 🌐
