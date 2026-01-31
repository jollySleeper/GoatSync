# End-to-End Encryption (E2EE) in GoatSync

> ⚠️ **NOTE**: This document is NOT committed to git. It's a personal learning guide.

---

## Table of Contents

1. [Quick Answer: HTTP vs HTTPS](#quick-answer-http-vs-https)
2. [What E2EE Protects](#what-e2ee-protects)
3. [What HTTPS Protects](#what-https-protects)
4. [Risk Assessment](#risk-assessment)
5. [How E2EE Works in GoatSync](#how-e2ee-works-in-goatsync)
6. [Code Walkthrough](#code-walkthrough)
7. [Crypto Algorithms Explained](#crypto-algorithms-explained)
8. [Data Flow Examples](#data-flow-examples)

---

## Quick Answer: HTTP vs HTTPS

### Will E2EE work without HTTPS?

**YES!** E2EE is completely independent of HTTPS.

| Scenario | E2EE | HTTPS | Security Level |
|----------|------|-------|----------------|
| Local network only | ✅ | ❌ | **Good** - Your data is encrypted |
| Over internet | ✅ | ❌ | **Risky** - Auth tokens exposed |
| Local network | ✅ | ✅ | **Better** - Defense in depth |
| Over internet | ✅ | ✅ | **Best** - Full protection |

### What's at Risk Without HTTPS?

| What | Risk Level | Why |
|------|------------|-----|
| Your calendar events/contacts | ✅ **SAFE** | Encrypted before leaving your device |
| Your password (during login) | ⚠️ **MEDIUM** | Challenge-response reduces exposure |
| Your auth token | 🔴 **EXPOSED** | Sent in every request header |
| Metadata (when you sync) | 🔴 **EXPOSED** | Attacker knows you're syncing |
| Which collections you access | 🔴 **EXPOSED** | URL paths are visible |

### My Recommendation

| Use Case | HTTP OK? | Why |
|----------|----------|-----|
| **Home network only** | ✅ Yes | Trusted network, low risk |
| **Tailscale/WireGuard VPN** | ✅ Yes | VPN provides encryption |
| **Public internet** | ❌ No | Use HTTPS or VPN |
| **Coffee shop WiFi** | ❌❌ No | Definitely use HTTPS |

---

## What E2EE Protects

### Protected by E2EE (Safe without HTTPS)

```
✅ Calendar event titles
✅ Calendar event descriptions
✅ Calendar event locations
✅ Calendar event attendees
✅ Contact names
✅ Contact phone numbers
✅ Contact emails
✅ Contact addresses
✅ Task descriptions
✅ Notes content
✅ Collection names
✅ Everything inside your data
```

### How It Works (Simplified)

```
Your Device                          Server
───────────                          ──────

"Meeting with                        
 Alice at 3pm"                       
      │                              
      ▼                              
┌─────────────┐                      
│  Encrypt    │  Your encryption     
│  with YOUR  │  key never leaves    
│  key        │  your device!        
└─────────────┘                      
      │                              
      ▼                              
"xK8j2mN9pL..."  ──────────────────► "xK8j2mN9pL..."
 (encrypted)                          (still encrypted)
                                     
                                      Server can only see:
                                      • Encrypted blob
                                      • Size of blob
                                      • When it was saved
                                      
                                      Server CANNOT see:
                                      • "Meeting with Alice"
                                      • "3pm"
                                      • Anything inside
```

---

## What HTTPS Protects

### Protected by HTTPS (NOT protected without it)

```
🔴 Auth token in headers → Anyone on network can steal your session
🔴 API endpoints you call → Attacker knows you're syncing calendars
🔴 Request/response sizes → Can infer activity patterns
🔴 Your IP address → Identifies you (though server sees this anyway)
```

### What an Attacker on Your Network Could Do WITHOUT HTTPS

1. **Steal your auth token** → Impersonate you
2. **See your sync patterns** → Know when you're active
3. **Modify responses** → Inject fake encrypted blobs (won't decrypt, but DoS)
4. **See collection UIDs** → Know which calendars you access

### What an Attacker CANNOT Do Even Without HTTPS

1. **Read your calendar events** → Encrypted with your key
2. **Read your contacts** → Encrypted with your key
3. **Modify your data** → Would fail signature verification
4. **Decrypt anything** → Key never sent over network

---

## Risk Assessment

### Scenario 1: Home Network, HTTP Only

```
Risk Level: LOW ✅

┌─────────────────────────────────────────────────────────┐
│  Your Home Network                                       │
│                                                          │
│   Phone ─────┐                                          │
│              │    All traffic stays                     │
│   Laptop ────┼────► inside your network ────► GoatSync  │
│              │                              (Raspberry  │
│   Desktop ───┘                               Pi, NAS)   │
│                                                          │
│   Who could attack?                                     │
│   • Someone on your WiFi (family, guests)              │
│   • Your ISP (if they inspect LAN traffic - rare)      │
│                                                          │
│   Your data is still encrypted!                         │
└─────────────────────────────────────────────────────────┘
```

### Scenario 2: VPN (Tailscale/WireGuard), HTTP Only

```
Risk Level: LOW ✅

┌─────────────────────────────────────────────────────────┐
│                                                          │
│   Phone ─────┐      VPN encrypts                        │
│              │      everything        ┌──────────────┐  │
│   Laptop ────┼─────► ═══════════════► │  Your VPS    │  │
│              │      WireGuard tunnel  │  GoatSync    │  │
│                                       └──────────────┘  │
│                                                          │
│   VPN provides the encryption that HTTPS would!         │
│   HTTP inside VPN tunnel = effectively HTTPS            │
└─────────────────────────────────────────────────────────┘
```

### Scenario 3: Public Internet, HTTP Only

```
Risk Level: HIGH 🔴

┌─────────────────────────────────────────────────────────┐
│                                                          │
│   Phone ─────► Coffee Shop WiFi ─────► Internet ────►   │
│                      │                                   │
│                      ▼                                   │
│               ┌─────────────┐                           │
│               │  Attacker   │                           │
│               │  Can see:   │                           │
│               │  • Auth token                           │
│               │  • API calls                            │
│               │  • Timing                               │
│               └─────────────┘                           │
│                                                          │
│   Your DATA is still encrypted!                         │
│   But attacker can IMPERSONATE you with stolen token!   │
└─────────────────────────────────────────────────────────┘
```

---

## How E2EE Works in GoatSync

### The Big Picture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         E2EE ARCHITECTURE                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  CLIENT SIDE (Your Device)              SERVER SIDE (GoatSync)              │
│  ════════════════════════               ════════════════════════            │
│                                                                              │
│  Your Password                          Never knows your password!          │
│       │                                                                      │
│       ▼                                                                      │
│  ┌─────────────┐                                                            │
│  │   Argon2    │  Key derivation                                           │
│  │  (in app)   │  happens on device                                        │
│  └─────────────┘                                                            │
│       │                                                                      │
│       ▼                                                                      │
│  Master Key                             Never sent to server!               │
│       │                                                                      │
│       ├──────────────┐                                                      │
│       │              │                                                      │
│       ▼              ▼                                                      │
│  Login Key      Encryption Key                                              │
│  (Ed25519)      (for data)                                                  │
│       │              │                                                      │
│       │              │                                                      │
│       │              └─────────────────┐                                    │
│       │                                │                                    │
│       ▼                                ▼                                    │
│  ┌─────────────┐               ┌─────────────┐                             │
│  │   Sign      │               │   Encrypt   │                             │
│  │  Challenge  │               │    Data     │                             │
│  └─────────────┘               └─────────────┘                             │
│       │                                │                                    │
│       │ signature                      │ encrypted blob                     │
│       │                                │                                    │
│       ▼                                ▼                                    │
│  ─────────────────────────────────────────────────────────────────────►    │
│                              NETWORK                                        │
│  ◄─────────────────────────────────────────────────────────────────────    │
│                                                                              │
│                                        │                                    │
│                                        ▼                                    │
│                                ┌─────────────────┐                          │
│                                │    GoatSync     │                          │
│                                │    Database     │                          │
│                                │                 │                          │
│                                │  Stores only:   │                          │
│                                │  • Encrypted    │                          │
│                                │    blobs        │                          │
│                                │  • Public keys  │                          │
│                                │  • Metadata     │                          │
│                                │                 │                          │
│                                │  Cannot see:    │                          │
│                                │  • Your data    │                          │
│                                │  • Your keys    │                          │
│                                └─────────────────┘                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Types in EteSync/GoatSync

| Key | Purpose | Stored Where | Who Knows It |
|-----|---------|--------------|--------------|
| **Password** | User authentication | Your brain | Only you |
| **Master Key** | Derives other keys | Never stored | Only your device (in memory) |
| **Login Key** (Ed25519 private) | Sign login challenges | Client only | Only your device |
| **Login Pubkey** (Ed25519 public) | Verify signatures | Server | Server (for verification) |
| **Encryption Key** | Encrypt collection keys | Client only | Only your device |
| **Collection Key** | Encrypt items in collection | Encrypted on server | Shared with members |
| **Account Encryption Key** | Master key for your account | Encrypted on server | Only you |

### What's Stored on Server (All Encrypted!)

```sql
-- User info (from GoatSync database)
django_userinfo:
  owner_id: 1
  login_pubkey: <Ed25519 PUBLIC key - for signature verification>
  pubkey: <Your PUBLIC key - for sharing>
  encrypted_content: <Your ENCRYPTED account data>
  salt: <For key derivation>

-- Collection (encrypted!)
django_collection:
  uid: "abc123..."
  -- Server CANNOT read content!

-- Items (encrypted!)
django_collectionitem:
  uid: "xyz789..."
  encryption_key: <ENCRYPTED collection key>
  -- Server CANNOT read content!

-- Revisions (encrypted!)
django_collectionitemrevision:
  meta: <ENCRYPTED metadata>
  -- Server CANNOT read content!
```

---

## Code Walkthrough

### 1. Key Derivation (internal/crypto/etebase.go)

```go
// GetEncryptionKey derives an encryption key using BLAKE2b
// This is called during login/signup to derive the server-side encryption key
func (e *Etebase) GetEncryptionKey(salt []byte) ([32]byte, error) {
    // Step 1: Hash the server's secret key
    // This creates a key from the server's ENCRYPTION_SECRET env variable
    key := blake2b.Sum256([]byte(e.secretKey))
    
    // Step 2: Use BLAKE2b with personalization
    // The "etebase-auth" personalization ensures this key is only for auth
    h, err := blake2b.New(&blake2b.Config{
        Size:   32,                          // 256-bit output
        Key:    key[:],                      // Keyed hash
        Salt:   salt[:16],                   // First 16 bytes of salt
        Person: []byte("etebase-auth"),      // Personalization string
    })
    
    // This produces a deterministic key that:
    // - Is unique per user (different salt)
    // - Is tied to this server (ENCRYPTION_SECRET)
    // - Can only be used for authentication ("etebase-auth")
}
```

### 2. Challenge-Response Login (internal/service/auth.go)

```go
// LoginChallenge creates an encrypted challenge for the client
func (s *AuthService) LoginChallenge(ctx context.Context, username string) (*LoginChallengeResponse, error) {
    // Step 1: Get user's salt (stored during signup)
    user, err := s.userRepo.GetByUsername(ctx, username)
    salt := user.UserInfo.Salt
    
    // Step 2: Derive encryption key for this challenge
    // Same key derivation as client will do
    encKey, _ := s.crypto.GetEncryptionKey(salt)
    
    // Step 3: Create challenge data
    challengeData := ChallengeData{
        Timestamp: time.Now().Unix(),
        UserID:    user.ID,
    }
    
    // Step 4: Encrypt challenge with SecretBox
    // Only someone with the same key (derived from password) can decrypt
    encrypted := s.crypto.Encrypt(encKey, msgpack(challengeData))
    
    return &LoginChallengeResponse{
        Salt:      salt,              // Client uses this to derive key
        Challenge: encrypted,         // Client must decrypt and sign
        Version:   user.UserInfo.Version,
    }
}

// Login verifies the client's response to the challenge
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
    user, _ := s.userRepo.GetByUsername(ctx, req.Username)
    
    // Step 1: Get user's public login key (Ed25519)
    loginPubkey := user.UserInfo.LoginPubkey  // 32 bytes
    
    // Step 2: Verify Ed25519 signature
    // Client signed: challenge response with their PRIVATE key
    // We verify with their PUBLIC key
    err := crypto.VerifySignature(loginPubkey, req.Challenge, req.Signature)
    if err != nil {
        return nil, errors.ErrBadSignature  // Wrong password!
    }
    
    // Step 3: Check challenge is valid (not expired, right user)
    // ... validation ...
    
    // Step 4: Create auth token
    token := generateSecureToken()  // Random 64-char token
    s.tokenRepo.Create(ctx, user.ID, token)
    
    return &LoginResult{Token: token}, nil
}
```

### 3. Signature Verification (internal/crypto/etebase.go)

```go
// VerifySignature verifies an Ed25519 signature
func VerifySignature(publicKey, message, signature []byte) error {
    // Ed25519 signature verification
    // - publicKey: 32 bytes (stored on server during signup)
    // - message: the challenge response
    // - signature: 64 bytes (created by client with private key)
    
    if len(publicKey) != ed25519.PublicKeySize {
        return errors.New("invalid public key size")
    }
    
    if len(signature) != ed25519.SignatureSize {
        return errors.New("invalid signature size")
    }
    
    // Verify returns true if signature is valid
    if !ed25519.Verify(publicKey, message, signature) {
        return errors.ErrBadSignature
    }
    
    return nil
}
```

### 4. SecretBox Encryption (internal/crypto/etebase.go)

```go
// Encrypt encrypts data using NaCl SecretBox (XSalsa20-Poly1305)
func (e *Etebase) Encrypt(key [32]byte, plaintext []byte) []byte {
    // Step 1: Generate random nonce (24 bytes)
    var nonce [24]byte
    rand.Read(nonce[:])
    
    // Step 2: Encrypt with SecretBox
    // XSalsa20 for encryption + Poly1305 for authentication
    ciphertext := secretbox.Seal(nonce[:], plaintext, &nonce, &key)
    
    // Output: nonce (24) + ciphertext (len(plaintext) + 16 for auth tag)
    return ciphertext
}

// Decrypt decrypts data using NaCl SecretBox
func (e *Etebase) Decrypt(key [32]byte, ciphertext []byte) ([]byte, error) {
    // Step 1: Extract nonce from beginning
    var nonce [24]byte
    copy(nonce[:], ciphertext[:24])
    
    // Step 2: Decrypt
    plaintext, ok := secretbox.Open(nil, ciphertext[24:], &nonce, &key)
    if !ok {
        return nil, errors.New("decryption failed")
    }
    
    return plaintext, nil
}
```

---

## Crypto Algorithms Explained

### 1. BLAKE2b (Key Derivation)

```
What: A cryptographic hash function (like SHA-256 but better)
Why:  Fast, secure, supports keyed hashing and personalization
Used: Deriving encryption keys from passwords/secrets

BLAKE2b Features Used:
┌─────────────────────────────────────────────────────────────┐
│  BLAKE2b(                                                    │
│    message = "",           // Can be empty                   │
│    key = server_secret,    // Makes output unique to server  │
│    salt = user_salt,       // Makes output unique to user    │
│    person = "etebase-auth" // Makes output unique to purpose │
│  )                                                           │
│                                                              │
│  Output: 32 bytes of deterministic but unpredictable data   │
└─────────────────────────────────────────────────────────────┘
```

### 2. Ed25519 (Authentication)

```
What: Elliptic curve digital signature algorithm
Why:  Fast, secure, 64-byte signatures, 32-byte keys
Used: Proving you know your password without sending it

How Login Works:
┌─────────────────────────────────────────────────────────────┐
│  CLIENT                           SERVER                     │
│                                                              │
│  1. Request challenge ──────────────────────►               │
│                                                              │
│  2. ◄────────────── Send encrypted challenge                │
│     (encrypted with key derived from password)               │
│                                                              │
│  3. Decrypt challenge (proves you know password)            │
│     Sign with Ed25519 PRIVATE key                            │
│                                                              │
│  4. Send signature ──────────────────────────►              │
│                                                              │
│  5. Verify with Ed25519 PUBLIC key                          │
│     (stored during signup)                                   │
│                                                              │
│  Result: Server knows you have the private key              │
│          without ever seeing it!                             │
└─────────────────────────────────────────────────────────────┘

Password is NEVER sent over the network!
```

### 3. XSalsa20-Poly1305 (SecretBox)

```
What: Authenticated encryption (encrypt + verify integrity)
Why:  Fast, secure, prevents tampering
Used: Encrypting challenges, encrypting data

SecretBox = XSalsa20 (encryption) + Poly1305 (authentication)
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│  Input:  plaintext + 32-byte key + 24-byte nonce            │
│                                                              │
│  XSalsa20:                                                   │
│  ─────────                                                   │
│  • Stream cipher (generates keystream from key+nonce)        │
│  • XOR plaintext with keystream = ciphertext                 │
│  • Fast, secure, no patterns                                 │
│                                                              │
│  Poly1305:                                                   │
│  ─────────                                                   │
│  • Creates 16-byte authentication tag                        │
│  • If anyone modifies ciphertext, tag won't match            │
│  • Prevents tampering                                        │
│                                                              │
│  Output: nonce (24) + ciphertext (N) + auth tag (16)        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Data Flow Examples

### Example 1: User Signup

```
CLIENT                                              SERVER
──────                                              ──────

1. User enters: username, email, password

2. Client generates:
   • Ed25519 keypair (loginPubkey, loginPrivkey)
   • Another keypair for encryption (pubkey, privkey)
   • Random salt (16 bytes)
   • Encrypts account data with derived key
   
3. Client sends to server:
   ┌─────────────────────────────────────────┐
   │ {                                        │
   │   username: "alice",                     │
   │   email: "alice@example.com",            │
   │   salt: <16 random bytes>,               │
   │   loginPubkey: <Ed25519 public key>,     │  ◄─ Server stores this
   │   pubkey: <Encryption public key>,       │  ◄─ For sharing later
   │   encryptedContent: <Encrypted data>     │  ◄─ Server can't read
   │ }                                        │
   └─────────────────────────────────────────┘

4. Server stores in database:
   • loginPubkey → For verifying login signatures
   • pubkey → For member invitations
   • encryptedContent → Client's encrypted account data
   • salt → Sent back during login challenge
   
   Server NEVER has:
   • Password
   • Private keys
   • Decryption keys
```

### Example 2: Login Flow

```
CLIENT                                              SERVER
──────                                              ──────

1. "I want to login as alice"
   ─────────────────────────────────────────────────────►
   
2.                                   ◄─────────────────────
   Server sends:
   {
     salt: <alice's salt>,
     challenge: <encrypted blob>,
     version: 1
   }
   
3. Client:
   a. Derives key using password + salt
   b. Decrypts challenge (if password wrong, fails here!)
   c. Signs decrypted challenge with Ed25519 private key
   
4. Client sends signature
   ─────────────────────────────────────────────────────►
   {
     username: "alice",
     challenge: <decrypted challenge>,
     signature: <Ed25519 signature>
   }
   
5. Server:
   a. Gets alice's loginPubkey from database
   b. Verifies signature matches
   c. Creates auth token
   
6.                                   ◄─────────────────────
   { token: "abc123..." }
   
KEY INSIGHT: Password NEVER sent! Signature proves knowledge.
```

### Example 3: Syncing Calendar Event

```
CLIENT                                              SERVER
──────                                              ──────

1. User creates event: "Meeting with Alice at 3pm"

2. Client encrypts:
   ┌──────────────────────────────────────────────────────┐
   │  Plaintext:                                          │
   │  { title: "Meeting with Alice at 3pm",               │
   │    start: "2024-01-15T15:00:00",                     │
   │    ... }                                             │
   │                                                      │
   │  ↓ Encrypt with collection key                      │
   │                                                      │
   │  Ciphertext: "xK8j2mN9pL3qR7..."  (unreadable blob) │
   └──────────────────────────────────────────────────────┘

3. Client sends to server:
   POST /api/v1/collection/abc123/item/transaction/
   Authorization: Token xyz789
   
   {
     items: [{
       uid: "event123",
       content: "xK8j2mN9pL3qR7..."  ◄─ Encrypted!
     }]
   }

4. Server stores encrypted blob
   • Cannot read "Meeting with Alice"
   • Cannot read "3pm"
   • Only sees: "xK8j2mN9pL3qR7..."

5. Later, another device syncs:
   GET /api/v1/collection/abc123/item/
   
   Server returns: { content: "xK8j2mN9pL3qR7..." }
   
   Client decrypts with same collection key
   → "Meeting with Alice at 3pm"
```

---

## HTTP vs HTTPS Summary

### What's Safe Without HTTPS (Thanks to E2EE)

| Data | Protection | Safe? |
|------|------------|-------|
| Calendar events | E2EE encrypted | ✅ Yes |
| Contact details | E2EE encrypted | ✅ Yes |
| Task content | E2EE encrypted | ✅ Yes |
| Collection names | E2EE encrypted | ✅ Yes |
| Your password | Never sent (Ed25519) | ✅ Yes |

### What's Exposed Without HTTPS

| Data | Risk | Mitigation |
|------|------|------------|
| Auth token | Session hijacking | Use short-lived tokens |
| API endpoints called | Activity tracking | VPN or HTTPS |
| Request timing | Pattern analysis | VPN or HTTPS |
| Your IP address | Location tracking | VPN or Tor |

### Recommendations

```
┌─────────────────────────────────────────────────────────────┐
│                     DECISION TREE                            │
│                                                              │
│  Where are you accessing from?                              │
│                                                              │
│  ┌─────────────────┐                                        │
│  │ Home network    │ ──► HTTP is fine ✅                   │
│  │ only?           │     Your data is encrypted anyway      │
│  └─────────────────┘                                        │
│                                                              │
│  ┌─────────────────┐                                        │
│  │ Via VPN         │ ──► HTTP is fine ✅                   │
│  │ (Tailscale,     │     VPN encrypts the transport        │
│  │  WireGuard)?    │                                        │
│  └─────────────────┘                                        │
│                                                              │
│  ┌─────────────────┐                                        │
│  │ Public internet │ ──► Use HTTPS ⚠️                      │
│  │ without VPN?    │     Or set up Tailscale (free!)       │
│  └─────────────────┘                                        │
│                                                              │
│  ┌─────────────────┐                                        │
│  │ Paranoid about  │ ──► HTTPS + VPN + Tor 🔒              │
│  │ metadata?       │     Maximum protection                 │
│  └─────────────────┘                                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## TL;DR

1. **E2EE works without HTTPS** - Your data is encrypted before leaving your device
2. **Password never sent** - Ed25519 signature proves you know it
3. **Server is blind** - Cannot read your calendars, contacts, or tasks
4. **HTTP risk is auth token** - Someone could steal your session
5. **For home use, HTTP is fine** - Your data is still encrypted
6. **For public internet, use HTTPS or VPN** - Protect your session

**GoatSync's E2EE means even if someone hacks the server, they only get encrypted blobs they can't decrypt!** 🔒

