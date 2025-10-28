# Secure WebSocket Token Configuration for Frontend

This guide explains how to remove hardcoded secrets from your frontend application (`web/src/hooks/useNixonApi.js`) and instead use environment variables to manage your WebSocket authentication token. This is a critical security best practice, especially when deploying to target hardware or production environments.

## The Problem: Hardcoded Secrets

In your current setup, the WebSocket authentication token (`"nixon-default-secret"`) is directly embedded within your frontend JavaScript file:

```javascript
const token = "nixon-default-secret";
```

This poses several security risks:

*   **Vulnerability:** Anyone with access to your compiled frontend code can easily discover this secret.
*   **Maintenance Overhead:** If you need to change the secret, you must modify the code, rebuild the frontend, and redeploy.
*   **Environment Specificity:** It doesn't allow for different secrets in development, staging, and production environments.

## The Solution: Environment Variable Injection with Vite

Vite, your frontend build tool, supports injecting environment variables directly into your client-side code at **build time**. By prefixing environment variables with `-VITE_` (e.g., `VITE_WS_SECRET`), Vite makes them available in your code via `import.meta.env`.

This approach allows you to:

*   **Securely Manage Secrets:** Your actual secret is not stored in the source code.
*   **Flexible Deployment:** You can set different `VITE_WS_SECRET` values during your build process for various environments without changing the code.
*   **Development Fallback:** You can maintain a "default" secret for local development when the environment variable isn't explicitly set.

## Step-by-Step Implementation Guide

Follow these instructions to modify `web/src/hooks/useNixonApi.js` to use an environment variable for the WebSocket token.

### **1. Modify `web/src/hooks/useNixonApi.js`**

This step updates your frontend code to read the WebSocket secret from an environment variable.

1.  **Open the file:** `web/src/hooks/useNixonApi.js` in your editor.

2.  **Locate the `token` declaration:**
    Find the line that currently looks like this (around line 31):

    ```javascript
    const token = "nixon-default-secret";
    ```

3.  **Replace that line** with the following, which reads the token from `import.meta.env.VITE_WS_SECRET` and falls back to `"nixon-default-secret"` if the environment variable is not defined:

    ```javascript
    const token = import.meta.env.VITE_WS_SECRET || "nixon-default-secret"; // MODIFIED: Reads token from environment variable or uses fallback
     ```

4.  **Verify the surrounding code:**
    The relevant section of your file should now look like this:

    ```javascript
    Line 30:        const connectWebSocket = useCallback(() => {
    Line 31:          const token = import.meta.env.VITE_WS_SECRET || "nixon-default-secret";
    Line 32:          const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    Line 33:          const socketUrl = `${wsProtocol}//${window.location.host}/ws?token=${token}`; // MODIFIED: Use wsProtocol
    Line 34:          if (socketRef.current && socketRef.current.readyState < 2) return;
    ```

5.  **Save the file:** `web/src/hooks/useNixonApi.js`.

### **2. Prepare for Production Builds on Target Hardware**

For your development environment within Firebase Studio, the fallback `nixon-default-secret` will continue to be used as `VITE_WS_SECRET` is not explicitly set for `npm run dev`.

However, when you perform a **production build (`npm run build`)** for deployment to your target hardware, you **must** provide the `VITE_WS_SECRET` environment variable.

**Example for a production build command:**

When building your frontend for production, ensure `VITE_WS_SECRET` is set in your build environment:

```bash
cd web
VITE_WS_SECRET="your_secure_production_secret_here" npm run build
```

Replace `"your_secure_production_secret_here"` with the actual secret your Go backend uses for WebSocket authentication. This command will inject your production secret into the compiled JavaScript bundle in `web/dist`.