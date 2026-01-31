// Package service provides business logic layer implementations.
package service

import (
	"context"
	"strings"
	"time"

	"goatsync/internal/config"
	"goatsync/internal/crypto"
	"goatsync/internal/model"
	"goatsync/internal/repository"
	pkgerrors "goatsync/pkg/errors"

	"github.com/vmihailenco/msgpack/v5"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	crypto    *crypto.Etebase
	cfg       *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		crypto:    crypto.NewEtebase(cfg.EncryptionSecret),
		cfg:       cfg,
	}
}

// LoginChallengeRequest is the request for login_challenge
type LoginChallengeRequest struct {
	Username string `msgpack:"username"`
}

// LoginChallengeResponse is the response for login_challenge
type LoginChallengeResponse struct {
	Salt      []byte `msgpack:"salt"`
	Challenge []byte `msgpack:"challenge"`
	Version   int    `msgpack:"version"`
}

// LoginChallenge generates a login challenge for a user.
// This matches the Python implementation in authentication.py:login_challenge
func (s *AuthService) LoginChallenge(ctx context.Context, username string) (*LoginChallengeResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Check if user exists
	if user == nil {
		return nil, pkgerrors.ErrUserNotFound
	}

	// Check if user has UserInfo (is properly initialized)
	if user.UserInfo == nil {
		return nil, pkgerrors.ErrUserNotInit
	}

	// Get salt from user info
	salt := user.UserInfo.Salt

	// Derive encryption key using BLAKE2b with salt and personalization
	encKey, err := s.crypto.GetEncryptionKey(salt)
	if err != nil {
		return nil, err
	}

	// Create challenge data
	// Python: challenge_data = {"timestamp": int(datetime.now().timestamp()), "userId": user.id}
	challengeData := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"userId":    user.ID,
	}

	// Encode challenge data with msgpack
	challengeBytes, err := msgpack.Marshal(challengeData)
	if err != nil {
		return nil, err
	}

	// Encrypt challenge with SecretBox
	encryptedChallenge, err := s.crypto.Encrypt(encKey, challengeBytes)
	if err != nil {
		return nil, err
	}

	return &LoginChallengeResponse{
		Salt:      salt,
		Challenge: encryptedChallenge,
		Version:   user.UserInfo.Version,
	}, nil
}

// LoginRequest is the request for login
type LoginRequest struct {
	Response  []byte `msgpack:"response"`
	Signature []byte `msgpack:"signature"`
}

// LoginResponseData is the decrypted response data from the client
type LoginResponseData struct {
	Username  string `msgpack:"username"`
	Challenge []byte `msgpack:"challenge"`
	Host      string `msgpack:"host"`
	Action    string `msgpack:"action"`
}

// UserResponse is the user data returned after login
type UserResponse struct {
	Username         string `msgpack:"username"`
	Email            string `msgpack:"email"`
	Pubkey           []byte `msgpack:"pubkey"`
	EncryptedContent []byte `msgpack:"encryptedContent"`
}

// LoginOut is the response for login
type LoginOut struct {
	Token string       `msgpack:"token"`
	User  UserResponse `msgpack:"user"`
}

// Login validates the login request and returns a token.
// This matches the Python implementation in authentication.py:login
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, host string) (*LoginOut, error) {
	// Decode the response to get username
	var responseData LoginResponseData
	if err := msgpack.Unmarshal(req.Response, &responseData); err != nil {
		return nil, pkgerrors.ErrInvalidRequest.WithDetail("Failed to decode login response")
	}

	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, responseData.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, pkgerrors.ErrUserNotFound
	}
	if user.UserInfo == nil {
		return nil, pkgerrors.ErrUserNotInit
	}

	// Validate the login request
	if err := s.validateLoginRequest(ctx, &responseData, req, user, "login", host); err != nil {
		return nil, err
	}

	// Create auth token
	token := &model.AuthToken{
		UserID: user.ID,
	}
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	// Return login response
	return &LoginOut{
		Token: token.Key,
		User: UserResponse{
			Username:         user.FirstName, // Original username casing
			Email:            user.Email,
			Pubkey:           user.UserInfo.Pubkey,
			EncryptedContent: user.UserInfo.EncryptedContent,
		},
	}, nil
}

// validateLoginRequest validates the login request against the challenge
func (s *AuthService) validateLoginRequest(
	ctx context.Context,
	responseData *LoginResponseData,
	req *LoginRequest,
	user *model.User,
	expectedAction string,
	hostFromRequest string,
) error {
	// Derive encryption key
	encKey, err := s.crypto.GetEncryptionKey(user.UserInfo.Salt)
	if err != nil {
		return err
	}

	// Decrypt the challenge from the response
	challengeBytes, err := s.crypto.Decrypt(encKey, responseData.Challenge)
	if err != nil {
		return pkgerrors.ErrBadSignature
	}

	// Decode challenge data
	var challengeData map[string]interface{}
	if err := msgpack.Unmarshal(challengeBytes, &challengeData); err != nil {
		return pkgerrors.ErrInvalidRequest.WithDetail("Failed to decode challenge data")
	}

	// Validate action
	if responseData.Action != expectedAction {
		return pkgerrors.NewWrongActionError(expectedAction)
	}

	// Validate timestamp (challenge expiry)
	timestamp, ok := toInt64(challengeData["timestamp"])
	if !ok {
		return pkgerrors.ErrInvalidRequest.WithDetail("Invalid timestamp in challenge")
	}

	now := time.Now().Unix()
	if now-timestamp > int64(s.cfg.ChallengeValidSeconds) {
		return pkgerrors.ErrChallengeExpired
	}

	// Validate userId
	userID, ok := toUint64(challengeData["userId"])
	if !ok {
		return pkgerrors.ErrInvalidRequest.WithDetail("Invalid userId in challenge")
	}

	if uint(userID) != user.ID {
		return pkgerrors.ErrWrongUser
	}

	// Validate host (skip in debug mode)
	if !s.cfg.Debug {
		expectedHost := strings.Split(hostFromRequest, ":")[0]
		gotHost := strings.Split(responseData.Host, ":")[0]
		if gotHost != expectedHost {
			return pkgerrors.NewWrongHostError(expectedHost, gotHost)
		}
	}

	// Verify Ed25519 signature
	// The client signs the response with their loginPubkey
	if err := crypto.VerifySignature(user.UserInfo.LoginPubkey, req.Response, req.Signature); err != nil {
		return pkgerrors.ErrBadSignature
	}

	return nil
}

// Logout invalidates a token
func (s *AuthService) Logout(ctx context.Context, tokenKey string) error {
	return s.tokenRepo.Delete(ctx, tokenKey)
}

// SignupRequest is the request for signup
type SignupRequest struct {
	User             SignupUser `msgpack:"user"`
	Salt             []byte     `msgpack:"salt"`
	LoginPubkey      []byte     `msgpack:"loginPubkey"`
	Pubkey           []byte     `msgpack:"pubkey"`
	EncryptedContent []byte     `msgpack:"encryptedContent"`
}

// SignupUser contains user data for signup
type SignupUser struct {
	Username string `msgpack:"username"`
	Email    string `msgpack:"email"`
}

// Signup creates a new user account
func (s *AuthService) Signup(ctx context.Context, req *SignupRequest) (*LoginOut, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByUsername(ctx, req.User.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, pkgerrors.ErrUserExists
	}

	// Also check by email
	existingUser, err = s.userRepo.GetByEmail(ctx, req.User.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, pkgerrors.ErrUserExists
	}

	// Create user
	user := &model.User{
		Username:  req.User.Username,
		Email:     req.User.Email,
		FirstName: req.User.Username, // Store original casing
		IsActive:  true,
	}

	userInfo := &model.UserInfo{
		Version:          1,
		Salt:             req.Salt,
		LoginPubkey:      req.LoginPubkey,
		Pubkey:           req.Pubkey,
		EncryptedContent: req.EncryptedContent,
	}

	if err := s.userRepo.Create(ctx, user, userInfo); err != nil {
		return nil, err
	}

	// Create auth token
	token := &model.AuthToken{
		UserID: user.ID,
	}
	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	return &LoginOut{
		Token: token.Key,
		User: UserResponse{
			Username:         req.User.Username,
			Email:            req.User.Email,
			Pubkey:           req.Pubkey,
			EncryptedContent: req.EncryptedContent,
		},
	}, nil
}

// ChangePasswordRequest is the request for change_password
type ChangePasswordRequest struct {
	Response  []byte `msgpack:"response"`
	Signature []byte `msgpack:"signature"`
}

// ChangePasswordResponseData is the decrypted response data for password change
type ChangePasswordResponseData struct {
	Username         string `msgpack:"username"`
	Challenge        []byte `msgpack:"challenge"`
	Host             string `msgpack:"host"`
	Action           string `msgpack:"action"`
	LoginPubkey      []byte `msgpack:"loginPubkey"`
	EncryptedContent []byte `msgpack:"encryptedContent"`
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, user *model.User, req *ChangePasswordRequest, host string) error {
	// Decode the response
	var responseData ChangePasswordResponseData
	if err := msgpack.Unmarshal(req.Response, &responseData); err != nil {
		return pkgerrors.ErrInvalidRequest.WithDetail("Failed to decode change password response")
	}

	// Validate the request (similar to login)
	loginReq := &LoginRequest{
		Response:  req.Response,
		Signature: req.Signature,
	}
	loginResponseData := &LoginResponseData{
		Username:  responseData.Username,
		Challenge: responseData.Challenge,
		Host:      responseData.Host,
		Action:    responseData.Action,
	}

	if err := s.validateLoginRequest(ctx, loginResponseData, loginReq, user, "changePassword", host); err != nil {
		return err
	}

	// Update user info with new credentials
	user.UserInfo.LoginPubkey = responseData.LoginPubkey
	user.UserInfo.EncryptedContent = responseData.EncryptedContent

	return s.userRepo.UpdateUserInfo(ctx, user.UserInfo)
}

// GetUserByToken retrieves a user by auth token
func (s *AuthService) GetUserByToken(ctx context.Context, tokenKey string) (*model.User, error) {
	token, err := s.tokenRepo.GetByKey(ctx, tokenKey)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, pkgerrors.ErrInvalidToken
	}

	return token.User, nil
}

// toInt64 converts a msgpack-decoded value to int64.
// msgpack may decode integers as various types depending on value size.
func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int32:
		return int64(n), true
	case int:
		return int64(n), true
	case uint64:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint:
		return int64(n), true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}

// toUint64 converts a msgpack-decoded value to uint64.
// msgpack may decode integers as various types depending on value size.
func toUint64(v interface{}) (uint64, bool) {
	switch n := v.(type) {
	case uint64:
		return n, true
	case uint32:
		return uint64(n), true
	case uint:
		return uint64(n), true
	case int64:
		return uint64(n), true
	case int32:
		return uint64(n), true
	case int:
		return uint64(n), true
	case float64:
		return uint64(n), true
	default:
		return 0, false
	}
}

