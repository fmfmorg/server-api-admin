package firstemployee

import (
	"encoding/json"
	"net/http"
	"server-api-admin/util/password"
	"server-api-admin/util/postgresdb"

	"github.com/julienschmidt/httprouter"
)

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func createFirstEmployee(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !password.VerifyPassword(req.Password) {
		http.Error(w, "Invalid password", http.StatusBadRequest)
		return
	}

	completeHash, encodedSalt, err := password.GeneratePasswordHash(req.Password)
	if err != nil {
		http.Error(w, "Failed to generate password hash", http.StatusInternalServerError)
		return
	}

	tx, _ := postgresdb.DB.BeginTx(r.Context(), nil)
	defer tx.Rollback()

	var userID string
	err = tx.QueryRowContext(
		r.Context(),
		`INSERT INTO "user" (email, password_hash, salt, staff_permission, is_newsletter_subscribed, email_verified) VALUES ($1, $2, $3, 9223372036854775807, TRUE, TRUE) RETURNING user_id`,
		req.Email,
		completeHash,
		encodedSalt,
	).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.ExecContext(
		r.Context(),
		"INSERT INTO newsletter_subscription (user_id) VALUES ($1)",
		userID,
	)
	if err != nil {
		http.Error(w, "Failed to subscribe newsletter", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
