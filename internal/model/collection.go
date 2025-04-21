package model

// CollectionType represents a type of collection owned by a user.
// Users can have multiple collections of different types (calendar, contacts, etc.)
//
// Django model reference:
//
//	class CollectionType(models.Model):
//	    owner = models.ForeignKey(AUTH_USER_MODEL, on_delete=CASCADE)
//	    uid = models.BinaryField(unique=True, max_length=1024)
type CollectionType struct {
	ID      uint   `gorm:"primaryKey"`
	OwnerID uint   `gorm:"not null;index"`
	UID     []byte `gorm:"type:bytea;uniqueIndex;not null"` // Binary UID

	// Relations
	Owner *User `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (CollectionType) TableName() string {
	return "django_collectiontype"
}

// Collection represents a collection of items (calendar, contacts, etc.)
//
// Django model reference:
//
//	class Collection(models.Model):
//	    main_item = models.OneToOneField("CollectionItem", null=True, on_delete=SET_NULL)
//	    uid = models.CharField(unique=True, max_length=43)
//	    owner = models.ForeignKey(AUTH_USER_MODEL, on_delete=CASCADE)
type Collection struct {
	ID         uint   `gorm:"primaryKey"`
	UID        string `gorm:"uniqueIndex;size:43;not null"` // Matches main_item.uid
	OwnerID    uint   `gorm:"not null;index"`
	MainItemID *uint  `gorm:"unique"` // Nullable, points to the "main" item

	// Relations
	Owner    *User           `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
	MainItem *CollectionItem `gorm:"foreignKey:MainItemID;constraint:OnDelete:SET NULL"`
	Items    []CollectionItem `gorm:"foreignKey:CollectionID"`
	Members  []CollectionMember `gorm:"foreignKey:CollectionID"`
}

// TableName specifies the table name for GORM
func (Collection) TableName() string {
	return "django_collection"
}

