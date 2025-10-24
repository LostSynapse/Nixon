# Nixon Development Roadmap

This document outlines the development plan for the Nixon project, reflecting the current 'Nixon-testing' codebase and incorporating future goals.

## Phase 1: Core Architecture (Complete)

This phase establishes the foundational features of the Nixon appliance, based on the current refactored codebase.

* \[x\] **Backend:** Modular Go backend using the standard `net/http` library.
* \[x\] **Database:** Migrated to GORM for robust database interactions.
* \[x\] **Audio Pipeline:** Unified, single-process audio pipeline built using native `go-gst` bindings (not command-line).
* \[x\] **Audio Source:** Uses `pipewiresrc` as the master audio source, enabling native routing and eliminating hardware conflicts.
* \[x\] **Audio Core:** A `tee` element splits the master signal into dynamic branches for VAD, pre-roll, recording, and streaming.
* \[x\] **Frontend:** Component-based React single-page application using a custom `useNixonApi` hook for state management.
* \[x\] **Configuration:** Centralized `config.json` driving all backend logic.
* \[x\] **Real-time:** Event-driven WebSocket communication with the backend pushing state updates.
* \[x\] **Logo:** Converted to a re-colorable React component with a separate hardcoded-black favicon.

## Phase 2: Stability & Feature Completion (Current Priority)

This phase focuses on hardening the new architecture and completing the partially-implemented features from the refactor.

* \[ \] **(Highest Priority) Dummy Audio Source:** Implement a fallback to `audiotestsrc` within the GStreamer pipeline. If the configured `pipewiresrc` device fails to initialize at startup, the pipeline must failover to this test source. This prevents a backend crash and allows the user to access the UI to correct the audio configuration.
* \[ \] **Hardware Capability Detection:** Implement the backend API endpoint (`/api/capabilities`) that executes `arecord --dump-hw-params` and parses the output for supported sample rates and bit depths. The frontend "Audio" tab must be connected to this endpoint to dynamically populate its dropdowns.
* \[ \] **Channel Mapping:** Implement the GStreamer logic (`deinterleave`/`interleave`) to dynamically select the stereo pair specified by the `master_channels` array in the config.
* \[ \] **Disk Management (Auto-purge):** Implement a background task to automatically delete the oldest *unprotected* recordings when disk usage exceeds a configurable threshold.

## Phase 3: User Management & Security (RBAC)

This phase introduces a full Role-Based Access Control (RBAC) system.

* \[ \] **Authentication:** Implement a user login system with secure password hashing and session management.
* \[ \] **User Roles:**
    * **Admin:** Full control over all system settings, user management, and can view/manage all recordings.
    * **User:** Can start/stop streams and recordings, and can only view/manage their own recordings.
* \[ \] **Recording Ownership:** Update the database schema to associate each recording with a `user_id`. The API and frontend will be modified to enforce this ownership.
* \[ \] **Dynamic Talkback Ports:** Upon a successful login via the companion app, the backend will dynamically assign the user an available network port for their talkback audio stream (dependent on Phase 4).

## Phase 4: Networked Control & Routing

This phase implements the "routing matrix" concept for both audio and MIDI, using open standards as the primary goal.

* \[ \] **AES70 (OCA) Control Framework:**
    * \[ \] Implement an AES70 device controller in Go. This will become the primary framework for all control and communication, replacing the simple REST API for pipeline and state control.
    * \[ \] Expose all controllable parameters (routing, mixing, config) as AES70 objects using a temporary manufacturer ID.
* \[ \] **Audio Routing Matrix (Talkback):**
    * \[ \] Develop a "Talkback" mobile app (as a PWA or native app).
    * \[ \] Create a "Routing" tab in the Nixon UI to visualize and control the PipeWire/JACK graph.
    * \[ \] Use AES70 commands to create and sever connections (`pw-link`), allowing routing of any source (e.g., Talkback user) to any "Virtual Destination" (e.g., `pipewiresink` for headphones).
* \[ \] **MIDI Routing Matrix:**
    * \[ \] Add backend logic to detect and manage MIDI devices (USB, 5-pin DIN via GPIO, and `rtpmidi` network streams) through PipeWire.
    * \[ \] Expand the "Routing" tab to show MIDI sources and destinations.
    * \[ \] Use AES70 commands to manage the MIDI routing (e.g., `pw-link` for MIDI ports).
* \[ \] **Multi-Device Routing (Nixon-to-Nixon):**
    * \[ \] Implement mDNS for automatic discovery of other Nixon appliances.
    * \[ \] Use AES70 for inter-device control.
    * \[ \] Create a hybrid audio transport system that automatically uses **AVB/AES67** (see Phase 5) if a compatible network is detected, and falls back to dynamically created **SRT** streams on standard networks.

## Phase 5: Professional Integrations & Usability

This phase adds advanced features and integrations.

* \[ \] **Professional AoIP Integration (AVB/AES67):**
    * \[ \] With the `i226-v` hardware solution identified and the `pipewiresrc` foundation in place, this is a high-priority integration.
    * \[ \] Implement `avbsrc` / `avbsink` and AES67 support, allowing Nixon to act as a native, multi-channel node on a pro-audio network.
* \[ \] **Low-Latency P2P Remote Collaboration ("Jamming"):**
    * \[ \] Implement a **WebRTC** mode for direct internet-based collaboration, brokered by a signaling server (using `network_settings` from `config.json`).
* \[ \] **Web Player Enhancements:**
    * \[ \] **Waveform Display:** Integrate a library to render a visual waveform of recordings.
    * \[ \] **Tagging/Marking:** Allow users to add time-stamped markers to recordings.
* \[ \] **System Dashboard:** Add a status page showing CPU, memory, and PipeWire graph diagnostics.
* \[ \] **Automatic File Management (Post-Recording Hooks):** Add a feature to execute a user-defined script after a recording is finished for automatic backup, transcoding, or notifications.
* \[ \] **Elgato Stream Deck Integration:** Develop an official plugin for one-press hardware control.
