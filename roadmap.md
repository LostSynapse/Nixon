# Nixon Development Roadmap

This document outlines the development plan for the Nixon project, incorporating the current state of the codebase, findings from the recent audit, and the long-term goal of creating a managed appliance.

## Current State & Audit Findings

The existing `Nixon-testing` codebase provides a solid foundation but requires immediate attention in several areas to ensure stability, maintainability, and security. The `audit.md` report highlights the following critical issues:

*   **Lack of Structured Logging:** The current use of the standard `log` package produces inconsistent and difficult-to-parse logs, hindering effective debugging.
*   **Inconsistent Error Handling:** Error handling is not standardized, with a mix of `log.Fatal`, `panic`, and unhandled errors, leading to unpredictable application behavior.
*   **Missing API Authentication:** The WebSocket API at `/ws` is unprotected, allowing any user on the network to execute arbitrary commands, posing a significant security risk.
*   **Outdated Dependencies:** Several Go and JavaScript dependencies have known vulnerabilities that must be addressed.

These findings directly inform the first phase of our roadmap.

## Phase 1: Foundational Stability & Refactoring

This phase is focused on addressing the critical issues identified in the audit to create a stable, secure, and maintainable platform.

*   **[In Progress] Implement Structured Logging:**
    *   Replace all `log` package calls with a structured logger (`slog`).
    *   Establish consistent, machine-readable log formats (JSON).
*   **[To Do] Standardize Error Handling:**
    *   Refactor the codebase to use a consistent error handling strategy, returning errors to the caller for appropriate action.
    *   Eliminate all uses of `log.Fatal` and `panic` for recoverable errors.
*   **[To Do] Secure the WebSocket API:**
    *   Implement a token-based authentication mechanism for the `/ws` endpoint.
    *   The UI will need to be updated to handle the authentication flow.
*   **[To Do] Dependency Management:**
    *   Update all outdated Go and NPM packages to their latest secure versions using `go get` and `npm update`.
    *   Perform a vulnerability scan to ensure all known issues are resolved.
*   **[To Do] Configuration Management:**
    *   Centralize all configuration into the `config.json` file.
    *   Remove hardcoded values from the source code, using `viper.SetDefault()` in `internal/config/config.go` to provide initial values.

## Phase 2: Core Feature Enhancement & Plugin Architecture

Once the foundation is stable, this phase will focus on building out core features and a robust plugin system, including refactoring existing components into plugins.

*   **[To Do] AES70 Core Interfaces (Native Go Foundation):** Establish the foundational Go interfaces and architectural hooks within `internal/aes70` to enable future, full native Go AES70 protocol implementation. This focuses on design and readiness, not full protocol logic. This is a critical prerequisite for future multi-device control.
*   **[To Do] Recording Management:**
    *   Implement robust recording controls (start, stop, pause).
    *   Develop a user interface for listing, downloading, and deleting recordings.
*   **[To Do] Audio & MIDI Routing (via Jackwire2 Plugin):**
    *   Refactor PipeWire/JACK control into a `jackwire2` plugin.
    *   Create a "Routing" tab in the UI to visualize and manage audio/MIDI streams through this plugin.
    *   Allow users to connect and disconnect audio/MIDI ports graphically.
*   **[To Do] Plugin System & Core Streaming Refactor:**
    *   Design and implement a Go-based plugin architecture.
    *   Refactor existing Icecast and SRT streaming functionalities into plugins, validating the plugin API.

## Phase 3: Managed Appliance Upgrade System (APT-Centric)

This phase addresses the critical requirement for a managed appliance by building a reliable, in-place upgrade system, primarily leveraging `apt` for Debian-based systems, incorporating detailed requirements.

*   **[To Do] Go Migration Tooling & Atomic DB Control:** Implement robust database migration tooling (`golang-migrate/migrate`) with external read-write locks and atomic transactions.
*   **[To Do] Configuration Versioning & Merging:** Ensure `config.json` includes a `ConfigVersion` and use `imdario/mergo` for non-destructive configuration merges, prioritizing user values.
*   **[To Do] Companion Executable:** Develop a privileged Go binary (`nixon-upgrade-helper`) to orchestrate upgrade/downgrade actions, called by Debian package hooks.
*   **[To Do] Integrity Guard & APT Lock:** Implement a startup check for migration locks in `cmd/nixon/main.go` and ensure `apt-mark hold` prevents unattended upgrades.
*   **[To Do] Debian Packaging Integration:** Develop comprehensive `preinst`, `postinst`, `pre-remove`, `post-remove` scripts to manage the upgrade/downgrade process robustly, including logging and rollback.
*   **[To Do] Master-Slave Orchestration:** Implement logic for a Master unit to trigger `apt upgrade` on Slaves, manage their "Ready for Adoption" state, and re-adopt them via AES70.
*   **[To Do] Optional UI-Triggered Upgrade:** Provide a secure UI on the Master unit to initiate and monitor the upgrade of connected Slaves.

## Phase 4: Advanced Monitoring & Local Features

This phase focuses on adding detailed monitoring and local network discovery features that do not immediately depend on full AES70 control.

*   **[To Do] Multi-Device Discovery (mDNS only):**
    *   Implement mDNS for automatic discovery of other Nixon appliances on the local network.
    *   Create a UI component to display a list of discovered Nixon appliances. This task focuses solely on local network discovery via mDNS, without full AES70 control.
*   **[To Do] Advanced Monitoring:**
    *   Provide detailed system monitoring in the UI, including CPU, memory, and disk usage, as well as real-time audio metering.
*   **[To Do] Professional Audio Transport (AVB/AES67 - Single Device):**
    *   Implement AVB and AES67 support as plugins, leveraging PipeWire for single-device audio transport capabilities.

## Phase 5: Full AES70 Control and Multi-Device Networking

This final phase implements the comprehensive AES70 protocol in native Go and leverages it for multi-device control.

*   **[To Do] Full AES70 Protocol Implementation (Native Go):**
    *   Develop the comprehensive native Go library for AES70 based on provided reference documents and client tools, implementing full device discovery, state management, command/control. This critical feature will be implemented after all other local features and single-device networking are operational.
*   **[To Do] Multi-Device Control (Leveraging Full AES70):**
    *   Implement master-slave control and orchestration, allowing a primary Nixon unit to manage and synchronize other Nixon units via the fully implemented AES70 protocol.
