package routes

import (
	"errors"
	"fmt"
	"goatsync/internal/models"
	"goatsync/internal/repository"
	"goatsync/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Types for Collection
type CollectionType struct {
	UID   []byte      `json:"uid" msgpack:"uid"`
	Owner models.User `json:"owner" msgpack:"owner"`
}

type Collection struct {
	UID      string      `json:"uid" msgpack:"uid"`
	Owner    models.User `json:"owner" msgpack:"owner"`
	MainItem *Item       `json:"main_item" msgpack:"main_item"`
	Stoken   string      `json:"stoken" msgpack:"stoken"`
	Created  time.Time   `json:"created" msgpack:"created"`
}

type CollectionOut struct {
	UID      string    `json:"uid" msgpack:"uid"`
	MainItem *Item     `json:"main_item" msgpack:"main_item"`
	Stoken   string    `json:"stoken" msgpack:"stoken"`
	Created  time.Time `json:"created" msgpack:"created"`
}

// Types for Items
type Item struct {
	UID        string       `json:"uid" msgpack:"uid"`
	Version    int          `json:"version" msgpack:"version"`
	Collection Collection   `json:"collection" msgpack:"collection"`
	Content    ItemRevision `json:"content" msgpack:"content"`
}

type ItemRevision struct {
	UID     string     `json:"uid" msgpack:"uid"`
	Meta    []byte     `json:"meta" msgpack:"meta"`
	Deleted bool       `json:"deleted" msgpack:"deleted"`
	Chunks  []ChunkRef `json:"chunks" msgpack:"chunks"`
}

type ChunkRef struct {
	UID     string `json:"uid" msgpack:"uid"`
	Content []byte `json:"content,omitempty" msgpack:"content,omitempty"`
}

// Response types
type CollectionListResponse struct {
	Data               []CollectionOut `json:"data" msgpack:"data"`
	Stoken             *string         `json:"stoken,omitempty" msgpack:"stoken,omitempty"`
	Done               bool            `json:"done" msgpack:"done"`
	RemovedMemberships []string        `json:"removedMemberships,omitempty" msgpack:"removedMemberships,omitempty"`
}

type CollectionItemListResponse struct {
	Data   []ItemOut `json:"data" msgpack:"data"`
	Stoken *string   `json:"stoken,omitempty" msgpack:"stoken,omitempty"`
	Done   bool      `json:"done" msgpack:"done"`
}

type ListMultiRequest struct {
	CollectionTypes []string `json:"collectionTypes" msgpack:"collectionTypes"`
	Stoken          string   `json:"stoken,omitempty" msgpack:"stoken,omitempty"`
	Limit           int      `json:"limit,omitempty" msgpack:"limit,omitempty"`
}

func ListMultiCollections(context *gin.Context) {
	var request ListMultiRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		utils.BadReqError(context, err)
		return
	}

	// Set default limit if not provided
	if request.Limit <= 0 {
		request.Limit = 50
	}

	// Get collections from database
	collections := make([]Collection, 0)
	// TODO: Replace this with actual database query
	// Example query:
	// err := db.Where("owner = ? AND type IN ?", username, request.CollectionTypes).
	//    Limit(request.Limit + 1).
	//    Find(&collections).Error

	// Check if there are more results
	done := true
	if len(collections) > request.Limit {
		done = false
		collections = collections[:request.Limit]
	}

	// Convert to CollectionOut
	collectionData := make([]CollectionOut, len(collections))
	for i, collection := range collections {
		collectionData[i] = CollectionOut{
			UID:      collection.UID,
			MainItem: collection.MainItem,
			Stoken:   collection.Stoken,
			Created:  collection.Created,
		}
	}

	// Generate new stoken
	newStoken := uuid.New().String()

	response := CollectionListResponse{
		Data:   collectionData,
		Stoken: &newStoken,
		Done:   done,
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

type ListCollectionsRequest struct {
	Stoken string `form:"stoken"`
	Limit  int    `form:"limit,default=50"`
}

func ListCollections(context *gin.Context) {
	// Bind query parameters
	var request ListCollectionsRequest
	if err := context.ShouldBindQuery(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default limit if not provided
	if request.Limit <= 0 {
		request.Limit = 50
	}

	// Get collections from database
	collections := make([]Collection, 0)
	// TODO: Replace this with actual database query
	// Example query:
	// query := db.Where("owner = ?", username)
	// if request.Stoken != "" {
	//     query = query.Where("stoken > ?", request.Stoken)
	// }
	// err := query.Limit(request.Limit + 1).Find(&collections).Error

	// Check if there are more results
	done := true
	if len(collections) > request.Limit {
		done = false
		collections = collections[:request.Limit]
	}

	// Convert to CollectionOut
	collectionData := make([]CollectionOut, len(collections))
	for i, collection := range collections {
		collectionData[i] = CollectionOut{
			UID:      collection.UID,
			MainItem: collection.MainItem,
			Stoken:   collection.Stoken,
			Created:  collection.Created,
		}
	}

	// Generate new stoken
	newStoken := uuid.New().String()

	response := CollectionListResponse{
		Data:               collectionData,
		Stoken:             &newStoken,
		Done:               done,
		RemovedMemberships: []string{}, // TODO: Implement removed memberships logic
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

// Add these types after your existing type definitions
type CreateCollectionRequest struct {
	Collection CollectionIn `json:"collection" msgpack:"collection"`
	MainItem   ItemIn       `json:"item" msgpack:"item"`
}

type CollectionIn struct {
	UID           string `json:"uid" msgpack:"uid"`
	EncryptionKey []byte `json:"encryptionKey" msgpack:"encryptionKey"`
	AccessLevel   string `json:"accessLevel" msgpack:"accessLevel"`
}

type ItemIn struct {
	UID     string `json:"uid" msgpack:"uid"`
	Version int    `json:"version" msgpack:"version"`
	Content []byte `json:"content" msgpack:"content"`
}

// Add this function after listCollections
func CreateCollection(context *gin.Context) {
	tokenStr := context.GetHeader("Authorization")
	if tokenStr == "" {
		utils.BadReqError(context, errors.New("missing token from Header"))
		return
	}

	tokens := repository.GetTokens()
	token, exists := tokens[tokenStr]
	if !exists {
		utils.UnauthorizedError(context, errors.New("invalid token"))
		return
	}

	// Bind request body
	var request CreateCollectionRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if request.Collection.UID == "" || request.MainItem.UID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	// Create new collection
	collection := Collection{
		UID:     request.Collection.UID,
		Owner:   models.User{Username: token.User.Username},
		Stoken:  uuid.New().String(),
		Created: time.Now(),
	}

	// Create main item
	item := Item{
		UID:        request.MainItem.UID,
		Version:    request.MainItem.Version,
		Collection: collection,
		Content: ItemRevision{
			UID:     uuid.New().String(),
			Meta:    request.MainItem.Content,
			Deleted: false,
			Chunks:  []ChunkRef{},
		},
	}
	collection.MainItem = &item

	// TODO: Save to database
	// Example:
	// err := db.Transaction(func(tx *gorm.DB) error {
	//     if err := tx.Create(&collection).Error; err != nil {
	//         return err
	//     }
	//     return tx.Create(&item).Error
	// })

	// Prepare response
	response := CollectionOut{
		UID:      collection.UID,
		MainItem: collection.MainItem,
		Stoken:   collection.Stoken,
		Created:  collection.Created,
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

// Add this type after your existing type definitions
type CollectionGetResponse struct {
	Collection  CollectionOut `json:"collection" msgpack:"collection"`
	AccessLevel string        `json:"accessLevel" msgpack:"accessLevel"`
}

func GetCollection(context *gin.Context) {
	// Get collection UID from URL parameter
	collectionUID := context.Param("collection_uid")
	if collectionUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing collection UID"})
		return
	}

	// Get collection from database
	var collection Collection
	// TODO: Replace this with actual database query
	// Example query:
	// err := db.Where("uid = ? AND owner = ?", collectionUID, username).
	//     Preload("MainItem").
	//     First(&collection).Error
	// if err != nil {
	//     if errors.Is(err, gorm.ErrRecordNotFound) {
	//         c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
	//         return
	//     }
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
	//     return
	// }

	// Prepare response
	response := CollectionGetResponse{
		Collection: CollectionOut{
			UID:      collection.UID,
			MainItem: collection.MainItem,
			Stoken:   collection.Stoken,
			Created:  collection.Created,
		},
		AccessLevel: "readwrite", // TODO: Implement access level logic
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

// Add these types after your existing type definitions
type ItemOut struct {
	UID     string       `json:"uid" msgpack:"uid"`
	Version int          `json:"version" msgpack:"version"`
	Content ItemRevision `json:"content" msgpack:"content"`
}

type ItemGetResponse struct {
	Item ItemOut `json:"item" msgpack:"item"`
}

// Add this function after getCollection
func GetItem(context *gin.Context) {
	// Get collection and item UIDs from URL parameters
	collectionUID := context.Param("collection_uid")
	itemUID := context.Param("item_uid")
	if collectionUID == "" || itemUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing collection or item UID"})
		return
	}

	// Get item from database
	var item Item
	// TODO: Replace this with actual database query
	// Example query:
	// err := db.Joins("Collection").
	//     Where("items.uid = ? AND collections.uid = ? AND collections.owner = ?",
	//           itemUID, collectionUID, username).
	//     First(&item).Error
	// if err != nil {
	//     if errors.Is(err, gorm.ErrRecordNotFound) {
	//         c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
	//         return
	//     }
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
	//     return
	// }

	// Prepare response
	response := ItemGetResponse{
		Item: ItemOut{
			UID:     item.UID,
			Version: item.Version,
			Content: item.Content,
		},
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

// Add these types after your existing type definitions
type ListItemsRequest struct {
	Stoken   string `form:"stoken"`
	Limit    int    `form:"limit,default=50"`
	Prefetch string `form:"prefetch"`
}

// Add this function after getCollection
func ListItems(context *gin.Context) {
	// Get collection UID from URL parameter
	collectionUID := context.Param("collection_uid")
	if collectionUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing collection UID"})
		return
	}

	// Bind query parameters
	var request ListItemsRequest
	if err := context.ShouldBindQuery(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default limit if not provided
	if request.Limit <= 0 {
		request.Limit = 50
	}

	// Get items from database
	items := make([]Item, 0)
	// TODO: Replace this with actual database query
	// Example query:
	// query := db.Joins("Collection").
	//     Where("collections.uid = ? AND collections.owner = ?", collectionUID, username)
	// if request.Stoken != "" {
	//     query = query.Where("items.stoken > ?", request.Stoken)
	// }
	// err := query.Limit(request.Limit + 1).Find(&items).Error

	// Check if there are more results
	done := true
	if len(items) > request.Limit {
		done = false
		items = items[:request.Limit]
	}

	// Convert to ItemOut
	itemData := make([]ItemOut, len(items))
	for i, item := range items {
		itemData[i] = ItemOut{
			UID:     item.UID,
			Version: item.Version,
			Content: item.Content,
		}
	}

	// Generate new stoken
	newStoken := uuid.New().String()

	response := CollectionItemListResponse{
		Data:   itemData,
		Stoken: &newStoken,
		Done:   done,
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

// Add these types after your existing type definitions
type ItemRevisionOut struct {
	UID     string     `json:"uid" msgpack:"uid"`
	Meta    []byte     `json:"meta" msgpack:"meta"`
	Deleted bool       `json:"deleted" msgpack:"deleted"`
	Chunks  []ChunkRef `json:"chunks" msgpack:"chunks"`
}

type ItemRevisionListResponse struct {
	Data     []ItemRevisionOut `json:"data" msgpack:"data"`
	Iterator *string           `json:"iterator,omitempty" msgpack:"iterator,omitempty"`
	Done     bool              `json:"done" msgpack:"done"`
}

// Add this function after listItems
func GetItemRevisions(context *gin.Context) {
	// Get URL parameters
	// itemUID := context.Param("item_uid")
	limit := 50 // Default limit
	if limitParam := context.Query("limit"); limitParam != "" {
		if _, err := fmt.Sscanf(limitParam, "%d", &limit); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
	}
	// iterator := context.Query("iterator")

	// Get item revisions from database
	revisions := make([]ItemRevision, 0)
	// TODO: Replace this with actual database query
	// Example query:
	// query := db.Model(&ItemRevision{}).
	//     Joins("Item").
	//     Joins("Item.Collection").
	//     Where("items.uid = ? AND collections.owner = ?", itemUID, username).
	//     Order("item_revisions.id DESC")
	// if iterator != "" {
	//     var iteratorRev ItemRevision
	//     if err := query.Where("uid = ?", iterator).First(&iteratorRev).Error; err != nil {
	//         c.JSON(http.StatusBadRequest, gin.H{"error": "invalid iterator"})
	//         return
	//     }
	//     query = query.Where("item_revisions.id < ?", iteratorRev.ID)
	// }
	// err := query.Limit(limit + 1).Find(&revisions).Error

	// Check if there are more results
	done := true
	if len(revisions) > limit {
		done = false
		revisions = revisions[:limit]
	}

	// Convert to ItemRevisionOut
	revisionData := make([]ItemRevisionOut, len(revisions))
	for i, rev := range revisions {
		revisionData[i] = ItemRevisionOut{
			UID:     rev.UID,
			Meta:    rev.Meta,
			Deleted: rev.Deleted,
			Chunks:  rev.Chunks,
		}
	}

	// Get next iterator
	var nextIterator *string
	if len(revisionData) > 0 {
		iter := revisionData[len(revisionData)-1].UID
		nextIterator = &iter
	}

	response := ItemRevisionListResponse{
		Data:     revisionData,
		Iterator: nextIterator,
		Done:     done,
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

// Add these types after your existing type definitions
type SubscriptionTicketResponse struct {
	Ticket string `json:"ticket" msgpack:"ticket"`
}

// Add this function after getItemRevisions
func GetSubscriptionTicket(context *gin.Context) {
	// Get collection UID from URL parameter
	collectionUID := context.Param("collection_uid")
	if collectionUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing collection UID"})
		return
	}

	// Generate subscription ticket
	// This should be a secure, time-limited token that can be used for WebSocket authentication
	ticket := uuid.New().String()

	// TODO: Store the ticket with expiration time in Redis or similar
	// Example:
	// err := redisClient.Set(ctx, ticket, username, 24*time.Hour).Err()
	// if err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate ticket"})
	//     return
	// }

	response := SubscriptionTicketResponse{
		Ticket: ticket,
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

// Add these types after your existing type definitions
type FetchUpdatesRequest struct {
	Items []ItemUpdateRequest `json:"items" msgpack:"items"`
}

type ItemUpdateRequest struct {
	UID  string `json:"uid" msgpack:"uid"`
	Etag string `json:"etag" msgpack:"etag"`
}

type FetchUpdatesResponse struct {
	Items []ItemOut `json:"items" msgpack:"items"`
}

// Add this function after getSubscriptionTicket
func FetchUpdates(context *gin.Context) {
	// Get collection UID from URL parameter
	collectionUID := context.Param("collection_uid")
	if collectionUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing collection UID"})
		return
	}

	// Bind request body
	var request FetchUpdatesRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get items from database that have changed
	changedItems := make([]Item, 0)
	// TODO: Replace this with actual database query
	// Example query:
	// for _, itemReq := range request.Items {
	//     var item Item
	//     err := db.Joins("Collection").
	//         Where("items.uid = ? AND collections.uid = ? AND collections.owner = ?",
	//               itemReq.UID, collectionUID, username).
	//         First(&item).Error
	//     if err != nil {
	//         continue
	//     }
	//     // Compare etag and add to changedItems if different
	//     if calculateEtag(item) != itemReq.Etag {
	//         changedItems = append(changedItems, item)
	//     }
	// }

	// Convert to ItemOut
	itemData := make([]ItemOut, len(changedItems))
	for i, item := range changedItems {
		itemData[i] = ItemOut{
			UID:     item.UID,
			Version: item.Version,
			Content: item.Content,
		}
	}

	response := FetchUpdatesResponse{
		Items: itemData,
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

// Add these types after your existing type definitions
type ItemTransaction struct {
	Item    ItemIn `json:"item" msgpack:"item"`
	Etag    string `json:"etag" msgpack:"etag"`
	Deleted bool   `json:"deleted" msgpack:"deleted"`
}

type ItemTransactionRequest struct {
	Items []ItemTransaction `json:"items" msgpack:"items"`
}

// Add this function after fetchUpdates
func ItemTransactions(context *gin.Context) {
	// Get collection UID from URL parameter
	collectionUID := context.Param("collection_uid")
	if collectionUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing collection UID"})
		return
	}

	// Bind request body
	var request ItemTransactionRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement database transaction
	// Example:
	// err := db.Transaction(func(tx *gorm.DB) error {
	//     for _, itemTx := range request.Items {
	//         var item Item
	//         err := tx.Joins("Collection").
	//             Where("items.uid = ? AND collections.uid = ? AND collections.owner = ?",
	//                   itemTx.Item.UID, collectionUID, username).
	//             First(&item).Error
	//
	//         if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
	//             return err
	//         }
	//
	//         // Check etag if item exists
	//         if err == nil && calculateEtag(item) != itemTx.Etag {
	//             return fmt.Errorf("item %s has been modified", itemTx.Item.UID)
	//         }
	//
	//         // Create or update item
	//         newItem := Item{
	//             UID:     itemTx.Item.UID,
	//             Version: itemTx.Item.Version,
	//             Content: ItemRevision{
	//                 UID:     uuid.New().String(),
	//                 Meta:    itemTx.Item.Content,
	//                 Deleted: itemTx.Deleted,
	//                 Chunks:  []ChunkRef{},
	//             },
	//         }
	//
	//         if err := tx.Save(&newItem).Error; err != nil {
	//             return err
	//         }
	//     }
	//     return nil
	// })
	//
	// if err != nil {
	//     c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	//     return
	// }

	context.Status(http.StatusNoContent)
}

// Add these types after your existing type definitions
type ItemBatchRequest struct {
	Items []ItemBatchEntry `json:"items" msgpack:"items"`
}

type ItemBatchEntry struct {
	Item    ItemIn `json:"item" msgpack:"item"`
	Deleted bool   `json:"deleted" msgpack:"deleted"`
}

// Add this function after itemTransaction
func ItemBatch(context *gin.Context) {
	// Get collection UID from URL parameter
	collectionUID := context.Param("collection_uid")
	if collectionUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing collection UID"})
		return
	}

	// Bind request body
	var request ItemBatchRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement database batch operation
	// Example:
	// err := db.Transaction(func(tx *gorm.DB) error {
	//     for _, batchItem := range request.Items {
	//         newItem := Item{
	//             UID:     batchItem.Item.UID,
	//             Version: batchItem.Item.Version,
	//             Collection: Collection{UID: collectionUID},
	//             Content: ItemRevision{
	//                 UID:     uuid.New().String(),
	//                 Meta:    batchItem.Item.Content,
	//                 Deleted: batchItem.Deleted,
	//                 Chunks:  []ChunkRef{},
	//             },
	//         }
	//
	//         if err := tx.Save(&newItem).Error; err != nil {
	//             return err
	//         }
	//     }
	//     return nil
	// })
	//
	// if err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//     return
	// }

	context.Status(http.StatusNoContent)
}

// Add these types after your existing type definitions
type ChunkUploadResponse struct {
	UID string `json:"uid" msgpack:"uid"`
}

// Add these functions after itemBatch
func UpdateChunk(context *gin.Context) {
	// Get URL parameters
	itemUID := context.Param("item_uid")
	chunkUID := context.Param("chunk_uid")
	if itemUID == "" || chunkUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing item or chunk UID"})
		return
	}

	// Read chunk content
	_, err := context.GetRawData()
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "failed to read chunk content"})
		return
	}

	// TODO: Implement chunk storage
	// Example:
	// chunk := ChunkRef{
	//     UID:     chunkUID,
	//     Content: content,
	// }
	// err = db.Transaction(func(tx *gorm.DB) error {
	//     var item Item
	//     if err := tx.Joins("Collection").
	//         Where("items.uid = ? AND collections.owner = ?", itemUID, username).
	//         First(&item).Error; err != nil {
	//         return err
	//     }
	//     return tx.Create(&chunk).Error
	// })
	// if err != nil {
	//     c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save chunk"})
	//     return
	// }

	response := ChunkUploadResponse{
		UID: chunkUID,
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

func DownloadChunk(context *gin.Context) {
	// Get URL parameters
	itemUID := context.Param("item_uid")
	chunkUID := context.Param("chunk_uid")
	if itemUID == "" || chunkUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "missing item or chunk UID"})
		return
	}

	// TODO: Implement chunk retrieval
	// Example:
	// var chunk ChunkRef
	// err := db.Joins("Item").
	//     Joins("Item.Collection").
	//     Where("chunks.uid = ? AND items.uid = ? AND collections.owner = ?",
	//           chunkUID, itemUID, username).
	//     First(&chunk).Error
	// if err != nil {
	//     if errors.Is(err, gorm.ErrRecordNotFound) {
	//         c.JSON(http.StatusNotFound, gin.H{"error": "chunk not found"})
	//     } else {
	//         c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
	//     }
	//     return
	// }

	// Send chunk content
	// c.Data(http.StatusOK, "application/octet-stream", chunk.Content)

	context.Status(http.StatusNotImplemented)
}
