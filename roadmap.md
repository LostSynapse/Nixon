# Nixon Development Roadmap

This document outlines the development plan for the Nixon project, combining initial goals with features from various design notes.

## Phase 1: Core Functionality (Complete)

This phase establishes the foundational features of the Nixon appliance.

* \[x] **Streaming Engine:** Low-latency SRT and Icecast streaming.
* \[x] **Recording Engine:** Manual start/stop/split recording.
* \[x] **Backend:** Modular Go backend with a REST API, WebSocket for real-time updates, and SQLite for metadata.
* \[x] **Frontend:** Component-based React single-page application with full recording management.
* \[x] **Configuration:** External `config.json` with a settings UI.

## Phase 2: Core Audio Engine Refactor & First-Boot Experience (In Progress)

This is a major architectural push to create a stable, professional-grade audio foundation, resolve core bugs, and deliver a polished out-of-the-box user experience.

* \[ ] **First-Boot Setup Wizard:**
    * \[ ] Implement a `systeminitialize` state machine (`3`: OEM boot, `2`: Factory Reset, `1`: User Setup, `0`: Normal).
    * \[ ] On state `3`, auto-generate a unique hostname (e.g., `nixon-a1b2c3d4`), store it in a persistent `nixonhost` config file, set the system hostname, set state to `2`, and reboot.
    * \[ ] On state `2`, purge all recordings and databases, restore factory default `config.json`, set state to `1`, and restart the application service.
    * \[ ] On state `1`, serve a setup wizard UI that guides the user through initial configuration.
    * \[ ] The wizard will ask the user to select an expertise level (Standard, Power User, Professional) to set the UI mode.
    * \[ ] The wizard will guide the user through setting the hostname, creating an optional admin account, and selecting the primary audio device.
* \[ ] **JACK/PipeWire Integration:**
    * \[ ] Integrate the JACK Audio Connection Kit as the core audio backend. PipeWire's compatibility layer (`pipewire-jack`) will be the recommended implementation.
    * \[ ] The `systemd` service will be updated to depend on and start after the `pipewire.service`.
* \[ ] **Unified GStreamer Pipeline (JACK-based):**
    * \[ ] Rewrite the entire audio pathway to use a single, master GStreamer pipeline with `jackaudiosrc` as the input, resolving all hardware access conflicts and fixing the VAD crash bug.
    * \[ ] Use a `tee` element to split the audio to branches for VAD, a pre-roll buffer, recording, and streaming.
* \[ ] **Implement Advanced Audio & VAD Features:**
    * \[ ] Implement automatic hardware capability detection (`arecord --dump-hw-params`) for sample rates and bit depths.
    * \[ ] Implement a configurable pre-roll buffer.
    * \[ ] Implement "Smart Split" with silence detection.
    * \[ ] Implement the foundation for a channel routing matrix by using `deinterleave`/`interleave` on the multi-channel JACK source.

## Phase 3: User Management & Security (RBAC)

This phase introduces a full Role-Based Access Control (RBAC) system.

* \[ ] **Authentication:** Implement a user login system with secure password hashing and session management. A persistent warning will be displayed if no admin account is created.
* \[ ] **Role Matrix:**
    * \[ ] Create a "Users & Roles" section in the settings for administrators.
    * \[ ] Implement a granular permission system and a UI matrix for creating and editing roles.
    * \[ ] Ship with sensible default roles (e.g., Administrator, Producer, Operator).
* \[ ] **Recording Ownership:** Update the database to associate recordings with a `user_id` and enforce ownership via the API.
* \[ ] **Dynamic Talkback Ports:** Upon a successful login via the companion app, the backend will dynamically assign the user an available network port for their talkback audio stream.

## Phase 4: Networked Audio & Collaboration

This phase expands Nixon into a fully networked audio tool.

* \[ ] **Companion "Talkback" Mobile App & Routing Matrix:**
    * \[ ] Develop a simple mobile app for remote control (Start/Stop/Split) and a push-to-talk microphone.
    * \[ ] Create a dedicated "Routing" tab on the main UI for the audio matrix.
    * \[ ] Implement a UI for creating "Virtual Destinations" (headphone mixes) and routing any source (local input, talkback user, network stream) to any destination.
* \[ ] **Multi-Device Management & Routing:**
    * \[ ] Implement mDNS for discovery of all Nixon appliances.
    * \[ ] Create a "Devices" UI for adopting/managing a fleet of Nixon units from a single controller, including promoting any member to controller.
    * \[ ] Implement a "master redirect" for the web UI of managed devices.
    * \[ ] Implement remote file management and aggregated disk usage monitoring.
    * \[ ] Create an abstracted transport layer that intelligently chooses the best protocol (PipeWire native, NetJack2, or SRT fallback).
* \[ ] **Low-Latency P2P Remote Collaboration ("Jamming"):**
    * \[ ] Implement a new mode using **WebRTC** for direct, low-latency, bidirectional audio link with another Nixon appliance over the internet.
    * \[ ] The backend will include a simple "Signaling Server" to broker connections.
    * \[ ] The UI will include a "Network" tab for configuring Signaling, STUN, and optional TURN servers.

## Phase 5: Professional Integrations & Usability (Future)

This phase adds advanced features and integrations.

* \[ ] **Web Player Enhancements:** Waveform display, time-stamped markers.
* \[ ] **System Dashboard:** A status page showing CPU, memory, and pipeline diagnostics, including an indicator for whether an RT Kernel is detected.
* \[ ] **Automatic File Management (Post-Recording Hooks).**
* \[ ] **Elgato Stream Deck Integration.**
* \[ ] **Professional AoIP Integration (Research):** Investigate and implement support for protocols like **AES67**, **Dante**, and **AVB**, leveraging the now-available I226-V NIC hardware and the JACK/PipeWire foundation.
