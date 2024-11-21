package middlewares

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"server-api-admin/config"
	"server-api-admin/util/postgresdb"
	"time"

	"github.com/julienschmidt/httprouter"
)

// CustomResponseWriter is a custom implementation of http.ResponseWriter that captures the response body and headers
type CustomResponseWriter struct {
	http.ResponseWriter
	body           bytes.Buffer
	statusCode     int
	headersWritten bool
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	if !w.headersWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(b)
}

func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	if !w.headersWritten {
		w.statusCode = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
		w.headersWritten = true
	}
}

// Middleware function to handle incoming requests
func Middleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		var sessionID, encryptedSessionID, userID string
		var expiresAt time.Time
		var staffPermission int
		var signedIn bool

		// Extract session ID from request
		sessionID = extractSessionID(r)

		// If session ID is missing, create a new one
		if sessionID != "" {
			// Start a transaction
			tx, err := postgresdb.DB.BeginTx(ctx, nil)
			if err != nil {
				http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
				return
			}
			defer tx.Rollback()

			// Decrypt session ID
			decryptedSessionID, err := DecryptSessionID(sessionID)
			if err != nil {
				http.Error(w, "Invalid session ID", http.StatusUnauthorized)
				return
			}
			sessionID = decryptedSessionID

			// Get user ID and cart ID from session in a single query
			userID, staffPermission, err = getUserIDFromSession(ctx, tx, sessionID)
			if err != nil {
				http.Error(w, "Failed to get user ID or cart ID from session", http.StatusInternalServerError)
				return
			}

			signedIn = true

			// Extend session expiry based on sliding_expiration
			encryptedSessionID, expiresAt, err = extendSessionExpiry(ctx, tx, sessionID)
			if err != nil {
				http.Error(w, "Failed to extend session expiry", http.StatusInternalServerError)
				return
			}

			// Commit the transaction
			err = tx.Commit()
			if err != nil {
				http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
				return
			}
		}
		// Add user ID to the request context
		ctx = context.WithValue(ctx, config.UserIDKey, userID)

		// Create a custom response writer to capture the response
		cw := &CustomResponseWriter{ResponseWriter: w}

		// Call the next handler with the updated context and custom response writer
		next(cw, r.WithContext(ctx), ps)

		// Combine response data and session info into a single JSON object
		var combinedResponse map[string]interface{}
		if cw.body.Len() > 0 {
			// Try to unmarshal the existing body into a map
			if err := json.Unmarshal(cw.body.Bytes(), &combinedResponse); err != nil {
				http.Error(w, "Failed to unmarshal response", http.StatusInternalServerError)
				return
			}
		} else {
			combinedResponse = make(map[string]interface{})
		}

		if cw.statusCode == 200 && r.Header.Get("X-GET-STAFF-RIGHTS") != "false" {
			// Serialize cart content and product details to JSON
			response := struct {
				StaffPermission int  `json:"staffPermission"`
				SignedIn        bool `json:"signedIn"`
			}{
				StaffPermission: staffPermission,
				SignedIn:        signedIn,
			}

			// Append the serialized JSON to the combined response
			combinedResponse["staffPermission"] = response.StaffPermission
			combinedResponse["signedIn"] = response.SignedIn
		}

		// Set the session cookie or return the session info in the response body
		if r.Header.Get("X-Renew-Session") != "false" && sessionID != "" && cw.statusCode == 200 {
			if isClientRequest(r) {
				http.SetCookie(w, &http.Cookie{
					Name:     "sessionID",
					Value:    encryptedSessionID,
					Expires:  expiresAt,
					HttpOnly: true,
					Path:     "/",
					SameSite: http.SameSiteStrictMode,
				})
			} else {
				sessionInfo := map[string]interface{}{
					"sessionID": encryptedSessionID,
					"expiresAt": expiresAt.UnixMilli(),
				}
				for k, v := range sessionInfo {
					combinedResponse[k] = v
				}
			}
		}

		// Marshal the combined response to JSON
		finalResponseJSON, err := json.Marshal(combinedResponse)
		if err != nil {
			http.Error(w, "Failed to serialize final response", http.StatusInternalServerError)
			return
		}

		// Write the final response to the client
		if !cw.headersWritten {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(cw.statusCode)
		}
		w.Write(finalResponseJSON)

		postgresdb.DB.Exec("DELETE FROM staff_sessions WHERE expires_at < now()")
	}
}

// Helper function to extract session ID from request
func extractSessionID(r *http.Request) string {
	var sessionID string

	// Check for session ID in cookies
	cookie, err := r.Cookie("sessionID")
	if err == nil {
		sessionID = cookie.Value

		if r.Header.Get("X-Request-Source") == "client" {
			sessionID, err = url.QueryUnescape(sessionID)
			if err != nil {
				return ""
			}
		}
	} else {
		// Check in headers for SSR requests
		sessionID = r.Header.Get("X-Session-ID")
	}

	return sessionID
}

// Helper function to get user ID and cart ID from session in a single query
func getUserIDFromSession(ctx context.Context, tx *sql.Tx, sessionID string) (string, int, error) {
	var userID string
	var staffPermission int
	err := tx.QueryRowContext(ctx, `
		SELECT s.user_id, u.staff_permission
		FROM staff_sessions s
		JOIN "user" u ON s.user_id = u.user_id
		WHERE s.session_id = $1
	`, sessionID).Scan(&userID, &staffPermission)
	if err != nil {
		log.Printf("Error getting user ID and staff permission from session: %v", err)
		return "", 0, err
	}
	return userID, staffPermission, nil
}

// Helper function to extend session expiry and return the encrypted session ID and expiry datetime
func extendSessionExpiry(ctx context.Context, tx *sql.Tx, sessionID string) (string, time.Time, error) {
	var expiresAt time.Time

	err := tx.QueryRowContext(
		ctx,
		"UPDATE staff_sessions SET expires_at = NOW() + INTERVAL '15 minutes' WHERE session_id = $1 RETURNING expires_at",
		sessionID,
	).Scan(&expiresAt)
	if err != nil {
		log.Printf("Error updating session expiry: %v", err)
		return "", time.Time{}, err
	}

	// Encrypt the session ID
	encryptedSessionID, err := EncryptSessionID(sessionID)
	if err != nil {
		log.Printf("Error encrypting session ID: %v", err)
		return "", time.Time{}, err
	}

	return encryptedSessionID, expiresAt, nil
}

// Helper function to check if the request is from a client
func isClientRequest(r *http.Request) bool {
	return r.Header.Get("X-Request-Source") == "client"
}
