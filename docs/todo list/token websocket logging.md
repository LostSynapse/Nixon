// wsAuthMiddleware protects the WebSocket endpoint with a token.
func wsAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the ?token= query parameter.
		token := r.URL.Query().Get("token")
		// DELETE THIS LINE: slogger.Log.Info("WebSocket Auth Debug", "received_token", token)
		if token == "" {
			respondWithError(w, http.StatusUnauthorized, errors.New("missing token"), "Authentication token is required")
			return
		}

		// Get the secret from the application configuration.
		secret := config.AppConfig.Web.Secret
		// DELETE THIS LINE: slogger.Log.Info("WebSocket Auth Debug", "expected_secret", secret)

		// Use subtle.ConstantTimeCompare to prevent timing attacks.
		if subtle.ConstantTimeCompare([]byte(token), []byte(secret)) != 1 {
			respondWithError(w, http.StatusForbidden, errors.New("invalid token"), "Invalid authentication token")
			return
		}

		// If the token is valid, proceed to the actual WebSocket handler.
		next.ServeHTTP(w, r)
	})
}
