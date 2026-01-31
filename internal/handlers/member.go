package handlers

import (
	"fmt"
	"goatsync/internal/repository"
	"goatsync/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Types for member operations
type AccessLevel string

const (
	AccessLevelRead      AccessLevel = "read"
	AccessLevelReadWrite AccessLevel = "readwrite"
	AccessLevelAdmin     AccessLevel = "admin"
)

type CollectionMember struct {
	ID          uint        `json:"id" msgpack:"id"`
	UserID      string      `json:"userId" msgpack:"userId"`
	Username    string      `json:"username" msgpack:"username"`
	AccessLevel AccessLevel `json:"accessLevel" msgpack:"accessLevel"`
	Collection  string      `json:"collection" msgpack:"collection"`
	Stoken      string      `json:"stoken" msgpack:"stoken"`
}

type MemberListResponse struct {
	Data     []CollectionMemberOut `json:"data" msgpack:"data"`
	Iterator *string               `json:"iterator,omitempty" msgpack:"iterator,omitempty"`
	Done     bool                  `json:"done" msgpack:"done"`
}

type CollectionMemberOut struct {
	Username    string      `json:"username" msgpack:"username"`
	AccessLevel AccessLevel `json:"accessLevel" msgpack:"accessLevel"`
}

type ModifyAccessLevelRequest struct {
	AccessLevel AccessLevel `json:"accessLevel" msgpack:"accessLevel"`
}

// Handler implementations
func ListMembers(context *gin.Context) {
	limit := 50
	if limitParam := context.Query("limit"); limitParam != "" {
		_, _ = fmt.Sscanf(limitParam, "%d", &limit)
	}
	iterator := context.Query("iterator")

	// TODO: Implement database query
	members := []CollectionMember{}
	// Example query:
	// err := db.Where("collection_id = ?", collectionID).
	//     Order("id").
	//     Limit(limit + 1).
	//     Find(&members).Error

	done := true
	if len(members) > limit {
		done = false
		members = members[:limit]
	}

	// Convert to output format
	memberData := make([]CollectionMemberOut, len(members))
	for i, member := range members {
		memberData[i] = CollectionMemberOut{
			Username:    member.Username,
			AccessLevel: member.AccessLevel,
		}
	}

	response := MemberListResponse{
		Data: memberData,
		// Iterator: getNextIterator(members),
		Iterator: &iterator,
		Done:     done,
	}

	if err := utils.SendResponse(context, &response); err != nil {
		utils.InternalError(context, err)
	}
}

func DeleteMember(context *gin.Context) {
	// username := strings.ToLower(context.Param("username"))
	// collectionID := context.GetString("collection_id")

	// TODO: Implement member deletion
	// Example:
	// err := db.Transaction(func(tx *gorm.DB) error {
	//     return tx.Where("collection_id = ? AND username = ?",
	//         collectionID, username).Delete(&CollectionMember{}).Error
	// })

	context.Status(http.StatusNoContent)
}

func UpdateMemberAccess(context *gin.Context) {
	var request ModifyAccessLevelRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement access level update
	// username := strings.ToLower(context.Param("username"))
	// Example:
	// err := db.Transaction(func(tx *gorm.DB) error {
	//     member := CollectionMember{}
	//     if err := tx.Where("collection_id = ? AND username = ?",
	//         collectionID, username).First(&member).Error; err != nil {
	//         return err
	//     }
	//     member.AccessLevel = request.AccessLevel
	//     member.Stoken = uuid.New().String()
	//     return tx.Save(&member).Error
	// })

	context.Status(http.StatusNoContent)
}

func LeaveMember(c *gin.Context) {
	token := c.GetHeader("Authorization")
	tokens := repository.GetTokens()
	_, exists := tokens[token]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// TODO: Implement member leave
	// collectionID := c.GetString("collection_id")
	// Example:
	// err := db.Transaction(func(tx *gorm.DB) error {
	//     return tx.Where("collection_id = ? AND username = ?",
	//         collectionID, username).Delete(&CollectionMember{}).Error
	// })

	c.Status(http.StatusNoContent)
}
