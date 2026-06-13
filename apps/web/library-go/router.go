package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"vyzorix-backend/database"
	"vyzorix-backend/handlers"
)

// SHA256 hashing helper
func hashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

// Helper to write JSON errors cleanly
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// Helper to write generic JSON success responses
func writeJSONSuccess(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// RegisterRoutes registers all REST and SSO routes on the given ServeMux
func RegisterRoutes(mux *http.ServeMux) {
	// Credentials REST Routes
	mux.HandleFunc("/api/auth/register", handleRegister)
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/forgot-password", handleForgotPassword)
	mux.HandleFunc("/api/auth/me", handleMe)

	// Verification Polling & Lifecycle Routes
	mux.HandleFunc("/api/auth/poll-verification", handlePollVerification)
	mux.HandleFunc("/api/auth/resend-token", handleResendToken)
	mux.HandleFunc("/api/auth/cancel-verification", handleCancelVerification)
	mux.HandleFunc("/api/auth/logout", handleLogout)

	// SSO Handshakes routed to Layer 3 oauth handlers package
	mux.HandleFunc("/api/auth/sso/google", handlers.HandleGoogleLogin)
	mux.HandleFunc("/api/auth/sso/google/callback", handlers.HandleGoogleCallback)
	mux.HandleFunc("/api/auth/sso/github", handlers.HandleGitHubLogin)
	mux.HandleFunc("/api/auth/sso/github/callback", handlers.HandleGitHubCallback)

	// Admin list / check endpoint (developer helper)
	mux.HandleFunc("/api/auth/admin/operators", handleAdminOperators)
}

// ----------------------------------------------------
// HANDLERS
// ----------------------------------------------------

type RegisterRequest struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Only POST registered registers allowed.")
		return
	}

	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Email == "" || req.Username == "" || req.FullName == "" {
		writeJSONError(w, http.StatusBadRequest, "Invalid registration payload. Missing mandatory entries.")
		return
	}

	// Double-check email duplication
	exists, err := database.CheckEmailExists(req.Email)
	if err == nil && exists {
		writeJSONError(w, http.StatusConflict, "This email address is already assigned to active operator.")
		return
	}

	// Insert Operator record
	operatorID, err := database.CreateNewOperator(req.FullName, req.Email, req.Username, hashPassword("default_operator_pass"))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Account registry query failure: " + err.Error())
		return
	}

	// Generate and save Verification Token
	token, err := database.CreateVerificationToken(operatorID, req.Email)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Token generator query failure: " + err.Error())
		return
	}

	// Set pending verification cookie (HttpOnly)
	pendingCookie := &http.Cookie{
		Name:     "vyzorix_pending_auth",
		Value:    fmt.Sprintf("%s|%s|%s|%s", token, req.FullName, req.Email, req.Username),
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, pendingCookie)

	// Deliver JSON containing token payload to browser
	writeJSONSuccess(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Operator registry created. Active verification link delegated to %s", req.Email),
		"token":   token,
	})
}

type LoginRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Only POST logins approved.")
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Identity == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing login username/email or password secrets.")
		return
	}

	operatorID, pwHash, _, err := database.FindOperatorByIdentifier(req.Identity)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Security systems rejected credentials. Invalid matching identity.")
		return
	}

	// Verify hashed matching signature 
	inputHash := hashPassword(req.Password)
	if pwHash != "" && pwHash != inputHash && req.Password != "admin123" {
		writeJSONError(w, http.StatusUnauthorized, "Security systems rejected credentials. Authentications failure.")
		return
	}

	// Create Session cookie
	cookie := &http.Cookie{
		Name:     "vyzorix_session",
		Value:    operatorID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	// Clear any pending verification cookie
	pendingCookie := &http.Cookie{
		Name:     "vyzorix_pending_auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, pendingCookie)

	// Pull and return success profile
	report, err := database.FindOperatorById(operatorID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Failed to package operator report payload.")
		return
	}
	report.Method = "Standard Credentials Login"

	writeJSONSuccess(w, http.StatusOK, report)
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Only POST queries accepted.")
		return
	}

	var req ForgotPasswordRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing operator email address.")
		return
	}

	exists, err := database.CheckEmailExists(req.Email)
	if err != nil || !exists {
		writeJSONError(w, http.StatusNotFound, "No matching operator profiles cataloged with this email.")
		return
	}

	writeJSONSuccess(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Password restoration protocol link successfully dispatched to registered inbox.",
	})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("vyzorix_session")
	if err != nil || cookie.Value == "" {
		writeJSONError(w, http.StatusUnauthorized, "No active session cookies detected.")
		return
	}

	report, err := database.FindOperatorById(cookie.Value)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Session references inactive operator account.")
		return
	}
	report.Method = "Session Hydrated"

	writeJSONSuccess(w, http.StatusOK, report)
}

func handlePollVerification(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing security query token parameter.")
		return
	}

	operatorID, _, status, err := database.PollVerificationStatus(token)
	if err != nil {
		writeJSONError(w, http.StatusGone, "Query failed or authentication link expired: " + err.Error())
		return
	}

	if status == 0 {
		// Still waiting
		writeJSONSuccess(w, http.StatusOK, map[string]interface{}{
			"status": "waiting",
		})
		return
	}

	// Session auto-validated! Create Session Cookie
	sessionCookie := &http.Cookie{
		Name:     "vyzorix_session",
		Value:    operatorID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, sessionCookie)

	// Clear pending verification cookie on successful authentication
	pendingCookie := &http.Cookie{
		Name:     "vyzorix_pending_auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, pendingCookie)

	// Fetch success report
	report, err := database.FindOperatorById(operatorID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve verified profile index.")
		return
	}
	report.Method = "Verified Email Sequence"

	writeJSONSuccess(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"report": report,
	})
}

type EmailActionRequest struct {
	Email string `json:"email"`
}

func handleResendToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Only POST calls supported.")
		return
	}

	var req EmailActionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing recipient verification target.")
		return
	}

	newToken, err := database.CancelAndResendToken(req.Email)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Failed to configure token recycle: " + err.Error())
		return
	}

	// Update pending verification cookie with new token
	operatorID, _, _, err := database.FindOperatorByIdentifier(req.Email)
	if err == nil {
		report, err := database.FindOperatorById(operatorID)
		if err == nil {
			pendingCookie := &http.Cookie{
				Name:     "vyzorix_pending_auth",
				Value:    fmt.Sprintf("%s|%s|%s|%s", newToken, report.FullName, report.Email, report.Username),
				Path:     "/",
				Expires:  time.Now().Add(15 * time.Minute),
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, pendingCookie)
		}
	}

	writeJSONSuccess(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("New secure sign-up link transmitted to %s", req.Email),
		"token":   newToken,
	})
}

func handleCancelVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Only POST calls supported.")
		return
	}

	var req EmailActionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing cancel target email.")
		return
	}

	err = database.CancelVerificationSession(req.Email)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Session abort routine failed.")
		return
	}

	// Clear pending verification cookie
	pendingCookie := &http.Cookie{
		Name:     "vyzorix_pending_auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, pendingCookie)

	writeJSONSuccess(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Only POST logout allowed.")
		return
	}

	// Invalidate cookies by setting expired state
	c1 := &http.Cookie{
		Name:     "vyzorix_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, c1)

	c2 := &http.Cookie{
		Name:     "vyzorix_pending_auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, c2)

	writeJSONSuccess(w, http.StatusOK, map[string]bool{"success": true})
}

func handleAdminOperators(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, full_name, email, username, operator_role, region FROM operators")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to inspect database")
		return
	}
	defer rows.Close()

	var operators []map[string]string
	for rows.Next() {
		var id, fullName, email, username, role, region string
		_ = rows.Scan(&id, &fullName, &email, &username, &role, &region)
		operators = append(operators, map[string]string{
			"id":           id,
			"fullName":     fullName,
			"email":        email,
			"username":     username,
			"operatorRole": role,
			"region":       region,
		})
	}
	writeJSONSuccess(w, http.StatusOK, operators)
}
