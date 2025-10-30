# 📋 Project Backlog: Complete 33-Item List (Last Known Good State)

## I. Foundational & Architectural Mandates (P1)
(11 Items)
These are non-negotiable prerequisites that must be addressed to resolve major security, integrity, and stability risks.

| Feature / Gap | Rationale | Priority |
| :--- | :--- | :--- |
| **Authentication UI & Flow** | Critical security gap. Defines the login/logout mechanism and token management. | **P1 (Critical Security)** |
| **Frontend Build & Deployment Strategy** | Defines how React assets are bundled (e.g., embedded) to ensure compliance with the Upgrade Document. | **P1 (Critical Architecture)** |
| **Code Review and Merge Policy** | Defines the process for quality assurance and prevents accidental merging of non-compliant code. | **P1 (Critical Process)** |
| **API Rate Limiting Policy** | Security mandate. Prevents a rogue plugin or user from flooding the Broker via WebSocket. | **P1 (Critical Security)** |
| **Secure IPC Protocol** | Defines the robust communication channel between the Central Broker and the privileged `nixon-migrator`. | **P1 (Critical Integrity)** |
| **Privilege Management** | Mandates a secondary security audit and defined scope for the `nixon-migrator` codebase. | **P1 (Critical Security)** |
| **UI Integration (System Lockout)** | Defines the "System Maintenance" screen displayed when the `nixon-migrator` applies its lock file. | **P1 (Critical Integrity)** |
| **Define RPC Contract (Migrator)** | Creates the slimmed-down RPC interface the Broker uses to monitor the `nixon-migrator`'s status. | **P1 (Critical Architecture)** |
| **localStorage Schema Versioning** | Future-proofing mandate. Prevents the frontend from crashing on software updates with stale local data. | **P1 (Critical Future-Proofing)** |
| **License Compliance Strategy** | Defines the legal process for auditing and complying with all third-party software licenses. | **P1 (Critical Process)** |
| **Global Path Configuration Mandate** | (Integrated) Defines architecture for managing all configurable file paths (Recording, Logs, Config) via a single Path Resolver Service (PRS). | **P1 (Critical Integrity)** |

---

## II. Core Feature Development (P2)
(12 Items)
These are essential, missing, user-facing features required for the appliance to function as a complete V1 product.

| Feature / Gap | Rationale | Priority |
| :--- | :--- | :--- |
| **Recording Management Page** | The UI for listing, downloading, starting/stopping, and deleting recordings. | **P2 (Immediate Core)** |
| **User Management Page (RBAC)** | The administrative UI for adding/deleting users, managing roles, and changing Feature Levels post-setup. | **P2 (Core System)** |
| **System Log Viewer** | The UI for tailing and viewing structured logs; essential for debugging the headless appliance. | **P2 (Core System)** |
| **Network Configuration UI** | The settings page for managing IP, subnet, gateway, and DHCP settings. | **P2 (Core Appliance)** |
| **System Storage & Log Management** | UI for viewing internal disk usage, log rotation, and cleaning caches. | **P2 (Core Appliance)** |
| **USB Storage & Drive Management** | UI for mounting, unmounting, and formatting external USB drives for recording. | **P2 (Core Appliance)** |
| **System Time (NTP) UI** | The settings page for managing time and Network Time Protocol (NTP) servers. | **P2 (Core Appliance)** |
| **Shutdown & Reboot UI** | Essential safety buttons to gracefully power down or restart the appliance. | **P2 (Core Appliance)** |
| **Factory Reset UI Trigger** | The "Danger Zone" button and confirmation modals to safely initiate a full system reset. | **P2 (Core Appliance)** |
| **Core Appliance Upgrade UI** | The UI for triggering and monitoring core system updates (e.g., `apt` integration). | **P2 (Core System)** |
| **Routing Page UI** | The graphical page for managing audio/MIDI port connections (the non-`jackwire2` core routing). | **P2 (Core Feature)** |
| **Schema Enhancement (Read-Only Fields)** | Adding types like read-only-text and progress-bar to the form schema to build dynamic status pages. | **P2 (Tooling)** |

---

## III. V2+ / Advanced Enhancements (P3)
(10 Items)
These items are advanced features that improve workflow or long-term product longevity and are scheduled for V2 or later.

| Feature / Gap | Rationale | Priority |
| :--- | :--- | :--- |
| **Internal Digital Mixer & DSP** | The top-level feature for creative audio control. | **P3 (Creative Core)** |
| **Remote Audio Monitoring ("Listen-In")** | Professional feature for checking audio quality remotely via a WebRTC/WebSocket stream. | **P3 (Creative Core)** |
| **System-Wide "Snapshots" (Preset Management)** | UI for saving/recalling the entire system state (settings, routes, layout). | **P3 (UX/Workflow)** |
| **Real-Time Collaboration & Presence** | UI to show which users are currently editing a page. | **P3 (UX/Collaboration)** |
| **Full MIDI Feature Definition** | Fleshing out the MIDI feature (device manager, routing, monitor) into a full platform. | **P3 (Feature Depth)** |
| **Optimistic UI Updates** | Design policy to make the UI feel instantaneous. | **P3 (Polish)** |
| **Client-Side Validation** | Policy to provide immediate error feedback before sending data to the server. | **P3 (Polish)** |
| **Localization/Translation Strategy** | The UI to select and manage the display language. | **P3 (Global)** |
| **General "About" Page** | A simple page for displaying version, serial, and legal information. | **P3 (UX)** |
| **Server-Side User Preferences** | Persistently saving user settings (theme, default feature level) on the backend. | **P3 (UX)** |
