package signin

import (
	"encoding/json"
	"net/http"
	"server-api-admin/config"
	"server-api-admin/util/middlewares"
	"server-api-admin/util/password"
	"server-api-admin/util/postgresdb"
	"time"

	"github.com/julienschmidt/httprouter"
)

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func signIn(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := r.Context().Value(config.UserIDKey).(string)
	if userID != "" {
		http.Error(w, "already signed in", http.StatusBadRequest)
		return
	}

	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	var passwordHash, salt string
	err = postgresdb.DB.QueryRowContext(
		r.Context(),
		`SELECT user_id, password_hash, salt FROM "user" WHERE email = $1 AND email_verified AND staff_permission > 0`,
		req.Email,
	).Scan(&userID, &passwordHash, &salt)
	if err != nil {
		http.Error(w, "Email or password is incorrect", http.StatusBadRequest)
		return
	}

	match, err := password.ComparePasswordWithHash(req.Password, passwordHash, salt)
	if err != nil || !match {
		http.Error(w, "Email or password is incorrect, or you are not staff", http.StatusBadRequest)
		return
	}

	var sessionID string
	var expiresAt time.Time
	err = postgresdb.DB.QueryRowContext(
		r.Context(),
		"INSERT INTO staff_sessions (user_id) VALUES ($1::uuid) RETURNING session_id, expires_at",
		userID,
	).Scan(&sessionID, &expiresAt)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	encryptedSessionID, err := middlewares.EncryptSessionID(sessionID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    encryptedSessionID,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
		// Secure:   true, // Set to true if using HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
}
