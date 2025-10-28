# Project Design Document: Embedded Appliance UI

## 1.0 Core Philosophy & Guiding Principles

This document defines the front-end architecture for a **modular, extensible, and resource-efficient** user interface for an embedded appliance. The UI will manage **all features and functions of the application**, including but not limited to Broadcast, Transport, Role-Based Access Control (RBAC), MIDI, and Audio Mixing modules.

The core philosophy is built on **Strict Decoupling** and **Resource Conservation**. The front-end (**React**) must be a "**dumb**" **rendering layer** that is dynamically driven by a "**smart**" **backend (Go)**. The failure or disabling of any single module, especially third-party modules, must not impact the stability or performance of the core system or other modules.

The UI must also scale in complexity, presenting a simplified interface for standard users and progressively revealing advanced controls for professionals and developers, all from a single, unified codebase.

-----

## 2.0 Core Technology Stack & Standards

| Component | Standard | Rationale |
| :--- | :--- | :--- |
| **Framework** | **React** (Latest Stable Version) | Industry-standard, component-based, and performant. |
| **Styling** | **Tailwind CSS** | A utility-first framework that allows for rapid, custom styling and seamless responsive design (mobile/desktop) without pre-defined UI kits. |
| **Style Guide** | **Airbnb React/JSX Style Guide** | Provides a comprehensive, industry-standard set of rules for code consistency, enforced via **ESLint**. |
| **Form Handling** | **React Hook Form** | A performant library for managing form state. The backend API schema is specifically designed to match this library's requirements to eliminate a data-mapping layer on the frontend. |

### 2.1 Theming & Design Token Contract

To ensure visual consistency across all core UI and dynamically rendered plugin components, the frontend **must** implement a theme using **CSS Custom Properties** (Variables).

  * **Design Tokens:** All core theme attributes (colors, fonts, spacing, radii) must be defined as CSS variables (e.g., `var(--color-primary-main)`, `var(--color-warning-bg)`).
  * **Dynamic Components:** All dynamically rendered components (e.g., `info-box` from the Schema Contract) **must** use these CSS variables for styling.
  * **Benefit:** This creates a simple visual API, ensures all plugin-defined UI matches the core application's look and feel, and allows for seamless theme switching (e.g., **Light/Dark Mode**).

-----

## 3.0 Prescriptive Backend Architecture (Frontend-Impacting)

To achieve the goals of stability and simplicity, the front-end design *requires* the backend to adhere to the following architecture.

### 3.1 Plugin Engine: HashiCorp `go-plugin`

All extensible modules, both internal and third-party, **must** be managed by the `go-plugin` library.

  * **Benefit (Process Isolation):** This runs each plugin in a **separate OS process**. A crash or memory leak in one plugin will terminate only its own process, ensuring the core broker and all other modules remain stable. This is a critical requirement.
  * **Latency Note:** This abstraction applies to the **control plane** (settings, commands, status updates) only. High-bandwidth, low-latency media streams (e.g., PipeWire) will follow their own data plane, as defined by their respective modules.

### 3.2 The Central Broker Model

All communication between the React frontend and *any* module **must** be proxied through a single **Central Broker** component in the core Go application.

  * **Benefit (Total Abstraction):** The React frontend is completely unaware of whether a module is internal (a direct Go function call) or external (an RPC call via go-plugin). This creates a single, simple, and secure API surface.

### 3.3 API Communication Protocols

  * **Initial State Load:** A single **HTTP/S GET** request upon application load to retrieve the complete initial state (Section 3.4).
  * **Control Plane:** A single, persistent **WebSocket** connection. All real-time data, actions, settings commands, and status updates **must** be handled via a Pub/Sub message model over this connection.
  * **Data Visualization Plane (Optional):** A second, optional **WebSocket** connection may be initiated by specific widgets. This socket is exclusively for streaming high-frequency, non-critical visualization data (e.g., RMS, FFT data in binary format) from the backend. This data **must not** be sent over the main Control Plane socket.
  * **Authentication:** All API communication (HTTP GET and all WebSockets) will be secured using the existing token-based authentication system.

#### 3.3.1 Data Concurrency & Stale State Protection

To prevent data-loss from concurrent user sessions (e.g., two users editing the same form), all state-changing **WebSocket messages** (e.g., saving settings) **must** use a concurrency token.

1.  **Initial State:** The initial HTTP GET request **must** provide a concurrency token (e.g., `"_version": 123`) with any stateful data (like a settings form).
2.  **Save Action:** The "save" WebSocket message (e.g., `pluginID.settings.save`) **must** include this `_version` token in its JSON payload.
3.  **Broker Validation:** The **Central Broker** **must** check this token. If the backend's current version does not match the token from the user, the Broker **must reject** the save by sending a **Standardized Notification** (Section 4.2, `level: "error", code: "STALE_DATA"`) back over the WebSocket.

### 3.4 Data Transformation and Initial State

The Go backend (Google Style Guide) will produce JSON with `PascalCase` keys.

  * **Key Transformation:** The **Go backend (Central Broker) is responsible** for transforming the JSON keys from `PascalCase` to `camelCase` before transmission to the React frontend. This maintains the React frontend as a "dumb rendering layer."
  * **Initial State Retrieval:** Upon frontend connection, the UI **must** first use a **one-time HTTP/S GET request** to retrieve all static and current-state data (settings, layout, plugin status, etc.).
  * **State Re-Synchronization:** This HTTP/S GET request **must** also be re-triggered automatically by the frontend upon re-establishing a lost Control Plane WebSocket connection to prevent a stale UI state.

-----

## 4.0 API Versioning, Notifications, & Internationalization

To ensure long-term stability and maintainability, the following contracts are non-negotiable.

### 4.1 Versioning & Handshake

  * **Core Handshake:** The initial HTTP/S GET request (Section 3.4) **must** be used for an API version handshake. If the frontend and backend API versions are incompatible, the UI **must** display a "Software Mismatch" error and halt operation.
  * **Schema Versioning:** All JSON payloads for the **Dynamic Form Schema Contract** (Section 8.2.1) and **Widget Data** (Section 8.5.2) **must** include a `schemaVersion` key (e.g., `"schemaVersion": "1.2"`).
  * **Broker Validation:** The **Central Broker** is responsible for validating schema versions. If a plugin provides a schema version the Broker knows the frontend cannot support, the Broker **must not** send that schema. Instead, it must send a **Standardized Notification** (Section 4.2) to the frontend.

### 4.2 Standardized Notification Schema

All informational, success, and error messages sent from the Broker to the frontend **must** use the following standardized JSON schema. This allows the frontend to use a single, "dumb" notification/toast component.

```json
{
  "level": "success | info | warning | error",
  "source": "pluginID or broker",
  "code": "VALIDATION_FAILED or STALE_DATA or SAVE_SUCCESS",
  "messageKey": "notify.save.success",
  "messageContext": { "item": "Settings" }
}
```

  * **Broker Mandate:** The Broker **must** send a `level: "success"` notification upon the successful completion of actions like saving settings.

### 4.3 Internationalization (i18n)

To support localization, all user-facing strings **must** be replaced with **translation keys** across all API contracts.

  * **Contract Mandate:** Static strings like `label` or `pluginName` are forbidden. Contracts must use keys (e.g., `labelKey: "plugin.icecast.form.server_address"`).
  * **Broker Responsibility:** The **Central Broker** is responsible for collecting translation files (e.g., `en.json`, `es.json`) from all installed plugins and aggregating them.
  * **Frontend Responsibility:** The frontend is responsible for loading the appropriate, aggregated language file from the Broker (via the initial HTTP GET) and mapping all keys to the correct translated text.

-----

## 5.0 State & Data Flow (WebSocket Pub/Sub)

The core of the UI's real-time functionality is a **Pub/Sub** model over the single **Control Plane** WebSocket connection.

### 5.1 One-to-Many (Command Broadcast): "Start All"

  * **Action:** The React UI sends *one* message to the Broker (e.g., `system.command.startAll`).
  * **Broker:** The Broker broadcasts this event to *all* running modules.
  * **Modules:** Modules that have subscribed to this event (e.g., "Broadcast" modules) will execute their "start" logic. Other modules will ignore it.

### 5.2 Many-to-One (Status Aggregation): "LIVE" Indicator

This is a critical, backend-driven optimization to keep the frontend simple.

  * **Modules (The "Many"):** All "Broadcast" modules (Icecast, SRT, etc.) publish their individual status (e.g., `"active"`, `"inactive"`) to a shared aggregate topic (e.g., `broadcast.status.aggregate`).
  * **Broker (The "Aggregator"):** The Broker is the *only* subscriber to this topic. It maintains an internal state (a set) of all active broadcast modules.
  * **Frontend (The "One"):** The Broker publishes the *final, aggregated* boolean state (`true` if the set is not empty, `false` if it is) to a *single* topic (e.g., `system.live.status`). The React "LIVE" indicator component subscribes *only* to this one topic.

-----

## 6.0 Component & Module Architecture

### 6.1 Terminology

  * **Module:** The high-level functional unit (e.g., the Icecast plugin).
  * **Component:** The fundamental React building block.
  * **Widget Component:** A specific Component on the dashboard that represents a Module.

### 6.2 Module Definition (Manifest)

For the frontend to render a module, the module **must** provide the following metadata to the Broker:

  * **`pluginID` (string):** The unique, immutable ID (e.g., `com.icecast.streamer`). This **must** be used as the prefix for all its Pub/Sub topics.
  * **`pluginNameKey` (string):** The translation key for the human-readable display name (e.g., `plugin.icecast.name`).
  * **`pluginType` (string):** The category for UI grouping and logic.
      * `Broadcast`: (e.g., Icecast, SRT). Participates in "LIVE" aggregation.
      * `Transport`: (e.g., JACK, AES67, PipeWire). Does not participate in "LIVE" aggregation.
      * `Utility`: (e.g., RBAC, System Monitor).
  * **`minLevel` (string):** The minimum complexity level (`Standard`, `Advanced`, `Professional`, `Developer`) required for this module to be visible.
  * **`requiresInitialConfig` (boolean, optional):** If `true`, this flag triggers the "First Run" UI behavior (Section 6.4) when the module is first enabled.

### 6.3 Module State & Resource Conservation

The Broker **must** maintain a persistent **Plugin Status Table** containing `pluginID` and `isEnabled` (boolean).

  * **Resource Conservation:** If `isEnabled` is `false`, the Broker **must not** launch the `go-plugin` process for that module. This ensures disabled modules consume **zero CPU or RAM**.
  * **UI Behavior:** The frontend will query this table to populate the "Enable/Disable Modules" settings page.

### 6.4 Module "First Run" UI Behavior

When a user enables a module with the `requiresInitialConfig: true` flag, the Broker **must** instruct the frontend to act based on the user's active **Complexity Level**:

  * **Standard:** The frontend **must forcibly redirect** the user to the module's settings page (Section 8.2).
  * **Advanced:** The frontend **must open a modal dialog** confirming the plugin is enabled and prompting the user to "Configure Now" (redirects) or "Configure Later" (dismisses).
  * **Professional & Developer:** No action is taken. The system trusts the user to complete the configuration manually.

### 6.5 Plugin Lifecycle Management

To support a modular ecosystem, the Broker and UI must support the full plugin lifecycle.

  * **Broker API:** The **Central Broker** must provide data (via the initial HTTP GET) that lists all plugins (installed and available) and their `version`, `latest_available_version`, and `compatibility_status`.
  * **Frontend UI:** The frontend must provide a "Module Management" settings page that consumes this data. This UI will allow users to install, update, and uninstall modules (via WebSocket commands). It **must** clearly display compatibility warnings or available updates to the user.

-----

## 7.0 UI/UX Principles & Frontend Mandates

### 7.1 Hierarchical Complexity Levels

The entire UI operates on four, system-wide, hierarchical levels:

1.  **Standard**
2.  **Advanced** (includes Standard)
3.  **Professional** (includes Advanced)
4.  **Developer** (includes Professional)

### 7.2 Global State & Dynamic Configuration

  * **Frontend:** The active complexity level is stored in a **Global React Context** (`useComplexityLevel()`).
  * **Backend:** The Go Backend is responsible for filtering all configuration JSON based on the active level.
  * **Data Flow:** The frontend receives a simple, pre-filtered JSON payload (via the initial HTTP GET) containing only the settings and widgets relevant to the user's active level.
  * **Visibility:** If a user is at the "Standard" level, the UI does not load or render "Professional" level modules or settings. They are not grayed out; they are **not present**.

### 7.3 Accessibility & Navigability Mandates

To ensure the UI is robust, professional, and operable by power users and assistive technologies (like screen readers), the frontend implementation **must** adhere to the following:

  * **1. Full Keyboard Navigability:** All interactive elements in the application (form controls, widgets, buttons, tabs) **must** be fully focusable and operable using only a keyboard (Tab, Shift+Tab, Enter, Space, Arrow Keys). This is a non-negotiable requirement.
  * **2. Semantic HTML & ARIA:** The "dumb" frontend renderer **must** be responsible for:
      * **Generating Correct Associations:** All form inputs must have a unique `id` attribute, and their corresponding `<label>` must use the `htmlFor` attribute.
      * **Using Correct ARIA Roles:** All custom, non-native controls (like a `"toggle"`) **must** be built with the proper ARIA attributes (e.g., `role="switch"`, `aria-checked="true"`) to describe their state and function.

### 7.4 Client-Side State Persistence

To provide a seamless user experience, non-critical, transient UI state **must** be persisted in the browser's **`localStorage`**.

  * **Scope:** This includes, but is not limited to, the user's selected **Complexity Level** (7.1), the "open/closed" state of sidebars, and the last-active tab on any given settings page.
  * **Benefit:** This ensures the UI "remembers" the user's context on a page refresh without requiring any backend state management.

### 7.5 Contextual Help System

To onboard users and explain complex settings without cluttering the UI, a contextual help system is required.

  * **Schema Enhancement:** The **Dynamic Form Schema Contract** (Section 8.2.1) **must** support an optional `helpKey` (an i18n key).
  * **Frontend Implementation:** The "dumb" form renderer **must** be responsible for adding a small `(?)` help icon next to any label that provides a `helpKey`.
  * **UI Behavior:** Clicking this icon will trigger a popover or slide-out panel that displays the translated help text for that specific field.

### 7.6 Frontend Logic Centralization

To enforce the "dumb component" philosophy and ensure maintainability, all cross-cutting frontend logic **must** be centralized.

  * **Mandate:** All complex, non-visual logic (e.g., WebSocket connection management, i18n mapping, `localStorage` persistence, global state) **must** be abstracted into a **"Core UI Service Layer"** (e.g., global React Contexts and custom Hooks).
  * **Benefit:** Components will remain "dumb" and declarative. (e.g., a component will call `const { isOnline } = useConnection()` instead of managing the WebSocket lifecycle itself).

-----

## 8.0 Global UI Components & Core Features

### 8.1 Global Connection Status

The frontend **must** provide clear, persistent feedback about the state of the **Control Plane** WebSocket connection.

  * **Persistent Header Indicator:** The main UI header (top right) **must** display a status indicator:
      * **Green Indicator + "Online" Text:** When the WebSocket connection is active.
      * **Red Indicator + "Offline" Text:** When the WebSocket connection is lost.
  * **Modal Overlay:** If the connection is lost, the frontend **must** immediately display a **modal overlay** (e.g., "Connection Lost. Reconnecting...") that **disables all forms and controls**. This prevents user actions that are guaranteed to fail. The overlay is dismissed once the connection and state re-sync (Section 3.4) are successful.

### 8.2 Settings Pages (Schema-Driven Forms)

  * **Dynamic Rendering:** All module settings pages are dynamically rendered from a JSON schema provided by the module (and pre-filtered by the Broker, received in the initial HTTP GET).
  * **API Optimization:** The JSON schema format **must** be defined to match the API of the chosen React form library (**React Hook Form**). This **eliminates the need for a data-mapping layer** on the frontend.
  * **Saving Data:** Saving settings **must** be performed by sending a targeted Pub/Sub message (e.g., `pluginID.settings.save`) over the **Control Plane** WebSocket, containing the form data and concurrency token (Section 3.3.1).
  * **Save Contract:** All settings save messages sent by the frontend **must** be treated as a **partial update (a JSON merge)** by the Central Broker. The Broker is responsible for merging the submitted fields into the existing configuration, not overwriting the entire object. This protects higher-level settings from being destroyed by a lower-level user.
  * **Performance (Code Splitting):** The JavaScript and CSS for each module's settings page **must** be code-split (e.g., using `React.lazy()`) and loaded on-demand when the user navigates to that page. This keeps the initial application bundle small and fast.

#### 8.2.1 Dynamic Form Schema Contract

The settings page JSON will be a list of field definitions (an array of objects), each representing a single input control. This is the prescriptive vocabulary the backend must use.

| Control Type (`type`) | Description | Example Configurable Element (Non-Default) |
| :--- | :--- | :--- |
| **`text`** | Single-line text input. | `password` (Hides input, adds toggle visibility.) |
| **`number`** | Numeric input. | `min` / `max` (Enforces numerical boundaries.) |
| **`toggle`** | Simple `true`/`false` switch. | `descriptionKey` (A translation key for a helper string.) |
| **`select`** | Dropdown for predefined options. | `isMulti` (Allows selecting multiple options.) |
| **`radio`** | Select one option from a list. | `sortOrder` (Instructions for the frontend to sort the options.) |
| **`textarea`** | Multi-line text input. | `monospace` (Renders content using a fixed-width font.) |
| **`file`** | File upload input. | `accepts` (A string of file types, e.g., `"audio/*, .mp3"`.) |
| **`date` / `time`** | Date and/or time input. | `maxDate` / `minDate` (Sets the boundary for user selection.) |
| **`color`** | Color picker input. | `defaultColor` (A HEX string to set the initial color value.) |
| **`range`** | Slider input. | `step` (The increment value for the range slider.) |
| **`array-text`** | Allows the user to dynamically add/remove a list of text inputs. | `maxItems` (Maximum number of inputs the user can add.) |
| **`key-value`** | Allows the user to define paired key/value inputs. | `keyType` (Sets validation for the key field, e.g., `"regex"`.) |
| **`custom-component`**| Reference to a specific, pre-built React component to render. | **Security Constraint:** The list of valid `componentName` references must be **pre-vetted and immutable** by the core frontend application to prevent injection attacks. |
| **`nested-group`** | Renders a subsection of controls within the current form. | `fields` (An array of field definitions to render.) |
| **`divider`** | A visual separator. | `labelKey` (A translation key for a string to display in the divider.) |
| **`info-box`** | A box for displaying static contextual information. | `style` (e.g., `"warning"`, `"success"`, `"error"` for color/icon styling.) |
| **`helpKey` (Optional)**| (On any control) A translation key for contextual help. | `helpKey: "plugin.icecast.help.mountpoint"` (Renders a `(?)` icon) |

*Note: All user-facing strings (e.g., `label`, `placeholder`, `description`) must be replaced with their i18n equivalents (e.g., `labelKey`, `placeholderKey`, `descriptionKey`).*

### 8.3 Master Controls Logic (Start All / Stop All)

The Central Broker will manage the complete state of both the "Start All" and "Stop All" buttons, including visibility and any visual cues like muting, based on the collective state of all currently enabled Broadcast modules.

  * **Conditional Visibility:** The buttons are **only rendered** if the Broker reports that at least one `Broadcast` module is currently `isEnabled`. If zero `Broadcast` modules are enabled, the buttons are **hidden**.
  * **State Management Rules:**
    1.  **All enabled modules are broadcasting:** The **Stop All** button will be displayed in its **normal (illuminated) state**, while the **Start All** button will be **muted**.
    2.  **All enabled modules are stopped:** The **Start All** button will be displayed in its **normal (illuminated) state**, while the **Stop All** button will be **muted**.
    3.  **Mixed state (some broadcasting, some stopped):** Both buttons will be displayed in their **normal (illuminated) state**.
  * **Action:** Triggers the `system.command.startAll` (one-to-many) Pub/Sub event.

### 8.4 "LIVE" Indicator

  * **Implementation:** A single React component that subscribes *only* to the `system.live.status` (many-to-one) Pub/Sub topic.

### 8.5 Dynamic Dashboard Layout Editor

This feature provides users (at a designated complexity level, e.g., **Advanced** and above) the ability to fully customize their primary dashboard view.

  * **Implementation:** Uses a **React Grid Layout** library (e.g., `react-grid-layout`) to manage draggable and resizable `Widget Components`.
  * **Performance (Code Splitting):** The JavaScript and CSS for each `Widget Component` **must** be code-split (e.g., using `React.lazy()`) and loaded on-demand.
  * **Configuration Storage:** All layout definitions (Widget ID, position `x, y`, width `w`, height `h`) are serialized into a **single JSON object** and persisted on the **Go Backend**.
  * **User Profiles:** The backend supports storing multiple, named layout configurations per user, allowing the user to switch between a **Monitoring Layout** and a **Control Layout**.
  * **Default Layout:** A core, immutable **Default Layout** is stored and loaded if a user has not configured a custom layout.
  * **Widget Component Wrapper & Boundary Enforcement:** Every Widget Component **must** be rendered inside a standard, size-responsive **Wrapper Component** that enforces the correct styling (using Design Tokens) and provides the drag handles. The wrapper component **must enforce strict boundaries** to prevent any plugin UI from rendering outside the defined widget space.

#### 8.5.1 Layout Management Interface

To manage multiple layout profiles (User Profiles), a hybrid UI is mandated:

  * **Main UI Header:** A **dropdown menu** must be present in the main UI. Its sole function is to **load or switch** between the user's available, named layouts.
  * **Layout Editor ("Edit Mode"):** Controls for **"Save," "Save As...," "Rename," and "Delete"** **must only appear** within the layout editor context. This aligns with the "progressive complexity" philosophy.
  * **RBAC Governance:** The behavior of this interface **must** be controlled by the RBAC module (Section 9.2).

#### 8.5.2 Widget Data Flow (Real-time)

  * **Control Plane Data:** The Central Broker utilizes the existing **Control Plane WebSocket** for targeted, low-frequency updates (e.g., listener count, on/off state).
  * **Subscription:** When a `Widget Component` loads, it immediately subscribes to a topic using its **unique `pluginID` as the prefix** (e.g., `com.icecast.streamer.widget.status`).
  * **Action Mechanism:** Widget control actions (e.g., clicking a "Mute" toggle) **must** communicate back to the Central Broker using a specific, targeted **Pub/Sub message** (e.g., `pluginID.command.action`) over the **Control Plane** socket.
  * **Performance (Throttling):** The frontend's subscription hook for the **Control Plane** **must** reserve the right to **throttle or debounce** incoming messages to protect core UI stability.
  * **Visualization Plane Data:** Widgets requiring high-frequency visualization data (e.g., level meters) **must** use the optional **Data Visualization Socket** (Section 3.3). They **must not** send this data over the main Control Plane socket.

#### 8.5.3 Widget Keyboard Navigation (Roving `tabindex`)

To support full keyboard operability for the dynamic grid (as mandated in Section 7.3), the **Widget Component Wrapper** must implement a "roving `tabindex`" pattern.

  * **Grid Navigation (Level 1):** The user will navigate between widget wrappers using the **Arrow Keys**. The wrapper component is responsible for managing its `tabindex` (either `0` or `-1`) and highlighting the active focus.
  * **Widget-Internal Navigation (Level 2):** Pressing **`Enter`** or **`Tab`** on a focused wrapper will move focus to the first control *inside* the widget. The **`Tab`** key will then be "trapped" within the widget's controls. Pressing **`Escape`** will return focus to the widget wrapper.

#### 8.5.4 Orphaned Widget Handling

To prevent UI errors when a user-defined layout contains a widget for a disabled module, the following logic is mandated:

  * **In "View Mode" (Default):** When the frontend loads the dashboard, if a widget in the layout JSON belongs to a module where `isEnabled` is `false`, the frontend **must render `null`**. This creates a non-destructive "hole" in the grid. If the module is re-enabled, the widget will reappear.
  * **In "Edit Mode" (Layout Editor):** When the user enters the layout editor, the frontend **must render** a built-in **"Disabled Module Placeholder"** component in the widget's grid space. This gives the user a visual target to permanently delete the orphaned widget from their layout if they choose.

-----

## 9.0 Security

### 9.1 Sanitation (Prescriptive)

All user-generated input **must** be sanitized and validated on **both the frontend (client-side) and the backend (server-side)**. This "**Defense in Depth**" is a non-negotiable security requirement.

### 9.2 (Optional) Role-Based Access Control (RBAC)

  * The system will support an optional, internal **Utility Module** for RBAC.
  * If enabled, this module will manage user logins and permissions, protecting the UI from unauthorized access.
  * The RBAC module **must** also govern the behavior of the **Layout Management Interface (8.5.1)**. This includes, but is not limited to: (a) enabling full management for admin roles, (b) scoping layout management to a user's own profile, and (c) assigning specific, locked-down layouts to restricted roles.

-----

## 10.0 First-Time Setup (Configuration Wizard)

  * **Trigger:** The backend **must** provide a "one-shot" configuration flag (e.g., `isConfigured: false`).
  * **Action:** If this flag is `false`, the frontend **must** redirect the user to a **Configuration Wizard** upon login.
  * **Scope:** This wizard will guide the user through mandatory setup steps, such as setting the desired **Complexity Level** and (Optional) Configuring the **RBAC / User Login** module.
  * **Extensibility:** This core wizard is *not* designed to be extended by third-party plugins, as it handles fundamental system setup.