package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	dao "github.com/pbdeuchler/assistant-server/dao/postgres"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthConfig struct {
	GCloudClientID     string
	GCloudClientSecret string
	GCloudProjectID    string
	BaseURL            string
}

type authDAO interface {
	CreateCredentials(ctx context.Context, c dao.Credentials) (dao.Credentials, error)
	GetCredentialsByUserAndType(ctx context.Context, userUID, credentialType string) (dao.Credentials, error)
	UpdateCredentials(ctx context.Context, id string, c dao.Credentials) (dao.Credentials, error)
}

type AuthHandlers struct {
	oauth2Config *oauth2.Config
	jwtSecret    []byte
	dao          authDAO
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type TokenResponse struct {
	Token     string         `json:"token"`
	ExpiresAt time.Time      `json:"expires_at"`
	User      GoogleUserInfo `json:"user"`
}

func NewAuthHandlers(cfg AuthConfig, dao authDAO) http.Handler {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.GCloudClientID,
		ClientSecret: cfg.GCloudClientSecret,
		RedirectURL:  cfg.BaseURL + "/oauth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
			"https://www.googleapis.com/auth/calendar.app.created",
			"https://www.googleapis.com/auth/calendar.calendarlist.readonly",
			"https://www.googleapis.com/auth/calendar.calendars",
			"https://www.googleapis.com/auth/calendar.calendars.readonly",
			"https://www.googleapis.com/auth/calendar.events",
			"https://www.googleapis.com/auth/calendar.events.freebusy",
			"https://www.googleapis.com/auth/calendar.events.owned",
			"https://www.googleapis.com/auth/calendar.events.owned.readonly",
			"https://www.googleapis.com/auth/calendar.events.public.readonly",
			"https://www.googleapis.com/auth/calendar.events.readonly",
			"https://www.googleapis.com/auth/calendar.freebusy",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	h := &AuthHandlers{
		oauth2Config: oauth2Config,
		dao:          dao,
	}

	r := chi.NewRouter()
	r.Use(httpLogger())
	r.Get("/google", h.googleAuth)
	r.Get("/google/callback", h.googleCallback)
	return r
}

func (h *AuthHandlers) googleAuth(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
		return
	}

	state, err := generateRandomState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state and user_id in session/cookie for verification
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   r.URL.Scheme == "https",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userID,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   r.URL.Scheme == "https",
		SameSite: http.SameSiteLaxMode,
	})

	url := h.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandlers) googleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state parameter
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Get user_id from cookie
	userIDCookie, err := r.Cookie("user_id")
	if err != nil {
		http.Error(w, "user_id cookie not found", http.StatusBadRequest)
		return
	}
	userID := userIDCookie.Value

	// Clear the cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No authorization code", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for token
	ctx := context.Background()
	token, err := h.oauth2Config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("Google OAuth2 token exchange successful", "expiry", token.Expiry)

	// Get user info from Google
	userInfo, err := h.getUserInfo(ctx, token)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Persist the full OAuth token as JSON
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		http.Error(w, "Failed to marshal token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	credential := dao.Credentials{
		ID:             uuid.NewString(),
		UserUID:        userID,
		CredentialType: "GOOGLE_CALENDAR",
		Value:          tokenJSON,
	}

	// Try to get existing credential first
	existingCred, err := h.dao.GetCredentialsByUserAndType(ctx, userID, "GOOGLE_CALENDAR")
	if err == nil {
		// Update existing credential
		_, err = h.dao.UpdateCredentials(ctx, existingCred.ID, credential)
		if err != nil {
			http.Error(w, "Failed to update credential: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Create new credential
		_, err = h.dao.CreateCredentials(ctx, credential)
		if err != nil {
			http.Error(w, "Failed to create credential: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	slog.Info("Google OAuth2 credential saved", "user_id", userID, "user_email", userInfo.Email)

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"user":    userInfo,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandlers) getUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := h.oauth2Config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

//	func (h *AuthHandlers) generateJWT(userInfo *GoogleUserInfo) (string, time.Time, error) {
//		expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
//
//		claims := jwt.MapClaims{
//			"sub":   userInfo.ID,
//			"email": userInfo.Email,
//			"name":  userInfo.Name,
//			"exp":   expiresAt.Unix(),
//			"iat":   time.Now().Unix(),
//		}
//
//		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//		tokenString, err := token.SignedString(h.jwtSecret)
//		if err != nil {
//			return "", time.Time{}, err
//		}
//
//		return tokenString, expiresAt, nil
//	}
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// JWT Middleware for protecting routes
// func JWTMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			authHeader := r.Header.Get("Authorization")
// 			if authHeader == "" {
// 				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
// 				return
// 			}
//
// 			tokenString := ""
// 			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
// 				tokenString = authHeader[7:]
// 			} else {
// 				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
// 				return
// 			}
//
// 			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 				}
// 				return jwtSecret, nil
// 			})
//
// 			if err != nil || !token.Valid {
// 				http.Error(w, "Invalid token", http.StatusUnauthorized)
// 				return
// 			}
//
// 			if claims, ok := token.Claims.(jwt.MapClaims); ok {
// 				// Add user info to request context
// 				ctx := context.WithValue(r.Context(), "user_id", claims["sub"])
// 				ctx = context.WithValue(ctx, "user_email", claims["email"])
// 				ctx = context.WithValue(ctx, "user_name", claims["name"])
// 				r = r.WithContext(ctx)
// 			}
//
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

