# 📝 Consolidated Project Design Document
DO NOT USE THIS DOCUMENT AT THIS TIME, IT IS A WORK IN PROGRESS AND CONTAINS CONFLICTING INFORMATION

> **Gemini's Annotated Commentary**
>
> This document consolidates all provided source materials (Source: project backlog.md, devplan.md, new mandates.md, roadmap.md, reference_appliance_backup.md, State machine Logic - factory reset.md, UX Design reference.md). The `UX Design reference.md` file serves as the foundational structure for the UI/UX sections (Source: UX Design reference.md, Line: 1).
>
> **High-Level Observations:**
>
> * **Primary Conflict:** A direct conflict exists regarding the recording "pause" function. The `new mandates.md` file (Source: new mandates.md, Line: 49) explicitly **disallows** it to protect file integrity (Source: new mandates.md, Line: 49), while the older `roadmap.md` file lists it as a requirement (Source: roadmap.md, Line: 43). This must be resolved. I have detailed this in the commentary under Section 13.1.
>
> * **Critical Dependency:** The **Managed Appliance Upgrade System** (detailed in `reference_appliance_backup.md` (Source: reference_appliance_backup.md, Line: 1) and `devplan.md` (Source: devplan.md, Line: 86)) is the most complex subsystem defined. It involves a privileged executable, the `nixon-migrator` (Source: reference_appliance_backup.md, Line: 28), and has numerous P1 (Critical) security and integrity items in the backlog (Source: project backlog.md, Lines: 22-25). The security audit for this component (`[P1-6]`) (Source: project backlog.md, Line: 23) should be considered a non-negotiable prerequisite to implementation.
>
> * **Architectural Strength:** The "Dumb" UI / "Smart" Broker architecture is a recurring and well-defined theme (Source: UX Design reference.md, Line: 8). This model, combined with the `go-plugin` process isolation (Source: UX Design reference.md, Line: 44), is the core strategy for achieving system stability.
>
> * **Immediate Actions:** The audit findings (Source: roadmap.md, Lines: 13-20) and the "Priority 1: Code Hygiene" tasks (Source: devplan.md, Line: 6) represent immediate, foundational work that must be completed to stabilize the codebase before new features are built on it.

## 1.0 Core Philosophy & Guiding Principles

This document defines the front-end architecture for a **modular, extensible, and resource-efficient** user interface for an embedded appliance (Source: UX Design reference.md, Line: 3). The UI will manage **all features and functions of the application** (Source: UX Design reference.md, Line: 3), including but not limited to Broadcast, Transport, Role-Based Access Control (RBAC), MIDI, and Audio Mixing modules (Source: UX Design reference.md, Line: 4).

The core philosophy is built on **Strict Decoupling** and **Resource Conservation** (Source: UX Design reference.md, Line: 7). The front-end (**React**) must be a "**dumb**" **rendering layer** (Source: UX Design reference.md, Line: 8) that is dynamically driven by a "**smart**" **backend (Go)** (Source: UX Design reference.md, Line: 8). The failure or disabling of any single module, especially third-party modules, must not impact the stability or performance of the core system or other modules (Source: UX Design reference.md, Line: 9).

The UI must also scale in complexity (Source: UX Design reference.md, Line: 11), presenting a simplified interface for standard users (Source: UX Design reference.md, Line: 12) and progressively revealing advanced controls for professionals and developers (Source: UX Design reference.md, Line: 12), all from a single, unified codebase (Source: UX Design reference.md, Line: 12).

## 2.0 Current State & Audit Findings

The existing `Nixon-testing` codebase provides a solid foundation (Source: roadmap.md, Line: 13) but requires immediate attention in several areas to ensure stability, maintainability, and security (Source: roadmap.md, Line: 13). The `audit.md` report highlights the following critical issues (Source: roadmap.md, Line: 14):

* **Lack of Structured Logging:** The current use of the standard `log` package produces inconsistent and difficult-to-parse logs (Source: roadmap.md, Line: 16), hindering effective debugging (Source: roadmap.md, Line: 16).

* **Inconsistent Error Handling:** Error handling is not standardized (Source: roadmap.md, Line: 17), with a mix of `log.Fatal`, `panic`, and unhandled errors (Source: roadmap.md, Line: 17), leading to unpredictable application behavior (Source: roadmap.md, Line: 18).

* **Missing API Authentication:** The WebSocket API at `/ws` is unprotected (Source: roadmap.md, Line: 19), allowing any user on the network to execute arbitrary commands (Source: roadmap.md, Line: 19), posing a significant security risk (Source: roadmap.md, Line: 19). (See related backlog item `[P1-1]` (Source: project backlog.md, Line: 18)).

* **Outdated Dependencies:** Several Go and JavaScript dependencies have known vulnerabilities that must be addressed (Source: roadmap.md, Line: 20).

These findings directly inform the first phase of our roadmap (Source: roadmap.md, Line: 22).

## 3.0 Core Technology Stack & Standards

| Component | Standard | Rationale | 
 | :--- | :--- | :--- | 
| **Framework** | **React** (Latest Stable Version) (Source: UX Design reference.md, Line: 20) | Industry-standard, component-based, and performant (Source: UX Design reference.md, Line: 20). | 
| **Styling** | **Tailwind CSS** (Source: UX Design reference.md, Line: 21) | A utility-first framework that allows for rapid, custom styling and seamless responsive design (mobile/desktop) without pre-defined UI kits (Source: UX Design reference.md, Line: 21). | 
| **Style Guide** | **Airbnb React/JSX Style Guide** (Source: UX Design reference.md, Line: 22) | Provides a comprehensive, industry-standard set of rules for code consistency (Source: UX Design reference.md, Line: 22), enforced via **ESLint** (Source: UX Design reference.md, Line: 22). | 
| **Form Handling** | **React Hook Form** (Source: UX Design reference.md, Line: 23) | A performant library for managing form state (Source: UX Design reference.md, Line: 23). The backend API schema is specifically designed to match this library's requirements to eliminate a data-mapping layer on the frontend (Source: UX Design reference.md, Line: 23). | 

## 4.0 Code Hygiene, Standards, and Dependency Management

This section integrates tasks from `devplan.md` (Source: devplan.md, Line: 6) and `roadmap.md` (Source: roadmap.md, Line: 24) that address the audit findings (Source: roadmap.md, Line: 14).

* **Go Style Guide (Mandate):** All code must adhere strictly to the **Google Go Style Guide** (Source: reference_appliance_backup.md, Line: 11).

* **Go Style Guide (Task):** Run `gofmt -s -w .` (Source: devplan.md, Line: 44) and `go vet ./...` (Source: devplan.md, Line: 44) across the entire project to ensure code is formatted and free of common errors (Source: devplan.md, Line: 44).

* **Go Style Guide (Task):** Manually review code for adherence to Google Go Style Guide conventions (Source: devplan.md, Line: 45), particularly regarding variable naming, interface usage, and package organization (Source: devplan.md, Line: 46).

* **Dependency Management (Go Task):** Run `go mod tidy` to clean up the Go workspace (Source: devplan.md, Line: 8).

* **Dependency Management (Go Task):** Run `go get -u ./...` to update all Go dependencies to their latest minor/patch versions (Source: devplan.md, Line: 9).

* **Dependency Management (Frontend Task):** Run `npm update` in the `web/` directory to update all frontend dependencies (Source: devplan.md, Line: 10).

* **Vulnerability Resolution (Task):** Run `go mod vendor` (Source: devplan.md, Line: 11) and `npm audit fix` (Source: devplan.md, Line: 11) to resolve any outstanding vulnerabilities (Source: devplan.md, Line: 11).

* **Vulnerability Resolution (Roadmap):** Update all outdated Go and NPM packages to their latest secure versions using `go get` and `npm update` (Source: roadmap.md, Line: 32).

* **Vulnerability Resolution (Roadmap):** Perform a vulnerability scan to ensure all known issues are resolved (Source: roadmap.md, Line: 33).

* **License Compliance (Backlog):** (See backlog item `[P1-10]` (Source: project backlog.md, Line: 27)) Defines the legal process for auditing and complying with all third-party software licenses (Source: project backlog.md, Line: 27).

## 5.0 Prescriptive Backend Architecture (Frontend-Impacting)

To achieve the goals of stability and simplicity (Source: UX Design reference.md, Line: 34), the front-end design *requires* the backend to adhere to the following architecture (Source: UX Design reference.md, Line: 35).

### 5.1 Critical Resource Protection Policy (RK3588)

This policy mandates the binding of processes and threads to specific CPU cores on the **Rockchip RK3588** (Source: new mandates.md, Line: 6) to ensure **ultra-low latency** for the Control Plane (Source: new mandates.md, Line: 7).

#### 5.1.1 CPU Core Partitioning Mandates

| Cluster | Cores | Primary Function | Priority | 
 | :--- | :--- | :--- | :--- | 
| **Cortex-A76** (Performance) (Source: new mandates.md, Line: 12) | 4 Cores (Source: new mandates.md, Line: 12) | **Zone A: Real-Time Control & UI.** (Source: new mandates.md, Line: 12) Dedicated to PipeWire communication, CGo interfaces, and critical UI responsiveness (Source: new mandates.md, Line: 12). | **Highest** (`SCHED_FIFO`) (Source: new mandates.md, Line: 12) | 
| **Cortex-A55** (Efficiency) (Source: new mandates.md, Line: 13) | 4 Cores (Source: new mandates.md, Line: 13) | **Zone B: Background Work.** (Source: new mandates.md, Line: 13) Dedicated to noisy tasks (File I/O, Logging, GC, File Pruning) (Source: new mandates.md, Line: 13). | **Lowest** (`SCHED_OTHER`) (Source: new mandates.md, Line: 13) | 

#### 5.1.2 Thread Binding (Go/CGo Isolation)

* **Real-Time Threads (A76):** Goroutines handling PipeWire/WirePlumber CGo calls **must** use **`runtime.LockOSThread()`** (Source: new mandates.md, Line: 17). The `nixon-migrator` will bind these TIDs to a dedicated **Cortex-A76** core (Source: new mandates.md, Line: 17) and set **Real-Time priority** (Source: new mandates.md, Line: 18).

* **Low-Priority Threads (A55):** Goroutines handling encoding, file management, and logging **must** use **`runtime.LockOSThread()`** (Source: new mandates.md, Line: 19). The `nixon-migrator` will bind these TIDs to the **Cortex-A55** cores (Source: new mandates.md, Line: 19) and set a **low OS priority** (Source: new mandates.md, Line: 20).

### 5.2 Plugin Engine: HashiCorp `go-plugin`

All extensible modules, both internal and third-party, **must** be managed by the `go-plugin` library (Source: UX Design reference.md, Line: 41).

* **Benefit (Process Isolation):** This runs each plugin in a **separate OS process** (Source: UX Design reference.md, Line: 44). A crash or memory leak in one plugin will terminate only its own process (Source: UX Design reference.md, Line: 44), ensuring the core broker and all other modules remain stable (Source: UX Design reference.md, Line: 45). This is a critical requirement (Source: UX Design reference.md, Line: 45).

* **Latency Note:** This abstraction applies to the **control plane** (settings, commands, status updates) only (Source: UX Design reference.md, Line: 46). High-bandwidth, low-latency media streams (e.g., PipeWire) will follow their own data plane (Source: UX Design reference.md, Line: 47), as defined by their respective modules (Source: UX Design reference.md, Line: 47).

* **Architecture Task:** Define the primary `Plugin` interface in `internal/plugin/plugin.go` (Source: devplan.md, Line: 56).

* **Architecture Task:** Implement a `PluginManager` in the same package to handle loading, initializing, and managing the lifecycle of plugins (Source: devplan.md, Line: 57).

* **Architecture Task:** Design the communication channel (e.g., Go channels) between the core application and the plugins (Source: devplan.md, Line: 58).

* **Refactor Task:** Refactor configuration loading in `internal/config/config.go` to handle a generic map of plugin settings (Source: devplan.md, Line: 59), removing hardcoded structs (Source: devplan.md, Line: 59).

* **Refactor Task (Core Components):** Create a new directory `internal/plugins/icecast` (Source: devplan.md, Line: 63). Move existing Icecast logic into this package (Source: devplan.md, Line: 63) and adapt it to the `Plugin` interface (Source: devplan.md, Line: 63).

* **Refactor Task (Core Components):** Create a new directory `internal/plugins/srt` (Source: devplan.md, Line: 64). Move existing SRT logic into this package (Source: devplan.md, Line: 64) and adapt it to the `Plugin` interface (Source: devplan.md, Line: 64).

* **Refactor Task (Core Components):** Create a new directory `internal/plugins/jackwire2` (Source: devplan.md, Line: 65). Move all JACK/PipeWire routing and control logic from `internal/pipewire` into this new package (Source: devplan.md, Line: 65) and adapt it to the `Plugin` interface (Source: devplan.md, Line: 65).

* **Refactor Task (Core Components):** Update API handlers and frontend calls (`useNixonApi.js`) to interact with all new plugins via the `PluginManager`'s generic methods (Source: devplan.md, Line: 66).

* **Roadmap Item:** Design and implement a Go-based plugin architecture (Source: roadmap.md, Line: 47).

* **Roadmap Item:** Refactor existing Icecast and SRT streaming functionalities into plugins (Source: roadmap.md, Line: 48), validating the plugin API (Source: roadmap.md, Line: 48).

### 5.3 The Central Broker Model

All communication between the React frontend and *any* module **must** be proxied through a single **Central Broker** component in the core Go application (Source: UX Design reference.md, Line: 51).

* **Benefit (Total Abstraction):** The React frontend is completely unaware of whether a module is internal (a direct Go function call) or external (an RPC call via go-plugin) (Source: UX Design reference.md, Line: 54). This creates a single, simple, and secure API surface (Source: UX Design reference.md, Line: 55).

* **Backlog Item:** (See backlog item `[P1-4]` (Source: project backlog.md, Line: 21)) An API Rate Limiting Policy is mandated to prevent a rogue plugin or user from flooding the Broker via WebSocket (Source: project backlog.md, Line: 21).

### 5.4 API Communication Protocols

* **Initial State Load:** A single **HTTP/S GET** request upon application load to retrieve the complete initial state (Section 5.5) (Source: UX Design reference.md, Line: 59).

* **Control Plane:** A single, persistent **WebSocket** connection (Source: UX Design reference.md, Line: 60). All real-time data, actions, settings commands, and status updates **must** be handled via a Pub/Sub message model over this connection (Source: UX Design reference.md, Line: 60).

* **Data Visualization Plane (Optional):** A second, optional **WebSocket** connection may be initiated by specific widgets (Source: UX Design reference.md, Line: 61). This socket is exclusively for streaming high-frequency, non-critical visualization data (e.g., RMS, FFT data in binary format) from the backend (Source: UX Design reference.md, Line: 62). This data **must not** be sent over the main Control Plane socket (Source: UX Design reference.md, Line: 62).

* **Authentication:** All API communication (HTTP GET and all WebSockets) will be secured using the existing token-based authentication system (Source: UX Design reference.md, Line: 63).

* **API Framework Refinement:** In `internal/common/structs.go`, define standardized request/response structs for WebSocket commands (Source: devplan.md, Line: 52).

* **API Framework Refinement:** Refactor the WebSocket message handling in `internal/websocket/websocket.go` into a more robust command dispatcher pattern to simplify adding new commands (Source: devplan.md, Line: 53).

* **WebSocket Security Task:** Add a new HTTP endpoint (e.g., `/login`) for a simple, token-based authentication mechanism (Source: devplan.md, Line: 38).

* **WebSocket Security Task:** Modify the WebSocket upgrade handler in `internal/websocket/websocket.go` to require a valid authentication token passed as a query parameter (Source: devplan.md, Line: 39).

* **WebSocket Security Task:** Update the frontend API hook (`web/src/hooks/useNixonApi.js`) to first authenticate with the `/login` endpoint (Source: devplan.md, Line: 40).

* **WebSocket Security Task:** Store the received token in memory and append it to the WebSocket connection URL (Source: devplan.md, Line: 41).

* **Roadmap Item (WebSocket Security):** Implement a token-based authentication mechanism for the `/ws` endpoint (Source: roadmap.md, Line: 30).

* **Roadmap Item (WebSocket Security):** The UI will need to be updated to handle the authentication flow (Source: roadmap.md, Line: 31). (See related backlog item `[P1-1]` (Source: project backlog.md, Line: 18)).

#### 5.4.1 Data Concurrency & Stale State Protection

To prevent data-loss from concurrent user sessions (e.g., two users editing the same form) (Source: UX Design reference.md, Line: 66), all state-changing **WebSocket messages** (e.g., saving settings) **must** use a concurrency token (Source: UX Design reference.md, Line: 66).

1. **Initial State:** The initial HTTP GET request **must** provide a concurrency token (e.g., `"_version": 123`) with any stateful data (like a settings form) (Source: UX Design reference.md, Line: 69).

2. **Save Action:** The "save" WebSocket message (e.g., `pluginID.settings.save`) **must** include this `_version` token in its JSON payload (Source: UX Design reference.md, Line: 70).

3. **Broker Validation:** The **Central Broker** **must** check this token (Source: UX Design reference.md, Line: 71). If the backend's current version does not match the token from the user (Source: UX Design reference.md, Line: 71), the Broker **must reject** the save by sending a **Standardized Notification** (Section 6.2, `level: "error", code: "STALE_DATA"`) back over the WebSocket (Source: UX Design reference.md, Line: 72).

### 5.5 Data Transformation and Initial State

The Go backend (Google Style Guide) will produce JSON with `PascalCase` keys (Source: UX Design reference.md, Line: 76).

* **Key Transformation:** The **Go backend (Central Broker) is responsible** for transforming the JSON keys from `PascalCase` to `camelCase` before transmission to the React frontend (Source: UX Design reference.md, Line: 79). This maintains the React frontend as a "dumb rendering layer" (Source: UX Design reference.md, Line: 79).

* **Initial State Retrieval:** Upon frontend connection, the UI **must** first use a **one-time HTTP/S GET request** to retrieve all static and current-state data (settings, layout, plugin status, etc.) (Source: UX Design reference.md, Line: 80).

* **State Re-Synchronization:** This HTTP/S GET request **must** also be re-triggered automatically by the frontend upon re-establishing a lost Control Plane WebSocket connection to prevent a stale UI state (Source: UX Design reference.md, Line: 81).

### 5.6 Configuration Centralization

* **Task 1.2:** Identify any configuration values hardcoded in the application source (Source: devplan.md, Line: 14).

* **Task 1.2:** Add corresponding fields to the `Config` structs in `internal/config/config.go` (Source: devplan.md, Line: 15).

* **Task 1.2:** In the `config.Load()` function, implement a `viper.SetDefault("struct.field", "defaultValue")` call for each configuration parameter (Source: devplan.md, Line: 16). This ensures the application can generate a default configuration if `config.json` is not present (Source: devplan.md, Line: 16).

* **Task 1.2:** Replace all identified hardcoded values in the Go source with references to the loaded configuration (e.g., `cfg.Server.ListenAddress`) (Source: devplan.md, Line: 17).

* **Roadmap Item:** Centralize all configuration into the `config.json` file (Source: roadmap.md, Line: 35).

* **Roadmap Item:** Remove hardcoded values from the source code (Source: roadmap.md, Line: 36), using `viper.SetDefault()` in `internal/config/config.go` to provide initial values (Source: roadmap.md, Line: 36).

* **Backlog Item:** (See backlog item `[P1-11]` (Source: project backlog.md, Line: 28)) Defines architecture for managing all configurable file paths (Recording, Logs, Config) via a single Path Resolver Service (PRS) (Source: project backlog.md, Line: 28). (User Reminder: This includes recording paths (Source: 2025-10-30)).

* **Backlog Item:** (See backlog item `[P1-9]` (Source: project backlog.md, Line: 26)) A `localStorage` schema versioning mandate is required to prevent frontend crashes on updates (Source: project backlog.md, Line: 26).

## 6.0 API Versioning, Notifications, & Internationalization

To ensure long-term stability and maintainability, the following contracts are non-negotiable (Source: UX Design reference.md, Line: 87).

### 6.1 Versioning & Handshake

* **Core Handshake:** The initial HTTP/S GET request (Section 5.5) **must** be used for an API version handshake (Source: UX Design reference.md, Line: 90). If the frontend and backend API versions are incompatible (Source: UX Design reference.md, Line: 90), the UI **must** display a "Software Mismatch" error and halt operation (Source: UX Design reference.md, Line: 90).

* **Schema Versioning:** All JSON payloads for the **Dynamic Form Schema Contract** (Section 10.2.1) and **Widget Data** (Section 10.5.2) **must** include a `schemaVersion` key (e.g., `"schemaVersion": "1.2"`) (Source: UX Design reference.md, Line: 91).

* **Broker Validation:** The **Central Broker** is responsible for validating schema versions (Source: UX Design reference.md, Line: 92). If a plugin provides a schema version the Broker knows the frontend cannot support (Source: UX Design reference.md, Line: 92), the Broker **must not** send that schema (Source: UX Design reference.md, Line: 93). Instead, it must send a **Standardized Notification** (Section 6.2) to the frontend (Source: UX Design reference.md, Line: 93).

### 6.2 Standardized Notification Schema

All informational, success, and error messages sent from the Broker to the frontend **must** use the following standardized JSON schema (Source: UX Design reference.md, Line: 96). This allows the frontend to use a single, "dumb" notification/toast component (Source: UX Design reference.md, Line: 96).

```
{
  "level": "success | info | warning | error",
  "source": "pluginID or broker",
  "code": "VALIDATION_FAILED or STALE_DATA or SAVE_SUCCESS",
  "messageKey": "notify.save.success",
  "messageContext": { "item": "Settings" }
}

```
(Source: UX Design reference.md, Lines: 98-105)

* **Broker Mandate:** The Broker **must** send a `level: "success"` notification upon the successful completion of actions like saving settings (Source: UX Design reference.md, Line: 107).

### 6.3 Internationalization (i18n)

To support localization, all user-facing strings **must** be replaced with **translation keys** across all API contracts (Source: UX Design reference.md, Line: 110).

* **Contract Mandate:** Static strings like `label` or `pluginName` are forbidden (Source: UX Design reference.md, Line: 112). Contracts must use keys (e.g., `labelKey: "plugin.icecast.form.server_address"`) (Source: UX Design reference.md, Line: 112).

* **Broker Responsibility:** The **Central Broker** is responsible for collecting translation files (e.g., `en.json`, `es.json`) from all installed plugins and aggregating them (Source: UX Design reference.md, Line: 113).

* **Frontend Responsibility:** The frontend is responsible for loading the appropriate, aggregated language file from the Broker (via the initial HTTP GET) (Source: UX Design reference.md, Line: 114) and mapping all keys to the correct translated text (Source: UX Design reference.md, Line: 114).

* **Backlog Item:** (See backlog item `[P3-8]` (Source: project backlog.md, Line: 70)) The UI to select and manage the display language (Source: project backlog.md, Line: 70).

### 6.4 Structured Logging & Error Handling

* **Structured Logging Task:** Initialize a global `slog.Logger` with a JSON handler in `cmd/nixon/main.go` (Source: devplan.md, Line: 20).

* **Structured Logging Task:** Pass the logger instance as a dependency to all top-level application components (e.g., API router, control manager) (Source: devplan.md, Line: 21).

* **Structured Logging Task:** Go through each file in the `internal/` directory and replace all `log.*` calls with corresponding `logger.Info`, `logger.Warn`, `logger.Error`, or `logger.Debug` calls (Source: devplan.md, Line: 22), adding structured key-value pairs where appropriate (Source: devplan.md, Line: 23).

* **Roadmap Item (Logging):** Replace all `log` package calls with a structured logger (`slog`) (Source: roadmap.md, Line: 27).

* **Roadmap Item (Logging):** Establish consistent, machine-readable log formats (JSON) (Source: roadmap.md, Line: 28).

* **Error Handling Task:** Review the entire Go codebase and remove all uses of `log.Fatal` and `panic` for recoverable errors (Source: devplan.md, Line: 26).

* **Error Handling Task:** Refactor functions to return errors instead of calling fatal logging functions (Source: devplan.md, Line: 27).

* **Error Handling Task:** Establish a consistent pattern of bubbling errors up to the API handlers (Source: devplan.md, Line: 28).

* **Error Handling Task:** Ensure API handlers return appropriate HTTP status codes and error messages based on the errors received from underlying components (Source: devplan.md, Line: 29).

* **Roadmap Item (Error Handling):** Refactor the codebase to use a consistent error handling strategy (Source: roadmap.md, Line: 29), returning errors to the caller for appropriate action (Source: roadmap.md, Line: 29).

* **Roadmap Item (Error Handling):** Eliminate all uses of `log.Fatal` and `panic` for recoverable errors (Source: roadmap.md, Line: 30).

## 7.0 State & Data Flow (WebSocket Pub/Sub)

The core of the UI's real-time functionality is a **Pub/Sub** model over the single **Control Plane** WebSocket connection (Source: UX Design reference.md, Line: 119).

### 7.1 One-to-Many (Command Broadcast): "Start All"

* **Action:** The React UI sends *one* message to the Broker (e.g., `system.command.startAll`) (Source: UX Design reference.md, Line: 122).

* **Broker:** The Broker broadcasts this event to *all* running modules (Source: UX Design reference.md, Line: 123).

* **Modules:** Modules that have subscribed to this event (e.g., "Broadcast" modules) will execute their "start" logic (Source: UX Design reference.md, Line: 124). Other modules will ignore it (Source: UX Design reference.md, Line: 124).

### 7.2 Many-to-One (Status Aggregation): "LIVE" Indicator

This is a critical, backend-driven optimization to keep the frontend simple (Source: UX Design reference.md, Line: 127).

* **Modules (The "Many"):** All "Broadcast" modules (Icecast, SRT, etc.) publish their individual status (e.g., `"active"`, `"inactive"`) to a shared aggregate topic (e.g., `broadcast.status.aggregate`) (Source: UX Design reference.md, Line: 130).

* **Broker (The "Aggregator"):** The Broker is the *only* subscriber to this topic (Source: UX Design reference.md, Line: 131). It maintains an internal state (a set) of all active broadcast modules (Source: UX Design reference.md, Line: 131).

* **Frontend (The "One"):** The Broker publishes the *final, aggregated* boolean state (`true` if the set is not empty, `false` if it is) to a *single* topic (e.g., `system.live.status`) (Source: UX Design reference.md, Line: 132). The React "LIVE" indicator component subscribes *only* to this one topic (Source: UX Design reference.md, Line: 132).

## 8.0 Component & Module Architecture

### 8.1 Terminology

* **Module:** The high-level functional unit (e.g., the Icecast plugin) (Source: UX Design reference.md, Line: 138).

* **Component:** The fundamental React building block (Source: UX Design reference.md, Line: 139).

* **Widget Component:** A specific Component on the dashboard that represents a Module (Source: UX Design reference.md, Line: 140).

### 8.2 Module Definition (Manifest)

For the frontend to render a module, the module **must** provide the following metadata to the Broker (Source: UX Design reference.md, Line: 143):

* **`pluginID` (string):** The unique, immutable ID (e.g., `com.icecast.streamer`) (Source: UX Design reference.md, Line: 146). This **must** be used as the prefix for all its Pub/Sub topics (Source: UX Design reference.md, Line: 146).

* **`pluginNameKey` (string):** The translation key for the human-readable display name (e.g., `plugin.icecast.name`) (Source: UX Design reference.md, Line: 147).

* **`pluginType` (string):** The category for UI grouping and logic (Source: UX Design reference.md, Line: 148).

  * `Broadcast`: (e.g., Icecast, SRT) (Source: UX Design reference.md, Line: 149). Participates in "LIVE" aggregation (Source: UX Design reference.md, Line: 149).

  * `Transport`: (e.g., JACK, AES67, PipeWire) (Source: UX Design reference.md, Line: 150). Does not participate in "LIVE" aggregation (Source: UX Design reference.md, Line: 150).

  * `Utility`: (e.g., RBAC, System Monitor) (Source: UX Design reference.md, Line: 151).

* **`minLevel` (string):** The minimum **Feature Level** (`Standard`, `Advanced`, `Professional`, `Developer`) required for this module to be visible (Source: UX Design reference.md, Line: 152).

* **`requiresInitialConfig` (boolean, optional):** If `true` (Source: UX Design reference.md, Line: 153), this flag triggers the "First Run" UI behavior (Section 8.4) when the module is first enabled (Source: UX Design reference.md, Line: 153).

### 8.3 Module State & Resource Conservation

The Broker **must** maintain a persistent **Plugin Status Table** containing `pluginID` and `isEnabled` (boolean) (Source: UX Design reference.md, Line: 156).

* **Resource Conservation:** If `isEnabled` is `false` (Source: UX Design reference.md, Line: 158), the Broker **must not** launch the `go-plugin` process for that module (Source: UX Design reference.md, Line: 158). This ensures disabled modules consume **zero CPU or RAM** (Source: UX Design reference.md, Line: 158).

* **UI Behavior:** The frontend will query this table to populate the "Enable/Disable Modules" settings page (Source: UX Design reference.md, Line: 159).

### 8.4 Module "First Run" UI Behavior

When a user enables a module with the `requiresInitialConfig: true` flag (Source: UX Design reference.md, Line: 162), the Broker **must** instruct the frontend to act based on the user's active **Feature Level** (Source: UX Design reference.md, Line: 162):

* **Standard:** The frontend **must forcibly redirect** the user to the module's settings page (Section 10.2) (Source: UX Design reference.md, Line: 165).

* **Advanced:** The frontend **must open a modal dialog** confirming the plugin is enabled (Source: UX Design reference.md, Line: 166) and prompting the user to "Configure Now" (redirects) or "Configure Later" (dismisses) (Source: UX Design reference.md, Line: 166).

* **Professional & Developer:** No action is taken (Source: UX Design reference.md, Line: 167). The system trusts the user to complete the configuration manually (Source: UX Design reference.md, Line: 167).

### 8.5 Plugin Lifecycle Management

To support a modular ecosystem, the Broker and UI must support the full plugin lifecycle (Source: UX Design reference.md, Line: 170).

* **Broker API:** The **Central Broker** must provide data (via the initial HTTP GET) that lists all plugins (installed and available) (Source: UX Design reference.md, Line: 173) and their `version`, `latest_available_version`, and `compatibility_status` (Source: UX Design reference.md, Line: 173).

* **Frontend UI:** The frontend must provide a "Module Management" settings page that consumes this data (Source: UX Design reference.md, Line: 174). This UI will allow users to install, update, and uninstall modules (via WebSocket commands) (Source: UX Design reference.md, Line: 174). It **must** clearly display compatibility warnings or available updates to the user (Source: UX Design reference.md, Line: 174).

## 9.0 UI/UX Principles & Frontend Mandates

### 9.1 Hierarchical Feature Levels

The entire UI operates on four, system-wide, hierarchical levels (Source: UX Design reference.md, Line: 179):

1.  **Standard** (Source: UX Design reference.md, Line: 181)
2.  **Advanced** (includes Standard) (Source: UX Design reference.md, Line: 182)
3.  **Professional** (includes Advanced) (Source: UX Design reference.md, Line: 183)
4.  **Developer** (includes Professional) (Source: UX Design reference.md, Line: 184)

### 9.2 Global State & Dynamic Configuration

* **Frontend:** The active **Feature Level** is stored in a **Global React Context** (`useFeatureLevel()`) (Source: UX Design reference.md, Line: 188).
* **Backend:** The Go Backend is responsible for filtering all configuration JSON based on the active level (Source: UX Design reference.md, Line: 189).
* **Data Flow:** The frontend receives a simple, pre-filtered JSON payload (via the initial HTTP GET) (Source: UX Design reference.md, Line: 190) containing only the settings and widgets relevant to the user's active level (Source: UX Design reference.md, Line: 190).
* **Visibility:** If a user is at the "Standard" level (Source: UX Design reference.md, Line: 191), the UI does not load or render "Professional" level modules or settings (Source: UX Design reference.md, Line: 191). They are not grayed out; they are **not present** (Source: UX Design reference.md, Line: 192).
* **Backlog Item:** (See backlog item `[P3-10]` (Source: project backlog.md, Line: 72)) Server-Side User Preferences for persistently saving user settings (theme, default feature level) on the backend (Source: project backlog.md, Line: 72).

### 9.3 Accessibility & Navigability Mandates

To ensure the UI is robust, professional, and operable by power users and assistive technologies (like screen readers) (Source: UX Design reference.md, Line: 195), the frontend implementation **must** adhere to the following (Source: UX Design reference.md, Line: 196):

* **1. Full Keyboard Navigability:** All interactive elements in the application (form controls, widgets, buttons, tabs) **must** be fully focusable and operable using only a keyboard (Tab, Shift+Tab, Enter, Space, Arrow Keys) (Source: UX Design reference.md, Line: 198). This is a non-negotiable requirement (Source: UX Design reference.md, Line: 198).
* **2. Semantic HTML & ARIA:** The "dumb" frontend renderer **must** be responsible for (Source: UX Design reference.md, Line: 199):
    * **Generating Correct Associations:** All form inputs must have a unique `id` attribute (Source: UX Design reference.md, Line: 201), and their corresponding `<label>` must use the `htmlFor` attribute (Source: UX Design reference.md, Line: 201).
    * **Using Correct ARIA Roles:** All custom, non-native controls (like a `"toggle"`) **must** be built with the proper ARIA attributes (e.g., `role="switch"`, `aria-checked="true"`) to describe their state and function (Source: UX Design reference.md, Line: 202).

### 9.4 Client-Side State Persistence

* To provide a seamless user experience, non-critical, transient UI state **must** be persisted in the browser's **`localStorage`** (Source: UX Design reference.md, Line: 206).
* **Scope:** This includes, but is not limited to, the user's selected **Feature Level** (9.1) (Source: UX Design reference.md, Line: 209), the "open/closed" state of sidebars (Source: UX Design reference.md, Line: 209), and the last-active tab on any given settings page (Source: UX Design reference.md, Line: 209).
* **Benefit:** This ensures the UI "remembers" the user's context on a page refresh without requiring any backend state management (Source: UX Design reference.md, Line: 210).
* **Backlog Item:** (See backlog item `[P1-9]` (Source: project backlog.md, Line: 26)) A `localStorage` schema versioning mandate is required to prevent frontend crashes on updates (Source: project backlog.md, Line: 26).
* **Backlog Item:** (See backlog item `[P3-7]` (Source: project backlog.md, Line: 69)) Client-Side Validation policy to provide immediate error feedback (Source: project backlog.md, Line: 69).
* **Backlog Item:** (See backlog item `[P3-6]` (Source: project backlog.md, Line: 68)) Optimistic UI Updates policy to make the UI feel instantaneous (Source: project backlog.md, Line: 68).

### 9.5 Contextual Help System

To onboard users and explain complex settings without cluttering the UI (Source: UX Design reference.md, Line: 214), a contextual help system is required (Source: UX Design reference.md, Line: 214).

* **Schema Enhancement:** The **Dynamic Form Schema Contract** (Section 10.2.1) **must** support an optional `helpKey` (an i18n key) (Source: UX Design reference.md, Line: 217).
* **Frontend Implementation:** The "dumb" form renderer **must** be responsible for adding a small `(?)` help icon next to any label that provides a `helpKey` (Source: UX Design reference.md, Line: 218).
* **UI Behavior:** Clicking this icon will trigger a popover or slide-out panel that displays the translated help text for that specific field (Source: UX Design reference.md, Line: 219).

### 9.6 Frontend Logic Centralization

To enforce the "dumb component" philosophy and ensure maintainability (Source: UX Design reference.md, Line: 222), all cross-cutting frontend logic **must** be centralized (Source: UX Design reference.md, Line: 222).

* **Mandate:** All complex, non-visual logic (e.g., WebSocket connection management, i18n mapping, `localStorage` persistence, global state) **must** be abstracted into a **"Core UI Service Layer"** (e.g., global React Contexts and custom Hooks) (Source: UX Design reference.md, Line: 225).
* **Benefit:** Components will remain "dumb" and declarative (Source: UX Design reference.md, Line: 226). (e.g., a component will call `const { isOnline } = useConnection()` instead of managing the WebSocket lifecycle itself) (Source: UX Design reference.md, Line: 226).
* **Backlog Item:** (See backlog item `[P1-2]` (Source: project backlog.md, Line: 19)) Defines how React assets are bundled (e.g., embedded) to ensure compliance with the Upgrade Document (Source: project backlog.md, Line: 19).

---

## 10.0 Global UI Components & Core Features

### 10.1 Global Connection Status

The frontend **must** provide clear, persistent feedback about the state of the **Control Plane** WebSocket connection (Source: UX Design reference.md, Line: 232).

* **Persistent Header Indicator:** The main UI header (top right) **must** display a status indicator (Source: UX Design reference.md, Line: 235):
    * **Green Indicator + "Online" Text:** When the WebSocket connection is active (Source: UX Design reference.md, Line: 236).
    * **Red Indicator + "Offline" Text:** When the WebSocket connection is lost (Source: UX Design reference.md, Line: 237).
* **Modal Overlay:** If the connection is lost (Source: UX Design reference.md, Line: 238), the frontend **must** immediately display a **modal overlay** (e.g., "Connection Lost. Reconnecting...") (Source: UX Design reference.md, Line: 238) that **disables all forms and controls** (Source: UX Design reference.md, Line: 238). This prevents user actions that are guaranteed to fail (Source: UX Design reference.md, Line: 239). The overlay is dismissed once the connection and state re-sync (Section 5.5) are successful (Source: UX Design reference.md, Line: 239).

### 10.2 Settings Pages (Schema-Driven Forms)

* **Dynamic Rendering:** All module settings pages are dynamically rendered from a JSON schema provided by the module (Source: UX Design reference.md, Line: 243) (and pre-filtered by the Broker (Source: UX Design reference.md, Line: 243), received in the initial HTTP GET (Source: UX Design reference.md, Line: 243)).
* **API Optimization:** The JSON schema format **must** be defined to match the API of the chosen React form library (**React Hook Form**) (Source: UX Design reference.md, Line: 244). This **eliminates the need for a data-mapping layer** on the frontend (Source: UX Design reference.md, Line: 245).
* **Saving Data:** Saving settings **must** be performed by sending a targeted Pub/Sub message (e.g., `pluginID.settings.save`) over the **Control Plane** WebSocket (Source: UX Design reference.md, Line: 246), containing the form data and concurrency token (Section 5.4.1) (Source: UX Design reference.md, Line: 247).
* **Save Contract:** All settings save messages sent by the frontend **must** be treated as a **partial update (a JSON merge)** by the Central Broker (Source: UX Design reference.md, Line: 248). The Broker is responsible for merging the submitted fields into the existing configuration (Source: UX Design reference.md, Line: 248), not overwriting the entire object (Source: UX Design reference.md, Line: 249). This protects higher-level settings from being destroyed by a lower-level user (Source: UX Design reference.md, Line: 249).
* **Performance (Code Splitting):** The JavaScript and CSS for each module's settings page **must** be code-split (e.g., using `React.lazy()`) (Source: UX Design reference.md, Line: 250) and loaded on-demand when the user navigates to that page (Source: UX Design reference.md, Line: 251). This keeps the initial application bundle small and fast (Source: UX Design reference.md, Line: 251).
* **Backlog Item:** (See backlog item `[P3-4]` (Source: project backlog.md, Line: 66)) UI to show which users are currently editing a page (Source: project backlog.md, Line: 66).

#### 10.2.1 Dynamic Form Schema Contract

The settings page JSON will be a list of field definitions (an array of objects) (Source: UX Design reference.md, Line: 255), each representing a single input control (Source: UX Design reference.md, Line: 255). This is the prescriptive vocabulary the backend must use (Source: UX Design reference.md, Line: 255).

| Control Type (`type`) | Description | Example Configurable Element (Non-Default) |
| :--- | :--- | :--- |
| **`text`** | Single-line text input (Source: UX Design reference.md, Line: 259). | `password` (Hides input, adds toggle visibility.) (Source: UX Design reference.md, Line: 259) |
| **`number`** | Numeric input (Source: UX Design reference.md, Line: 260). | `min` / `max` (Enforces numerical boundaries.) (Source: UX Design reference.md, Line: 260) |
| **`toggle`** | Simple `true`/`false` switch (Source: UX Design reference.md, Line: 261). | `descriptionKey` (A translation key for a helper string.) (Source: UX Design reference.md, Line: 261) |
| **`select`** | Dropdown for predefined options (Source: UX Design reference.md, Line: 262). | `isMulti` (Allows selecting multiple options.) (Source: UX Design reference.md, Line: 262) |
| **`radio`** | Select one option from a list (Source: UX Design reference.md, Line: 263). | `sortOrder` (Instructions for the frontend to sort the options.) (Source: UX Design reference.md, Line: 263) |
| **`textarea`** | Multi-line text input (Source: UX Design reference.md, Line: 264). | `monospace` (Renders content using a fixed-width font.) (Source: UX Design reference.md, Line: 264) |
| **`file`** | File upload input (Source: UX Design reference.md, Line: 265). | `accepts` (A string of file types, e.g., `"audio/*, .mp3"`.) (Source: UX Design reference.md, Line: 265) |
| **`date` / `time`** | Date and/or time input (Source: UX Design reference.md, Line: 266). | `maxDate` / `minDate` (Sets the boundary for user selection.) (Source: UX Design reference.md, Line: 266) |
| **`color`** | Color picker input (Source: UX Design reference.md, Line: 267). | `defaultColor` (A HEX string to set the initial color value.) (Source: UX Design reference.md, Line: 267) |
| **`range`** | Slider input (Source: UX Design reference.md, Line: 268). | `step` (The increment value for the range slider.) (Source: UX Design reference.md, Line: 268) |
| **`array-text`** | Allows the user to dynamically add/remove a list of text inputs (Source: UX Design reference.md, Line: 269). | `maxItems` (Maximum number of inputs the user can add.) (Source: UX Design reference.md, Line: 269) |
| **`key-value`** | Allows the user to define paired key/value inputs (Source: UX Design reference.md, Line: 270). | `keyType` (Sets validation for the key field, e.g., `"regex"`.) (Source: UX Design reference.md, Line: 270) |
| **`custom-component`**| Reference to a specific, pre-built React component to render (Source: UX Design reference.md, Line: 271). | **Security Constraint:** The list of valid `componentName` references must be **pre-vetted and immutable** by the core frontend application to prevent injection attacks (Source: UX Design reference.md, Line: 271). |
| **`nested-group`** | Renders a subsection of controls within the current form (Source: UX Design reference.md, Line: 272). | `fields` (An array of field definitions to render.) (Source: UX Design reference.md, Line: 272) |
| **`divider`** | A visual separator (Source: UX Design reference.md, Line: 273). | `labelKey` (A translation key for a string to display in the divider.) (Source: UX Design reference.md, Line: 273) |
| **`info-box`** | A box for displaying static contextual information (Source: UX Design reference.md, Line: 274). | `style` (e.g., `"warning"`, `"success"`, `"error"` for color/icon styling.) (Source: UX Design reference.md, Line: 274) |
| **`helpKey` (Optional)**| (On any control) A translation key for contextual help (Source: UX Design reference.md, Line: 275). | `helpKey: "plugin.icecast.help.mountpoint"` (Renders a `(?)` icon) (Source: UX Design reference.md, Line: 275) |

*Note: All user-facing strings (e.g., `label`, `placeholder`, `description`) must be replaced with their i18n equivalents (e.g., `labelKey`, `placeholderKey`, `descriptionKey`).* (Source: UX Design reference.md, Line: 277)

* **Backlog Item:** (See backlog item `[P2-12]` (Source: project backlog.md, Line: 50)) Schema Enhancement to add read-only-text and progress-bar types for dynamic status pages (Source: project backlog.md, Line: 50).

#### 10.2.2 Mandated Settings Pages (from Backlog)

* **Recording Management:** (See backlog item `[P2-1]` (Source: project backlog.md, Line: 39)) The UI for listing, downloading, starting/stopping, and deleting recordings (Source: project backlog.md, Line: 39).
* **User Management (RBAC):** (See backlog item `[P2-2]` (Source: project backlog.md, Line: 40)) The administrative UI for adding/deleting users, managing roles, and changing Feature Levels post-setup (Source: project backlog.md, Line: 40).
* **System Log Viewer:** (See backlog item `[P2-3]` (Source: project backlog.md, Line: 41)) The UI for tailing and viewing structured logs (Source: project backlog.md, Line: 41); essential for debugging the headless appliance (Source: project backlog.md, Line: 41).
* **Network Configuration:** (See backlog item `[P2-4]` (Source: project backlog.md, Line: 42)) The settings page for managing IP, subnet, gateway, and DHCP settings (Source: project backlog.md, Line: 42). (See Section 10.2.3 for specific mandates).
* **System Storage & Log Management:** (See backlog item `[P2-5]` (Source: project backlog.md, Line: 43)) UI for viewing internal disk usage, log rotation, and cleaning caches (Source: project backlog.md, Line: 43).
* **USB Storage & Drive Management:** (See backlog item `[P2-6]` (Source: project backlog.md, Line: 44)) UI for mounting, unmounting, and formatting external USB drives for recording (Source: project backlog.md, Line: 44).
* **System Time (NTP) UI:** (See backlog item `[P2-7]` (Source: project backlog.md, Line: 45)) The settings page for managing time and Network Time Protocol (NTP) servers (Source: project backlog.md, Line: 45).
* **Shutdown & Reboot UI:** (See backlog item `[P2-8]` (Source: project backlog.md, Line: 46)) Essential safety buttons to gracefully power down or restart the appliance (Source: project backlog.md, Line: 46).
* **Factory Reset UI Trigger:** (See backlog item `[P2-9]` (Source: project backlog.md, Line: 47)) The "Danger Zone" button and confirmation modals to safely initiate a full system reset (Source: project backlog.md, Line: 47).
* **Core Appliance Upgrade UI:** (See backlog item `[P2-10]` (Source: project backlog.md, Line: 48)) The UI for triggering and monitoring core system updates (e.g., `apt` integration) (Source: project backlog.md, Line: 48). (See Section 12.0).
* **Routing Page UI:** (See backlog item `[P2-11]` (Source: project backlog.md, Line: 49)) The graphical page for managing audio/MIDI port connections (the non-`jackwire2` core routing) (Source: project backlog.md, Line: 49).
* **General "About" Page:** (See backlog item `[P3-9]` (Source: project backlog.md, Line: 71)) A simple page for displaying version, serial, and legal information (Source: project backlog.md, Line: 71).

#### 10.2.3 Network Configuration and PTP Policy

This defines the mandatory settings for multi-adapter management (Source: new mandates.md, Line: 24), focusing on stability and professional compliance (Source: new mandates.md, Line: 24). (See related backlog item `[P2-4]` (Source: project backlog.md, Line: 42)).

##### 10.2.3.1 Address Configuration Policy

| Setting | Audio Data Network (PTP/AES67) | Control/Management Network |
| :--- | :--- | :--- |
| **IPv4 Method** | **Static (Mandated)** (Source: new mandates.md, Line: 29) | **DHCP (Default)** (Source: new mandates.md, Line: 29) |
| **IPv6 State** | **Disabled (Default)** (Source: new mandates.md, Line: 30) | **Disabled (Default)** (Source: new mandates.md, Line: 30) |
| **PTP Role** | Defaults to a **PTP Slave** role in Domain 0 (Source: new mandates.md, Line: 31). | N/A (Source: new mandates.md, Line: 31) |
| **EEE (Energy Efficient Ethernet)** | Must be set to **OFF by default** and configurable (Source: new mandates.md, Line: 32). | Configurable (Source: new mandates.md, Line: 32). |

##### 10.2.3.2 Network Interface UI Mandate

* **Multi-Adapter:** The UI must support configuration for all available network interfaces via an **Adapter Selector** (Dropdown/Tabs) (Source: new mandates.md, Line: 36).
* **VLAN Tagging (Developer Level):** The VLAN configuration block must be restricted to the **Developer Feature Level** (Source: new mandates.md, Line: 37). It must allow the creation and independent IP configuration of **multiple 802.1Q tagged virtual interfaces** (VLAN IDs 1-4094) (Source: new mandates.md, Line: 38).
* **PTP Verification:** The UI **must** display the adapter's PTP capability status (Hardware PTP Supported, Software PTP Only, Not Available) in a color-coded format (Source: new mandates.md, Line: 39).

### 10.3 Master Controls Logic (Start All / Stop All)

The Central Broker will manage the complete state of both the "Start All" and "Stop All" buttons (Source: UX Design reference.md, Line: 281), including visibility and any visual cues like muting (Source: UX Design reference.md, Line: 281), based on the collective state of all currently enabled Broadcast modules (Source: UX Design reference.md, Line: 282).

* **Conditional Visibility:** The buttons are **only rendered** if the Broker reports that at least one `Broadcast` module is currently `isEnabled` (Source: UX Design reference.md, Line: 285). If zero `Broadcast` modules are enabled, the buttons are **hidden** (Source: UX Design reference.md, Line: 285).
* **State Management Rules:** (Source: UX Design reference.md, Line: 286)
    1.  **All enabled modules are broadcasting:** The **Stop All** button will be displayed in its **normal (illuminated) state** (Source: UX Design reference.md, Line: 287), while the **Start All** button will be **muted** (Source: UX Design reference.md, Line: 287).
    2.  **All enabled modules are stopped:** The **Start All** button will be displayed in its **normal (illuminated) state** (Source: UX Design reference.md, Line: 288), while the **Stop All** button will be **muted** (Source: UX Design reference.md, Line: 288).
    3.  **Mixed state (some broadcasting, some stopped):** Both buttons will be displayed in their **normal (illuminated) state** (Source: UX Design reference.md, Line: 289).
* **Action:** Triggers the `system.command.startAll` (one-to-many) Pub/Sub event (Source: UX Design reference.md, Line: 290).

### 10.4 "LIVE" Indicator

* **Implementation:** A single React component that subscribes *only* to the `system.live.status` (many-to-one) Pub/Sub topic (Source: UX Design reference.md, Line: 294).

### 10.5 Dynamic Dashboard Layout Editor

This feature provides users (at a designated **Feature Level** (Source: UX Design reference.md, Line: 297), e.g., **Advanced** and above (Source: UX Design reference.md, Line: 297)) the ability to fully customize their primary dashboard view (Source: UX Design reference.md, Line: 297).

* **Implementation:** Uses a **React Grid Layout** library (e.g., `react-grid-layout`) to manage draggable and resizable `Widget Components` (Source: UX Design reference.md, Line: 300).
* **Performance (Code Splitting):** The JavaScript and CSS for each `Widget Component` **must** be code-split (e.g., using `React.lazy()`) and loaded on-demand (Source: UX Design reference.md, Line: 301).
* **Configuration Storage:** All layout definitions (Widget ID, position `x, y`, width `w`, height `h`) are serialized into a **single JSON object** (Source: UX Design reference.md, Line: 302) and persisted on the **Go Backend** (Source: UX Design reference.md, Line: 303).
* **User Profiles:** The backend supports storing multiple, named layout configurations per user (Source: UX Design reference.md, Line: 304), allowing the user to switch between a **Monitoring Layout** and a **Control Layout** (Source: UX Design reference.md, Line: 304).
* **Default Layout:** A core, immutable **Default Layout** is stored and loaded if a user has not configured a custom layout (Source: UX Design reference.md, Line: 305).
* **Widget Component Wrapper & Boundary Enforcement:** Every Widget Component **must** be rendered inside a standard, size-responsive **Wrapper Component** (Source: UX Design reference.md, Line: 306) that enforces the correct styling (using Design Tokens) (Source: UX Design reference.md, Line: 307) and provides the drag handles (Source: UX Design reference.md, Line: 307). The wrapper component **must enforce strict boundaries** to prevent any plugin UI from rendering outside the defined widget space (Source: UX Design reference.md, Line: 307).
* **Backlog Item:** (See backlog item `[P3-3]` (Source: project backlog.md, Line: 65)) System-Wide "Snapshots" (Preset Management) UI for saving/recalling the entire system state (settings, routes, layout) (Source: project backlog.md, Line: 65).

#### 10.5.1 Layout Management Interface

To manage multiple layout profiles (User Profiles), a hybrid UI is mandated (Source: UX Design reference.md, Line: 310):

* **Main UI Header:** A **dropdown menu** must be present in the main UI (Source: UX Design reference.md, Line: 313). Its sole function is to **load or switch** between the user's available, named layouts (Source: UX Design reference.md, Line: 313).
* **Layout Editor ("Edit Mode"):** Controls for **"Save," "Save As...," "Rename," and "Delete"** **must only appear** within the layout editor context (Source: UX Design reference.md, Line: 314). This aligns with the "progressive complexity" philosophy (Source: UX Design reference.md, Line: 315).
* **RBAC Governance:** The behavior of this interface **must** be controlled by the RBAC module (Section 11.2) (Source: UX Design reference.md, Line: 316).

#### 10.5.2 Widget Data Flow (Real-time)

* **Control Plane Data:** The Central Broker utilizes the existing **Control Plane WebSocket** for targeted, low-frequency updates (e.g., listener count, on/off state) (Source: UX Design reference.md, Line: 320).
* **Subscription:** When a `Widget Component` loads, it immediately subscribes to a topic using its **unique `pluginID` as the prefix** (e.g., `com.icecast.streamer.widget.status`) (Source: UX Design reference.md, Line: 321).
* **Action Mechanism:** Widget control actions (e.g., clicking a "Mute" toggle) **must** communicate back to the Central Broker using a specific, targeted **Pub/Sub message** (e.g., `pluginID.command.action`) over the **Control Plane** socket (Source: UX Design reference.md, Line: 322).
* **Performance (Throttling):** The frontend's subscription hook for the **Control Plane** **must** reserve the right to **throttle or debounce** incoming messages to protect core UI stability (Source: UX Design reference.md, Line: 323).
* **Visualization Plane Data:** Widgets requiring high-frequency visualization data (e.g., level meters) **must** use the optional **Data Visualization Socket** (Section 5.4) (Source: UX Design reference.md, Line: 324). They **must not** send this data over the main Control Plane socket (Source: UX Design reference.md, Line: 324).

#### 10.5.3 Widget Keyboard Navigation (Roving `tabindex`)

To support full keyboard operability for the dynamic grid (as mandated in Section 9.3) (Source: UX Design reference.md, Line: 327), the **Widget Component Wrapper** must implement a "roving `tabindex`" pattern (Source: UX Design reference.md, Line: 327).

* **Grid Navigation (Level 1):** The user will navigate between widget wrappers using the **Arrow Keys** (Source: UX Design reference.md, Line: 330). The wrapper component is responsible for managing its `tabindex` (either `0` or `-1`) (Source: UX Design reference.md, Line: 330) and highlighting the active focus (Source: UX Design reference.md, Line: 330).
* **Widget-Internal Navigation (Level 2):** Pressing **`Enter`** or **`Tab`** on a focused wrapper will move focus to the first control *inside* the widget (Source: UX Design reference.md, Line: 331). The **`Tab`** key will then be "trapped" within the widget's controls (Source: UX Design reference.md, Line: 331). Pressing **`Escape`** will return focus to the widget wrapper (Source: UX Design reference.md, Line: 331).

#### 10.5.4 Orphaned Widget Handling

To prevent UI errors when a user-defined layout contains a widget for a disabled module (Source: UX Design reference.md, Line: 334), the following logic is mandated (Source: UX Design reference.md, Line: 334):

* **In "View Mode" (Default):** When the frontend loads the dashboard (Source: UX Design reference.md, Line: 337), if a widget in the layout JSON belongs to a module where `isEnabled` is `false` (Source: UX Design reference.md, Line: 337), the frontend **must render `null`** (Source: UX Design reference.md, Line: 337). This creates a non-destructive "hole" in the grid (Source: UX Design reference.md, Line: 337). If the module is re-enabled, the widget will reappear (Source: UX Design reference.md, Line: 337).
* **In "Edit Mode" (Layout Editor):** When the user enters the layout editor (Source: UX Design reference.md, Line: 338), the frontend **must render** a built-in **"Disabled Module Placeholder"** component in the widget's grid space (Source: UX Design reference.md, Line: 338). This gives the user a visual target to permanently delete the orphaned widget from their layout if they choose (Source: UX Design reference.md, Line: 338).

---

## 11.0 Security & Integrity

### 11.1 Sanitation (Prescriptive)

All user-generated input **must** be sanitized and validated on **both the frontend (client-side) and the backend (server-side)** (Source: UX Design reference.md, Line: 345). This "**Defense in Depth**" is a non-negotiable security requirement (Source: UX Design reference.md, Line: 345).

* **Backlog Item:** (See backlog item `[P3-7]` (Source: project backlog.md, Line: 69)) Client-Side Validation policy to provide immediate error feedback (Source: project backlog.md, Line: 69).

### 11.2 (Optional) Role-Based Access Control (RBAC)

* The system will support an optional, internal **Utility Module** for RBAC (Source: UX Design reference.md, Line: 350).
* If enabled, this module will manage user logins and permissions (Source: UX Design reference.md, Line: 351), protecting the UI from unauthorized access (Source: UX Design reference.md, Line: 351). (See related backlog item `[P1-1]` (Source: project backlog.md, Line: 18)).
* The RBAC module **must** also govern the behavior of the **Layout Management Interface (10.5.1)** (Source: UX Design reference.md, Line: 352). This includes, but is not limited to: (a) enabling full management for admin roles (Source: UX Design reference.md, Line: 352), (b) scoping layout management to a user's own profile (Source: UX Design reference.md, Line: 353), and (c) assigning specific, locked-down layouts to restricted roles (Source: UX Design reference.md, Line: 353).
* **Backlog Item:** (See backlog item `[P2-2]` (Source: project backlog.md, Line: 40)) The administrative UI for adding/deleting users, managing roles, and changing Feature Levels post-setup (Source: project backlog.md, Line: 40).

### 11.3 Granular Visibility Policy (Final Rule)

* **Mandate:** The **Central Broker (Backend)** is the sole source of truth for **all visibility and feature access** (Source: new mandates.md, Line: 64). It must filter configuration data based on the user's **Feature Level** and **Role Permissions** *before* that data leaves the appliance (Source: new mandates.md, Line: 65).
* **Interactive Mandate:** The administrator's form must treat the group header as a **Master Switch** that controls the visibility of all its children (Source: new mandates.md, Line: 66). If any child is assigned a permission, the parent visibility is automatically enabled (Boolean OR) (Source: new mandates.md, Line: 67).

### 11.4 Sparse Assignment Reset Policy (Critical Workflow)

This policy dictates the precise flow of permission persistence during administrative configuration (Source: new mandates.md, Line: 70):

* **Rule 1 (Hiding):** Unchecking a parent group is non-destructive (Source: new mandates.md, Line: 72); persistence for children is retained (Source: new mandates.md, Line: 72).
* **Rule 2 (Targeted Assignment):** When a specific permission is assigned to any single element (Source: new mandates.md, Line: 73), the system **must reset the persistence for all immediate sibling elements and subgroups** within that element's parent container to unchecked (zeroed out) (Source: new mandates.md, Line: 74).
* **Rule 3 (Cascading Reset):** If this action forces any ancestor group's visibility switch to turn ON (checked) (Source: new mandates.md, Line: 75), that ancestor group **must also trigger a reset** on all its own children (Source: new mandates.md, Line: 75).

### 11.5 Additional Security Mandates

* **Backlog Item:** (See backlog item `[P1-3]` (Source: project backlog.md, Line: 20)) Defines the process for quality assurance and prevents accidental merging of non-compliant code (Source: project backlog.md, Line: 20).
* **Backlog Item:** (See backlog item `[P1-4]` (Source: project backlog.md, Line: 21)) Security mandate (Source: project backlog.md, Line: 21). Prevents a rogue plugin or user from flooding the Broker via WebSocket (Source: project backlog.md, Line: 21).
* **Backlog Item:** (See backlog item `[P1-6]` (Source: project backlog.md, Line: 23)) Mandates a secondary security audit and defined scope for the `nixon-migrator` codebase (Source: project backlog.md, Line: 23).

---

## 12.0 Managed Appliance Upgrade System (APT-Centric)

This section formalizes the mandatory requirements and methodology for deploying in-place code upgrades to the Go-based audio appliance (Source: reference_appliance_backup.md, Line: 3). The design guarantees data integrity, configuration persistence, and controlled orchestration (Source: reference_appliance_backup.md, Line: 3) in both **Networked** and **Air-Gapped (Isolated)** environments (Source: reference_appliance_backup.md, Line: 4).

The foundation of this solution is **atomicity** (the process either succeeds completely or reverts fully to the last valid state) (Source: reference_appliance_backup.md, Line: 6).

> **Gemini's Annotated Commentary: Upgrade System Risk**
>
> This is the most complex and high-risk system defined in the documents (Source: reference_appliance_backup.md, devplan.md, project backlog.md).
> * **Privileged Executable:** The `nixon-migrator` (`[R7]`) (Source: reference_appliance_backup.md, Line: 28) is a privileged binary (Source: devplan.md, Line: 95). Backlog item `[P1-6]` (Privilege Management) (Source: project backlog.md, Line: 23) correctly identifies this as a "Critical Security" risk (Source: project backlog.md, Line: 23). This audit must be performed before implementation to define the absolute minimum set of commands the migrator can execute.
> * **IPC Protocol:** The communication between the Central Broker and the `nixon-migrator` must be secured (Source: project backlog.md, Line: 22). The RPC contract `[P1-8]` (Source: project backlog.md, Line: 25) should be minimal, likely limited to "start update," "get status," and "get logs."
> * **UI Lockout:** The "Integrity Guard" `[R5]` (Source: reference_appliance_backup.md, Line: 26) and "UI Integration (System Lockout)" `[P1-7]` (Source: project backlog.md, Line: 24) are critical. The main app *must* refuse to start if a lock file exists (Source: reference_appliance_backup.md, Line: 26), and the UI must display a "maintenance mode" screen (Source: project backlog.md, Line: 24) (served by the migrator itself or a lightweight proxy) during this time.

### 12.1 Foundational Requirements and Tooling (R1–R10)

All development and implementation must adhere strictly to the **Google Go Style Guide (R8)** (Source: reference_appliance_backup.md, Line: 11).

| ID | Requirement | Standard & Implementation |
| :--- | :--- | :--- |
| **R1** | **Go Migration Tooling** (Source: reference_appliance_backup.md, Line: 15) | **Industry Standard:** All schema and configuration changes must be executed using a dedicated, mature Go migration package (e.g., `golang-migrate/migrate`) (Source: reference_appliance_backup.md, Line: 15). |
| **R2** | **Configuration Version** (Source: reference_appliance_backup.md, Line: 16) | The `config.json` must contain a mandatory `"ConfigVersion": "X.Y.Z"` field (Source: reference_appliance_backup.md, Line: 16). |
| **R3** | **Non-Destructive Merging** (Source: reference_appliance_backup.md, Line: 17) | **Tool:** Use **`imdario/mergo`** to preserve existing user values and secrets over new defaults during configuration upgrades (Source: reference_appliance_backup.md, Line: 17). |
| **R4** | **Atomic DB Control** (Source: reference_appliance_backup.md, Line: 18) | **Integrity:** Requires an external **file lock** on the SQLite database + internal **atomic SQL transaction** for all schema changes (Source: reference_appliance_backup.md, Line: 18). |
| **R5** | **Integrity Guard** (Source: reference_appliance_backup.md, Line: 19) | The main application must **refuse to start** if a "migration in progress" lock file is detected (Anticipation Point F) (Source: reference_appliance_backup.md, Line: 19). |
| **R6** | **APT Upgrade Lock** (Source: reference_appliance_backup.md, Line: 20) | Use `apt-mark hold` to prevent uncontrolled, unattended upgrades (Anticipation Point B) (Source: reference_appliance_backup.md, Line: 20). |
| **R7** | **Companion Executable** (Source: reference_appliance_backup.md, Line: 21) | **Tool:** Create a privileged Go binary, **`nixon-migrator`** (Source: reference_appliance_backup.md, Line: 21), using the **`klauspost/compress`** package (Source: reference_appliance_backup.md, Line: 21) for secure OS commands (`apt-mark`, `dpkg -i`) (Source: reference_appliance_backup.md, Line: 21). |
| **R8** | **Coding Standard** (Source: reference_appliance_backup.md, Line: 22) | All code must adhere strictly to the **Google Go Style Guide** (Source: reference_appliance_backup.md, Line: 22). |
| **R9** | **Secure Local Distribution** (Source: reference_appliance_backup.md, Line: 23) | The Master must use **AES70 file transfer** (or SCP/SFTP over SSH as fallback) for encrypted, high-performance package distribution to all Slaves (Source: reference_appliance_backup.md, Line: 23). |
| **R10** | **Backup/Restore Feature (Migration)** (Source: reference_appliance_backup.md, Line: 24) | Implement a Web UI to export/import a version-stamped archive (`.zip`) containing `config.json` and `nixon.db` (Source: reference_appliance_backup.md, Line: 24). **The restore process must use the `nixon-migrator` to perform a version-based data and configuration migration** against the currently running binary (Source: reference_appliance_backup.md, Line: 24). |

### 12.2 Related Backlog & Development Tasks

* **Backlog Item:** (See backlog item `[P1-5]` (Source: project backlog.md, Line: 22)) Defines the robust communication channel between the Central Broker and the privileged `nixon-migrator` (Source: project backlog.md, Line: 22).
* **Backlog Item:** (See backlog item `[P1-6]` (Source: project backlog.md, Line: 23)) Mandates a secondary security audit and defined scope for the `nixon-migrator` codebase (Source: project backlog.md, Line: 23).
* **Backlog Item:** (See backlog item `[P1-7]` (Source: project backlog.md, Line: 24)) Defines the "System Maintenance" screen displayed when the `nixon-migrator` applies its lock file (Source: project backlog.md, Line: 24).
* **Backlog Item:** (See backlog item `[P1-8]` (Source: project backlog.md, Line: 25)) Creates the slimmed-down RPC interface the Broker uses to monitor the `nixon-migrator`'s status (Source: project backlog.md, Line: 25).
* **Task 4.1 (Go Migration):** Integrate a dedicated, mature Go migration package (e.g., `golang-migrate/migrate`) for database schema changes (R1) (Source: devplan.md, Line: 87).
* **Task 4.1 (Go Migration):** Implement an external read-write lock on the SQLite file using a suitable Go library before starting any migration (R4) (Source: devplan.md, Line: 88).
* **Task 4.1 (Go Migration):** Ensure all SQL upgrade steps run within a single, atomic internal transaction (R4) (Source: devplan.md, Line: 89).
* **Task 4.2 (Config Versioning):** Ensure `config.json` contains a mandatory `"ConfigVersion": "X.Y.Z"` field (R2) (Source: devplan.md, Line: 91).
* **Task 4.2 (Config Versioning):** Implement logic to use `imdario/mergo` for non-destructive merges (Source: devplan.md, Line: 92), prioritizing existing user-defined values and secrets over new defaults during configuration upgrades (R3) (Source: devplan.md, Line: 92).
* **Task 4.2 (Config Versioning):** The helper executable should read the `ConfigVersion` to determine necessary transformations (Source: devplan.md, Line: 93).
* **Task 4.3 (Companion Executable):** Create a privileged Go binary (`cmd/nixon-upgrade-helper/main.go`) (R7) (Source: devplan.md, Line: 95).
* **Task 4.3 (Companion Executable):** Implement command-line flags/subcommands for Debian package hook stages (`preinst`, `postinst`, `pre-remove`, `post-remove`) (Source: devplan.md, Line: 96).
* **Task 4.3 (Companion Executable):** This executable will be responsible for orchestrating upgrade/downgrade actions (Source: devplan.md, Line: 97).
* **Task 4.4 (Integrity Guard):** Implement a mechanism in `cmd/nixon/main.go` to check for a "migration in progress" lock file on startup (R5) (Source: devplan.md, Line: 99). If found, the application must refuse to start and prompt for recovery/revert (Source: devplan.md, Line: 100).
* **Task 4.4 (Integrity Guard):** Ensure the Debian package uses `apt-mark hold` to prevent uncontrolled, unattended upgrades (R6) (Source: devplan.md, Line: 101).
* **Task 4.5 (Debian Packaging):** Develop comprehensive `preinst`, `postinst`, `pre-remove`, and `post-remove` scripts for the Debian package (Source: devplan.md, Line: 103).
* **Task 4.5 (Debian Packaging):** These scripts will robustly call `nixon-upgrade-helper` with appropriate flags (Source: devplan.md, Line: 104), handle state transitions (Source: devplan.md, Line: 104), and ensure system integrity during all upgrade/downgrade phases (Source: devplan.md, Line: 104).
* **Task 4.5 (Debian Packaging):** Implement failure recovery and rollback mechanisms in `post-remove` to revert to the last known working state (Source: devplan.md, Line: 105).
* **Task 4.5 (Debian Packaging):** Ensure proper logging of upgrade steps and outcomes (Source: devplan.md, Line: 106), with detailed errors for debugging (Source: devplan.md, Line: 106).
* **Task 4.6 (Master-Slave Orchestration):** (See Section 12.4.4) (Source: devplan.md, Line: 108).
* **Task 4.7 (UI Trigger):** (See Section 12.4.4) (Source: devplan.md, Line: 116).

### 12.3 Orchestration and Failure Anticipation (Anticipation Points)

| ID | Point | Purpose |
| :--- | :--- | :--- |
| **A** | **Atomic Revert on Failure** (Source: reference_appliance_backup.md, Line: 36) | Guarantees the system restores the **last known working state** (data and config) if the migration fails (Source: reference_appliance_backup.md, Line: 36). |
| **F** | **Unattended Migration Recovery** (Source: reference_appliance_backup.md, Line: 37) | The system must detect an interrupted migration lock (R5) (Source: reference_appliance_backup.md, Line: 37) and automatically execute the restore procedure upon reboot (Source: reference_appliance_backup.md, Line: 37). |
| **G** | **Multi-Device Orchestration** (Source: reference_appliance_backup.md, Line: 38) | The Master unit directs all Slaves to update sequentially (Source: reference_appliance_backup.md, Line: 38), ensuring **Slave data/config is ignored** (Source: reference_appliance_backup.md, Line: 38) and Slaves enter a "Ready for Adoption" state (Source: reference_appliance_backup.md, Line: 38). |
| **H** | **Pre-Migration Readiness Check** (Source: reference_appliance_backup.md, Line: 39) | A mandatory dry run executed by the Master (Source: reference_appliance_backup.md, Line: 39), verifying permissions, locks, and Slave readiness before committing to the live upgrade (Source: reference_appliance_backup.md, Line: 39). |

### 12.4 Dual-Path Upgrade Methodology

The upgrade is orchestrated by the **Master Unit** and its **`nixon-migrator` (R7)** utility (Source: reference_appliance_backup.md, Line: 43).

#### 12.4.1 A. Pre-Migration Phase (Guaranteed Revert Point)

This phase guarantees an atomic restore point (Source: reference_appliance_backup.md, Line: 47) and is triggered by the `nixon-migrator` (R7) (Source: reference_appliance_backup.md, Line: 47).

1.  **Stop Service & Lock DB (R4):** `nixon-migrator` stops the main service and applies a mandatory file lock on the database (Source: reference_appliance_backup.md, Line: 49).
2.  **Backup (A):** Creates a secure, version-stamped backup of the active `config.json` (R2) and the entire database file (`nixon.db`) (Source: reference_appliance_backup.md, Line: 51).
3.  **Readiness Check (H):** Executes a dry run (Source: reference_appliance_backup.md, Line: 53), verifying permissions, locks, and confirming the Slave units are ready (Source: reference_appliance_backup.md, Line: 53).
4.  **Slave Lock (R6, G):** Master commands all Slaves (via AES70) to execute `apt-mark hold` locally (Source: reference_appliance_backup.md, Line: 55).

#### 12.4.2 B. Execution and Migration Phase (The Dual Path)

| Path | Convenience (Networked) | Security (Isolated/Air-Gapped) |
| :--- | :--- | :--- |
| **1. Package Staging** | Master executes `apt update` (R7) (Source: reference_appliance_backup.md, Line: 61). | User uploads the `.deb` file via the Master's UI (R10) (Source: reference_appliance_backup.md, Line: 61). Master verifies the signature and stages the file (Source: reference_appliance_backup.md, Line: 61). |
| **2. Distribution (R9)** | Slaves are commanded to execute `apt install` locally (Source: reference_appliance_backup.md, Line: 62). | Master uses **AES70/SCP** (R9) to push the staged `.deb` file to all Slaves' local storage (Source: reference_appliance_backup.md, Line: 62). |
| **3. Migration (R3, R4)** | `nixon-migrator` performs **Config Merge (R3)** and **DB Migration (R1, R4)** on the Master unit's local data (Source: reference_appliance_backup.md, Line: 63). | `nixon-migrator` performs **Config Merge (R3)** and **DB Migration (R1, R4)** on the Master unit's local data (Source: reference_appliance_backup.md, Line: 63). |
| **4. Slave Execution (G)** | Master commands Slaves to execute `apt upgrade` locally (Source: reference_appliance_backup.md, Line: 64). | Master commands Slaves to execute **`dpkg -i <filename>`** locally (Source: reference_appliance_backup.md, Line: 64). |
| **5. Final Cleanup (R5):** | If successful across all units, the temporary lock file is deleted, and all services are restarted (Source: reference_appliance_backup.md, Line: 65). | If successful across all units, the temporary lock file is deleted, and all services are restarted (Source: reference_appliance_backup.md, Line: 65). |

#### 12.4.3 C. Failure and Recovery (Atomic Rollback)

1.  The `nixon-migrator` (R7) is triggered by the package manager's failure hook (`postrm`) (Source: reference_appliance_backup.md, Line: 69).
2.  The executable restores the backup `config.json` and `nixon.db` files to the active directories (A) (Source: reference_appliance_backup.md, Line: 71).
3.  The system is restored to its exact, uncorrupted, pre-upgrade state (Source: reference_appliance_backup.md, Line: 73), guaranteeing data integrity (Source: reference_appliance_backup.md, Line: 73).

#### 12.4.4 Master-Slave Orchestration & UI Trigger Tasks

* **Task 4.6 (Master-Slave Orchestration):** **Master-Side Trigger:** Implement logic on the Master unit to detect, notify, and remotely trigger `apt upgrade` on Slave units (G) (Source: devplan.md, Line: 109). This might involve SSH or an internal API for secure communication (R9) (Source: devplan.md, Line: 110).
* **Task 4.6 (Master-Slave Orchestration):** **Slave-Side Coordination:** Slaves will update and enter a "Ready for Adoption" state after their `postinst` script completes (G) (Source: devplan.md, Line: 111).
* **Task 4.6 (Master-Slave Orchestration):** **Re-adoption:** The Master will re-adopt the updated Slaves via AES70 (once available) to re-establish control (Source: devplan.md, Line: 112).
* **Task 4.6 (Master-Slave Orchestration):** **Rollback:** Implement a rollback mechanism where the Master can revert Slaves to a previous working state if an upgrade fails on a Slave (A) (Source: devplan.md, Line: 113).
* **Task 4.7 (Optional UI-Triggered Upgrade):** Create a secure API endpoint on the Master unit that allows an administrator to initiate the upgrade process for specified Slaves (Source: devplan.md, Line: 117).
* **Task 4.7 (Optional UI-Triggered Upgrade):** This endpoint will call the underlying orchestration logic for Master-Slave upgrades (Source: devplan.md, Line: 118).
* **Task 4.7 (Optional UI-Triggered Upgrade):** Create a "System" page in the UI to display current versions, check for updates, and provide a button to trigger this process (Source: devplan.md, Line: 119). (See related backlog item `[P2-10]` (Source: project backlog.md, Line: 48)).

---

## 13.0 Core Feature Definitions

### 13.1 Recording Management

* **Backlog Item:** (See backlog item `[P2-1]` (Source: project backlog.md, Line: 39)) The UI for listing, downloading, starting/stopping, and deleting recordings (Source: project backlog.md, Line: 39).
* **Backlog Item:** (See backlog item `[P1-11]` (Source: project backlog.md, Line: 28)) Defines architecture for managing all configurable file paths, including Recording (Source: project backlog.md, Line: 28).

#### 13.1.1 Source, Control, and Location Mandates

| Mandate | Specification | Rationale |
| :--- | :--- | :--- |
| **Source** | **Full Multi-track.** (Source: new mandates.md, Line: 48) Must capture all individual audio streams (inputs, network sources, and all future internal mixer outputs/mixes) (Source: new mandates.md, Line: 48). | Provides essential flexibility for professional post-production (Source: new mandates.md, Line: 48). |
| **Control** | **START / STOP Only.** (Source: new mandates.md, Line: 49) **PAUSE/RESUME is disallowed.** (Source: new mandates.md, Line: 49) | Maintains file integrity and synchronization across multiple tracks (Source: new mandates.md, Line: 49). |
| **Default Location** | Local, relative directory: **`./recording`** (Source: new mandates.md, Line: 50). | Provides a reliable out-of-the-box target (Source: new mandates.md, Line: 50). |

> **Gemini's Annotated Commentary: Feature Conflict**
>
> A direct conflict exists regarding the recording "pause" functionality (Source: new mandates.md, Line: 49 / roadmap.md, Line: 43).
> * **The Mandate:** `new mandates.md` explicitly states, "**PAUSE/RESUME is disallowed**" (Source: new mandates.md, Line: 49). The rationale given is "Maintains file integrity and synchronization across multiple tracks" (Source: new mandates.md, Line: 49).
> * **The Conflict:** `roadmap.md` states in Phase 2: "Implement robust recording controls (start, stop, pause)" (Source: roadmap.md, Line: 43).
> * **Related Task:** `devplan.md` (Task 3.1 & 3.2) only mentions `StartRecording` and `StopRecording` (Source: devplan.md, Line: 71), with no mention of "pause."
>
> **Recommendation:**
> The `new mandates.md` document is presented as a "Consolidated Architectural Mandate" (Source: new mandates.md, Line: 3) and "Final Rule" (Source: new mandates.md, Line: 63), giving it precedence. The rationale for file integrity is a strong technical constraint. I recommend formally adopting the **START / STOP Only** mandate (Source: new mandates.md, Line: 49) and removing "pause" from the `roadmap.md` feature list (Source: roadmap.md, Line: 43).

#### 13.1.2 Auto-Record & File Management Policy

* **Auto-Record Trigger:** Must be signal-based (Source: new mandates.md, Line: 54), using a user-defined **Threshold (dBFS)** and **Pre-Roll duration** (circular buffer) on a selected audio input (Source: new mandates.md, Line: 55).
* **Smart Trimming:** The UI must provide a user-selectable option to **Trim Final Silence** (Source: new mandates.md, Line: 56), removing audio below a secondary threshold after the recording stops (Source: new mandates.md, Line: 56).
* **File Format Mandate:** The user selects a single **Primary Output Codec** (BWF, FLAC, MP3, etc.) (Source: new mandates.md, Line: 57).
* **Lazy Encoding:** This is an optional background process available for **any completed recording file** to convert it to a different format (e.g., BWF to FLAC) (Source: new mandates.md, Line: 58).
* **File Deletion:** Automatic deletion of the source file after successful lazy encoding is a **user-selectable option**, not a mandate (Source: new mandates.md, Line: 59).

#### 13.1.3 Development Tasks (Backend & Frontend)

* **Task 3.1 (Backend):** Implement `StartRecording`, `StopRecording` functions in `internal/control/manager.go` (Source: devplan.md, Line: 71).
* **Task 3.1 (Backend):** Implement `ListRecordings` to read from the designated recordings directory (Source: devplan.md, Line: 72).
* **Task 3.1 (Backend):** Implement `DeleteRecording` to remove a specified file (Source: devplan.md, Line: 73).
* **Task 3.1 (Backend):** Wire these new functions to WebSocket commands in `internal/websocket/websocket.go` (Source: devplan.md, Line: 74).
* **Task 3.2 (Frontend):** Implement state management for recording status (e.g., recording, stopped) in `RecordingControl.jsx` (Source: devplan.md, Line: 77).
* **Task 3.2 (Frontend):** Implement the UI buttons to send `start` and `stop` recording commands (Source: devplan.md, Line: 78).
* **Task 3.2 (Frontend):** In `RecordingsList.jsx`, implement the logic to fetch the list of recordings on component mount (Source: devplan.md, Line: 79).
* **Task 3.2 (Frontend):** Add UI elements to display each recording with buttons for download and delete (Source: devplan.md, Line: 80).

### 13.2 Audio, MIDI, & Routing

* **Backlog Item:** (See backlog item `[P2-11]` (Source: project backlog.md, Line: 49)) The graphical page for managing audio/MIDI port connections (the non-`jackwire2` core routing) (Source: project backlog.md, Line: 49).
* **Backlog Item:** (See backlog item `[P3-1]` (Source: project backlog.md, Line: 63)) The top-level feature for creative audio control (Internal Digital Mixer & DSP) (Source: project backlog.md, Line: 63).
* **Backlog Item:** (See backlog item `[P3-2]` (Source: project backlog.md, Line: 64)) Professional feature for checking audio quality remotely via a WebRTC/WebSocket stream (Source: project backlog.md, Line: 64).
* **Backlog Item:** (See backlog item `[P3-5]` (Source: project backlog.md, Line: 67)) Fleshing out the MIDI feature (device manager, routing, monitor) into a full platform (Source: project backlog.md, Line: 67).

#### 13.2.1 Development Tasks (Frontend)

* **Task 3.3 (Frontend):** Create a new "Routing" view/component in the React application (Source: devplan.md, Line: 82).
* **Task 3.3 (Frontend):** Fetch and display lists of available audio/MIDI sources and sinks from the `jackwire2` plugin (Source: devplan.md, Line: 83).
* **Task 3.3 (Frontend):** Implement a graphical interface (e.g., clickable lists, drag-and-drop) to allow users to create and remove connections between ports via the `jackwire2` plugin API (Source: devplan.md, Line: 84).
* **Roadmap Item:** Create a "Routing" tab in the UI to visualize and manage audio/MIDI streams through this plugin (Source: roadmap.md, Line: 44).
* **Roadmap Item:** Allow users to connect and disconnect audio/MIDI ports graphically (Source: roadmap.md, Line: 45).

### 13.3 Advanced Monitoring

* **Task 5.1:** Integrate a system statistics library (e.g., gopsutil) to gather CPU, memory, and disk usage (Source: devplan.md, Line: 124).
* **Task 5.1:** Stream these statistics over the WebSocket connection (Source: devplan.md, Line: 125).
* **Task 5.1:** Create the `DiskUsage.jsx` component and other monitoring widgets to visualize this data in the UI (Source: devplan.md, Line: 126).
* **Roadmap Item:** Provide detailed system monitoring in the UI (Source: roadmap.md, Line: 66), including CPU, memory, and disk usage (Source: roadmap.md, Line: 66), as well as real-time audio metering (Source: roadmap.md, Line: 66).
* **Backlog Item:** (See backlog item `[P2-5]` (Source: project backlog.md, Line: 43)) UI for viewing internal disk usage, log rotation, and cleaning caches (Source: project backlog.md, Line: 43).

---

## 14.0 Network Protocols & Multi-Device Control

### 14.1 AES70 (Core)

> **Gemini's Annotated Commentary: Phased Implementation**
>
> The plan for AES70 is correctly phased, which is a sound project management approach.
> * **Phase 1 (Foundation):** Task 2.1 (Source: devplan.md, Line: 48) and the corresponding roadmap item (Source: roadmap.md, Line: 39) establish the Go interfaces and architectural hooks. This "API-first" approach is critical. It allows other components that depend on AES70 (like the Upgrade System's "Re-adoption" step (Source: devplan.md, Line: 112)) to be developed in parallel against a stable, non-functional, interface.
> * **Phase 2 (Full Implementation):** Task 6.1 (Source: devplan.md, Line: 139) and the roadmap (Source: roadmap.md, Line: 73) correctly defer the complex protocol implementation until *after* all other local, single-device features are operational (Source: devplan.md, Line: 142). This de-risks the project by not blocking core functionality on this highly complex component.

* **Task 2.1 (Foundation):** Define the fundamental Go interfaces, data structures, and the `internal/aes70` package structure required to represent AES70 devices, their properties, and control commands (Source: devplan.md, Line: 49).
* **Task 2.1 (Foundation):** Create placeholder methods for key functionalities like device discovery, parameter get/set, and event handling (Source: devplan.md, Line: 50).
* **Task 2.1 (Foundation):** This task focuses on establishing the API contract and architectural hooks for AES70 (Source: devplan.md, Line: 51), making other components "ready" for its eventual full implementation (Source: devplan.md, Line: 51).
* **Task 2.1 (Foundation):** *No actual AES70 protocol implementation is done in this task* (Source: devplan.md, Line: 52).
* **Roadmap Item (Foundation):** Establish the foundational Go interfaces and architectural hooks within `internal/aes70` (Source: roadmap.md, Line: 39) to enable future, full native Go AES70 protocol implementation (Source: roadmap.md, Line: 39). This focuses on design and readiness, not full protocol logic (Source: roadmap.md, Line: 40). This is a critical prerequisite for future multi-device control (Source: roadmap.md, Line: 40).
* **Task 6.1 (Full Implementation):** Develop a comprehensive native Go library within the `internal/aes70` package (Source: devplan.md, Line: 140) based on the provided PDF reference documents and analysis of the JavaScript client/test tool (Source: devplan.md, Line: 140).
* **Task 6.1 (Full Implementation):** Implement full AES70 device discovery, state management, command/control, and event handling (Source: devplan.md, Line: 141), integrating with the interfaces defined in Task 2.1 (Source: devplan.md, Line: 141).
* **Task 6.1 (Full Implementation):** *This task is executed after all other local features and single-device networking are operational* (Source: devplan.md, Line: 142).
* **Roadmap Item (Full Implementation):** Develop the comprehensive native Go library for AES70 based on provided reference documents and client tools (Source: roadmap.md, Line: 74), implementing full device discovery, state management, command/control (Source: roadmap.md, Line: 74). This critical feature will be implemented after all other local features and single-device networking are operational (Source: roadmap.md, Line: 75).

### 14.2 AES67 / AVB (via PipeWire)

* **Task 5.2 (AES67):** Integrate a Go library or PipeWire integration method for AES67 (Source: devplan.md, Line: 128).
* **Task 5.2 (AES67):** Develop a new plugin (`internal/plugins/aes67`) that leverages PipeWire to enable single-device AES67 audio transport (Source: devplan.md, Line: 129).
* **Task 5.2 (AES67):** Expose configuration and control options for AES67 via the plugin API (Source: devplan.md, Line: 130).
* **Task 5.3 (AVB):** Develop a new plugin (`internal/plugins/avb`) that leverages the existing PipeWire plugin for AVB to enable single-device AVB audio transport (Source: devplan.md, Line: 132).
* **Task 5.3 (AVB):** Expose configuration and control options for AVB via the plugin API (Source: devplan.md, Line: 133).
* **Task 5.3 (AVB):** *This task focuses on single-device AVB capabilities, utilizing the PipeWire plugin* (Source: devplan.md, Line: 134).
* **Roadmap Item (AVB/AES67):** Implement AVB and AES67 support as plugins (Source: roadmap.md, Line: 68), leveraging PipeWire for single-device audio transport capabilities (Source: roadmap.md, Line: 68).

### 14.3 Multi-Device Discovery & Control

* **Task 3.4 (mDNS Discovery):** Integrate a Go mDNS/zeroconf library (Source: devplan.md, Line: 86).
* **Task 3.4 (mDNS Discovery):** Register the Nixon service (`_nixon._tcp`) when the application starts (Source: devplan.md, Line: 87).
* **Task 3.4 (mDNS Discovery):** Implement a discovery mechanism to continuously scan for other Nixon services (Source: devplan.md, Line: 88).
* **Task 3.4 (mDNS Discovery):** Create a new UI component to display a list of discovered Nixon appliances on the network (Source: devplan.md, Line: 89).
* **Task 3.4 (mDNS Discovery):** *This task focuses solely on local network discovery via mDNS, without AES70 control* (Source: devplan.md, Line: 90).
* **Roadmap Item (mDNS Discovery):** Implement mDNS for automatic discovery of other Nixon appliances on the local network (Source: roadmap.md, Line: 63).
* **Roadmap Item (mDNS Discovery):** Create a UI component to display a list of discovered Nixon appliances (Source: roadmap.md, Line: 64). This task focuses solely on local network discovery via mDNS, without full AES70 control (Source: roadmap.md, Line: 64).
* **Task 6.2 (Full AES70 Control):** Develop the backend and frontend logic for master-slave control (Source: devplan.md, Line: 144), allowing a primary Nixon unit to manage and orchestrate other Nixon units via the fully implemented AES70 (Source: devplan.md, Line: 145).
* **Task 6.2 (Full AES70 Control):** This includes features like synchronizing settings, initiating actions across devices, and monitoring their status (Source: devplan.md, Line: 146).
* **Task 6.2 (Full AES70 Control):** *This task is executed after the full AES70 protocol implementation* (Source: devplan.md, Line: 147).
* **Roadmap Item (Full AES70 Control):** Implement master-slave control and orchestration (Source: roadmap.md, Line: 77), allowing a primary Nixon unit to manage and synchronize other Nixon units via the fully implemented AES70 protocol (Source: roadmap.md, Line: 78).

---

## 15.0 First-Time Setup & Factory Reset

### 15.1 First-Time Setup (Configuration Wizard)

* **Trigger:** If the backend provides the appropriate "first start" configuration flag (Source: UX Design reference.md, Line: 358), the frontend will load the standard interface and immediately display a **modal window** containing the setup wizard (Source: UX Design reference.md, Line: 358).
* **Core Design:** This wizard is **optional** and intended to guide users, not block them (Source: UX Design reference.md, Line: 360). The system is ready for use out-of-the-box with "Standard" defaults (Source: UX Design reference.md, Line: 360).
    * A **"Cancel"** button is always visible in the top right of the modal (Source: UX Design reference.md, Line: 362).
    * "Next" / "Back" buttons are provided on each page as appropriate (Source: UX Design reference.md, Line: 363).
* **Completion Logic:** Clicking "Cancel" at any step or "Finish" at the end **must** signal to the backend that the setup wizard is complete (Source: UX Design reference.md, Line: 364). The frontend will then close the modal and continue running (Source: UX Design reference.md, Line: 365).

### 15.2 Wizard Page Flow

1.  **Page 1: Welcome & Core Configuration** (Source: UX Design reference.md, Line: 369)
    * Contains a welcome note (Source: UX Design reference.md, Line: 371).
    * **Feature Level:** Guides the user to choose their desired **Feature Level** (Standard, Advanced, Professional) (Source: UX Design reference.md, Line: 372), with a brief description of each (Source: UX Design reference.md, Line: 372).
        * Includes a note that the Feature Level can be changed at any time in the settings menu (Source: UX Design reference.md, Line: 373).
        * If the user cancels at this step, the frontend will notify the backend of completion and continue to run with the default "Standard" Feature Level (Source: UX Design reference.md, Line: 374).
    * **Hostname:** Provides a text input to configure the system's hostname (Source: UX Design reference.md, Line: 375).

2.  **Page 2: Audio Interface** (Source: UX Design reference.md, Line: 377)
    * Prompts the user to configure the primary audio interface (Source: UX Design reference.md, Line: 379).
    * The level of detail on this page **must** be dependent on the **Feature Level** selected in Page 1 (Source: UX Design reference.md, Line: 380).
    * **Standard Level:** Presents a simple dropdown of all available audio hardware (Source: UX Design reference.md, Line: 381), a dropdown for sample rate (Source: UX Design reference.md, Line: 381), and a dropdown for bit depth (Source: UX Design reference.md, Line: 381).
    * **OEM Appliances:** This step may be pre-configured and show the current settings (Source: UX Design reference.md, Line: 382).
    * **Defaults:** If not pre-configured, the system will default to `44.1kHz` and `16-bit` (Source: UX Design reference.md, Line: 383).

3.  **Page 3: Access & Security (RBAC)** (Source: UX Design reference.md, Line: 385)
    * Prompts the user to configure security settings (Source: UX Design reference.md, Line: 387), with a (discouraged) option to skip this step (Source: UX Design reference.md, Line: 387).
    * **Standard Level:** A simple form to set a username, real name, and set/confirm a password for the primary user (Source: UX Design reference.md, Line: 388).
    * **Advanced / Professional Levels:** Includes the Standard-level form (Source: UX Design reference.md, Line: 389), plus an option to add other users and assign one of the default roles (Source: UX Design reference.md, Line: 389).

4.  **Page 4: Finish** (Source: UX Design reference.md, Line: 391)
    * A final page with a congratulatory note, well wishes, and relevant support information (Source: UX Design reference.md, Line: 393).

* **Extensibility:** This core wizard is *not* designed to be extended by third-party plugins (Source: UX Design reference.md, Line: 397), as it handles fundamental system setup (Source: UX Design reference.md, Line: 397).

### 15.3 Backend State Machine Logic (Setup & Reset)

* **`systeminitialize` State Machine:** The backend must read a `systeminitialize` flag from `config.json` (Source: State machine Logic - factory reset.md, Line: 11).
    * **State `3` (OEM First Boot):** On detecting this state, the backend must (Source: State machine Logic - factory reset.md, Line: 12):
        1.  Generate a unique hostname (format: `nixon-[12-hex-chars]`) (Source: State machine Logic - factory reset.md, Line: 13).
        2.  Save this hostname to a new, separate config file named `nixonhost` (Source: State machine Logic - factory reset.md, Line: 14).
        3.  Set the system's actual hostname using `hostnamectl` (Source: State machine Logic - factory reset.md, Line: 15).
        4.  Update `config.json` to set `systeminitialize = 2` (Source: State machine Logic - factory reset.md, Line: 16).
        5.  Trigger a full system reboot (Source: State machine Logic - factory reset.md, Line: 17).

    * **State `2` (Factory Reset):** On detecting this state, the backend must (Source: State machine Logic - factory reset.md, Line: 19):
        1.  Delete `studio.db` and all files in the `./recordings` directory (Source: State machine Logic - factory reset.md, Line: 20).
        2.  Restore the factory default `config.json` (which must have `systeminitialize = 1`) (Source: State machine Logic - factory reset.md, Line: 21).
        3.  Trigger a graceful restart of the application service (Source: State machine Logic - factory reset.md, Line: 22).

    * **State `1` (User Setup):** On detecting this state, the web server must serve a dedicated "Setup Wizard" UI instead of the main application (Source: State machine Logic - factory reset.md, Line: 24).

* **Setup Wizard UI (Backend Task):** This UI will guide the user through (Source: State machine Logic - factory reset.md, Line: 26):
    1.  Selecting an expertise level (Standard, Power User, Professional) to set the UI mode (Source: State machine Logic - factory reset.md, Line: 27).
    2.  Setting a custom hostname (pre-filled with the generated one) (Source: State machine Logic - factory reset.md, Line: 28).
    3.  Creating an optional administrator account (display a persistent warning on the main UI if skipped) (Source: State machine Logic - factory reset.md, Line: 29).
    4.  Selecting the primary audio device (Source: State machine Logic - factory reset.md, Line: 30).
    
    Upon completion or skipping, it must call an API endpoint that sets `systeminitialize = 0` (Source: State machine Logic - factory reset.md, Line: 32).
* **Backlog Item:** (See backlog item `[P2-9]` (Source: project backlog.md, Line: 47)) The "Danger Zone" button and confirmation modals to safely initiate a full system reset (Source: project backlog.md, Line: 47).

---

## 16.0 Appendix A: Keyed Project Backlog

This is the complete 33-item list from `project backlog.md` (Source: project backlog.md, Line: 1), with unique keys assigned for tracking and cross-referencing within this document.

### I. Foundational & Architectural Mandates (P1)

(11 Items) (Source: project backlog.md, Line: 4)
These are non-negotiable prerequisites that must be addressed to resolve major security, integrity, and stability risks (Source: project backlog.md, Line: 5).

| Key | Feature / Gap | Rationale | Priority |
| :--- | :--- | :--- | :--- |
| **[P1-1]** | **Authentication UI & Flow** (Source: project backlog.md, Line: 10) | Critical security gap (Source: project backlog.md, Line: 10). Defines the login/logout mechanism and token management (Source: project backlog.md, Line: 10). | **P1 (Critical Security)** (Source: project backlog.md, Line: 10) |
| **[P1-2]** | **Frontend Build & Deployment Strategy** (Source: project backlog.md, Line: 11) | Defines how React assets are bundled (e.g., embedded) to ensure compliance with the Upgrade Document (Source: project backlog.md, Line: 11). | **P1 (Critical Architecture)** (Source: project backlog.md, Line: 11) |
| **[P1-3]** | **Code Review and Merge Policy** (Source: project backlog.md, Line: 12) | Defines the process for quality assurance and prevents accidental merging of non-compliant code (Source: project backlog.md, Line: 12). | **P1 (Critical Process)** (Source: project backlog.md, Line: 12) |
| **[P1-4]** | **API Rate Limiting Policy** (Source: project backlog.md, Line: 13) | Security mandate (Source: project backlog.md, Line: 13). Prevents a rogue plugin or user from flooding the Broker via WebSocket (Source: project backlog.md, Line: 13). | **P1 (Critical Security)** (Source: project backlog.md, Line: 13) |
| **[P1-5]** | **Secure IPC Protocol** (Source: project backlog.md, Line: 14) | Defines the robust communication channel between the Central Broker and the privileged `nixon-migrator` (Source: project backlog.md, Line: 14). | **P1 (Critical Integrity)** (Source: project backlog.md, Line: 14) |
| **[P1-6]** | **Privilege Management** (Source: project backlog.md, Line: 15) | Mandates a secondary security audit and defined scope for the `nixon-migrator` codebase (Source: project backlog.md, Line: 15). | **P1 (Critical Security)** (Source: project backlog.md, Line: 15) |
| **[P1-7]** | **UI Integration (System Lockout)** (Source: project backlog.md, Line: 16) | Defines the "System Maintenance" screen displayed when the `nixon-migrator` applies its lock file (Source: project backlog.md, Line: 16). | **P1 (Critical Integrity)** (Source: project backlog.md, Line: 16) |
| **[P1-8]** | **Define RPC Contract (Migrator)** (Source: project backlog.md, Line: 17) | Creates the slimmed-down RPC interface the Broker uses to monitor the `nixon-migrator`'s status (Source: project backlog.md, Line: 17). | **P1 (Critical Architecture)** (Source: project backlog.md, Line: 17) |
| **[P1-9]** | **localStorage Schema Versioning** (Source: project backlog.md, Line: 18) | Future-proofing mandate (Source: project backlog.md, Line: 18). Prevents the frontend from crashing on software updates with stale local data (Source: project backlog.md, Line: 18). | **P1 (Critical Future-Proofing)** (Source: project backlog.md, Line: 18) |
| **[P1-10]** | **License Compliance Strategy** (Source: project backlog.md, Line: 19) | Defines the legal process for auditing and complying with all third-party software licenses (Source: project backlog.md, Line: 19). | **P1 (Critical Process)** (Source: project backlog.md, Line: 19) |
| **[P1-11]** | **Global Path Configuration Mandate** (Source: project backlog.md, Line: 20) | (Integrated) Defines architecture for managing all configurable file paths (Recording, Logs, Config) via a single Path Resolver Service (PRS) (Source: project backlog.md, Line: 20). | **P1 (Critical Integrity)** (Source: project backlog.md, Line: 20) |

### II. Core Feature Development (P2)

(12 Items) (Source: project backlog.md, Line: 24)
These are essential, missing, user-facing features required for the appliance to function as a complete V1 product (Source: project backlog.md, Line: 25).

| Key | Feature / Gap | Rationale | Priority |
| :--- | :--- | :--- | :--- |
| **[P2-1]** | **Recording Management Page** (Source: project backlog.md, Line: 30) | The UI for listing, downloading, starting/stopping, and deleting recordings (Source: project backlog.md, Line: 30). | **P2 (Immediate Core)** (Source: project backlog.md, Line: 30) |
| **[P2-2]** | **User Management Page (RBAC)** (Source: project backlog.md, Line: 31) | The administrative UI for adding/deleting users, managing roles, and changing Feature Levels post-setup (Source: project backlog.md, Line: 31). | **P2 (Core System)** (Source: project backlog.md, Line: 31) |
| **[P2-3]** | **System Log Viewer** (Source: project backlog.md, Line: 32) | The UI for tailing and viewing structured logs (Source: project backlog.md, Line: 32); essential for debugging the headless appliance (Source: project backlog.md, Line: 32). | **P2 (Core System)** (Source: project backlog.md, Line: 32) |
| **[P2-4]** | **Network Configuration UI** (Source: project backlog.md, Line: 33) | The settings page for managing IP, subnet, gateway, and DHCP settings (Source: project backlog.md, Line: 33). | **P2 (Core Appliance)** (Source: project backlog.md, Line: 33) |
| **[P2-5]** | **System Storage & Log Management** (Source: project backlog.md, Line: 34) | UI for viewing internal disk usage, log rotation, and cleaning caches (Source: project backlog.md, Line: 34). | **P2 (Core Appliance)** (Source: project backlog.md, Line: 34) |
| **[P2-6]** | **USB Storage & Drive Management** (Source: project backlog.md, Line: 35) | UI for mounting, unmounting, and formatting external USB drives for recording (Source: project backlog.md, Line: 35). | **P2 (Core Appliance)** (Source: project backlog.md, Line: 35) |
| **[P2-7]** | **System Time (NTP) UI** (Source: project backlog.md, Line: 36) | The settings page for managing time and Network Time Protocol (NTP) servers (Source: project backlog.md, Line: 36). | **P2 (Core Appliance)** (Source: project backlog.md, Line: 36) |
| **[P2-8]** | **Shutdown & Reboot UI** (Source: project backlog.md, Line: 37) | Essential safety buttons to gracefully power down or restart the appliance (Source: project backlog.md, Line: 37). | **P2 (Core Appliance)** (Source: project backlog.md, Line: 37) |
| **[P2-9]** | **Factory Reset UI Trigger** (Source: project backlog.md, Line: 38) | The "Danger Zone" button and confirmation modals to safely initiate a full system reset (Source: project backlog.md, Line: 38). | **P2 (Core Appliance)** (Source: project backlog.md, Line: 38) |
| **[P2-10]** | **Core Appliance Upgrade UI** (Source: project backlog.md, Line: 39) | The UI for triggering and monitoring core system updates (e.g., `apt` integration) (Source: project backlog.md, Line: 39). | **P2 (Core System)** (Source: project backlog.md, Line: 39) |
| **[P2-11]** | **Routing Page UI** (Source: project backlog.md, Line: 40) | The graphical page for managing audio/MIDI port connections (the non-`jackwire2` core routing) (Source: project backlog.md, Line: 40). | **P2 (Core Feature)** (Source: project backlog.md, Line: 40) |
| **[P2-12]** | **Schema Enhancement (Read-Only Fields)** (Source: project backlog.md, Line: 41) | Adding types like read-only-text and progress-bar to the form schema to build dynamic status pages (Source: project backlog.md, Line: 41). | **P2 (Tooling)** (Source: project backlog.md, Line: 41) |

### III. V2+ / Advanced Enhancements (P3)

(10 Items) (Source: project backlog.md, Line: 45)
These items are advanced features that improve workflow or long-term product longevity and are scheduled for V2 or later (Source: project backlog.md, Line: 46).

| Key | Feature / Gap | Rationale | Priority |
| :--- | :--- | :--- | :--- |
| **[P3-1]** | **Internal Digital Mixer & DSP** (Source: project backlog.md, Line: 51) | The top-level feature for creative audio control (Source: project backlog.md, Line: 51). | **P3 (Creative Core)** (Source: project backlog.md, Line: 51) |
| **[P3-2]** | **Remote Audio Monitoring ("Listen-In")** (Source: project backlog.md, Line: 52) | Professional feature for checking audio quality remotely via a WebRTC/WebSocket stream (Source: project backlog.md, Line: 52). | **P3 (Creative Core)** (Source: project backlog.md, Line: 52) |
| **[P3-3]** | **System-Wide "Snapshots" (Preset Management)** (Source: project backlog.md, Line: 53) | UI for saving/recalling the entire system state (settings, routes, layout) (Source: project backlog.md, Line: 53). | **P3 (UX/Workflow)** (Source: project backlog.md, Line: 53) |
| **[P3-4]** | **Real-Time Collaboration & Presence** (Source: project backlog.md, Line: 54) | UI to show which users are currently editing a page (Source: project backlog.md, Line: 54). | **P3 (UX/Collaboration)** (Source: project backlog.md, Line: 54) |
| **[P3-5]** | **Full MIDI Feature Definition** (Source: project backlog.md, Line: 55) | Fleshing out the MIDI feature (device manager, routing, monitor) into a full platform (Source: project backlog.md, Line: 55). | **P3 (Feature Depth)** (Source: project backlog.md, Line: 55) |
| **[P3-6]** | **Optimistic UI Updates** (Source: project backlog.md, Line: 56) | Design policy to make the UI feel instantaneous (Source: project backlog.md, Line: 56). | **P3 (Polish)** (Source: project backlog.md, Line: 56) |
| **[P3-7]** | **Client-Side Validation** (Source: project backlog.md, Line: 57) | Policy to provide immediate error feedback before sending data to the server (Source: project backlog.md, Line: 57). | **P3 (Polish)** (Source: project backlog.md, Line: 57) |
| **[P3-8]** | **Localization/Translation Strategy** (Source: project backlog.md, Line: 58) | The UI to select and manage the display language (Source: project backlog.md, Line: 58). | **P3 (Global)** (Source: project backlog.md, Line: 58) |
| **[P3-9]** | **General "About" Page** (Source: project backlog.md, Line: 59) | A simple page for displaying version, serial, and legal information (Source: project backlog.md, Line: 59). | **P3 (UX)** (Source: project backlog.md, Line: 59) |
| **[P3-10]** | **Server-Side User Preferences** (Source: project backlog.md, Line: 60) | Persistently saving user settings (theme, default feature level) on the backend (Source: project backlog.md, Line: 60). | **P3 (UX)** (Source: project backlog.md, Line: 60) |
