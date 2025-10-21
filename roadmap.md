# Nixon Development Roadmap

This document outlines the development plan for the Nixon project, combining initial goals with features from various design notes.

## Phase 1: Core Functionality (Complete)

This phase establishes the foundational features of the Nixon appliance.

* \[x\] **Streaming Engine:** Low-latency SRT and Icecast streaming via GStreamer.

* \[x\] **Recording Engine:** Manual start/stop/split recording to FLAC format.

* \[x\] **Backend:** Modular Go backend with a REST API, WebSocket for real-time updates, and SQLite for metadata.

* \[x\] **Frontend:** Component-based React single-page application with full recording management and real-time status indicators.

* \[x\] **Configuration:** External `config.json` with a settings UI for System, Audio, and Icecast configuration. Dynamic audio hardware detection.

## Phase 2: Intelligent Recording & Automation (In Progress)

This phase focuses on making the recording process smarter and more automated.

* \[x\] **Voice Activity Detection (VAD):** Implement GStreamer VAD pipeline for audio monitoring, enabling auto-start/stop of recordings. UI toggle for this feature is complete.

* \[ \] **Smart Split:** Implement backend logic for the "Smart Split" feature. When enabled, the VAD pipeline will automatically trigger a recording split after a user-configurable period of silence.

* \[ \] **Pre-roll Buffer:** Implement an in-memory ring buffer (`queue` element in GStreamer) that constantly captures the last 15-30 seconds of audio. When a recording starts (manually or via VAD), this buffer will be prepended to the file.

* \[ \] **Disk Management (Auto-purge):** If storage exceeds a configurable threshold (e.g., 97%), automatically delete the oldest *unprotected* recordings to free up space.

## Phase 3: User Management & Security (RBAC)

This phase introduces a full Role-Based Access Control (RBAC) system.

* \[ \] **Authentication:** Implement a user login system with secure password hashing and session management.

* \[ \] **User Roles:**

  * **Admin:** Full control over all system settings, user management, and can view/manage all recordings.

  * **User:** Can start/stop streams and recordings, and can only view/manage their own recordings. Cannot change critical system settings.

* \[ \] **Recording Ownership:** Update the database schema to associate each recording with a `user_id`. The API and frontend will be modified to enforce this ownership.

* \[ \] **Dynamic Talkback Ports:** Upon a successful login via the companion app, the backend will dynamically assign the user an available network port for their talkback audio stream.

## Phase 4: Networked Audio & Collaboration

This phase expands Nixon into a networked audio tool.

* \[ \] **Companion "Talkback" Mobile App & Routing Matrix:**

  * Develop a simple mobile app for remote control (Start/Stop/Split) and a push-to-talk microphone.

  * Create a "Talkback" tab in the settings UI to define "Virtual Destinations" (e.g., "Host Headphones") mapped to physical sound card outputs.

  * Implement a routing matrix in the UI to assign which talkback users can speak to which destinations.

* \[ \] **Multi-Device Routing (Nixon-to-Nixon):**

  * Implement mDNS for automatic discovery of other Nixon appliances on the local network.

  * Create a hybrid transport system that defaults to dynamically created **SRT** streams for inter-device audio routing.

* \[ \] **Low-Latency P2P Remote Collaboration ("Jamming"):**

  * Implement a new mode using **WebRTC** to establish a direct, low-latency, bidirectional audio link with another Nixon appliance over the internet.

  * The backend will include a simple "Signaling Server" to broker the initial P2P connection via a shared "Session Code".

## Phase 5: Professional Integrations & Usability (Future)

This phase adds advanced features and integrations.

* \[ \] **Advanced Audio Configuration:**

  * **Channel Mapping:** Implement backend logic for selecting which input channels are used for the master stereo pair.

  * **Bit Depth & Sample Rate:** Implement backend logic to make the "Bit Depth" and "Sample Rate" options fully functional, with the GStreamer pipeline dynamically adapting.

* \[ \] **Web Player Enhancements:**

  * **Waveform Display:** Integrate a library to render a visual waveform of recordings.

  * **Tagging/Marking:** Allow users to add time-stamped markers to recordings.

* \[ \] **System Dashboard:** Add a status page showing CPU usage, memory usage, and GStreamer pipeline diagnostics.

* \[ \] **Automatic File Management (Post-Recording Hooks):** Add a feature to execute a user-defined script after a recording is finished for automatic backup, transcoding, or notifications.

* \[ \] **Elgato Stream Deck Integration:** Develop an official plugin for one-press hardware control.

* \[ \] **Professional AVB Integration (Research):** De-prioritized due to significant hardware and Linux driver support challenges on SBCs. This remains a long-term goal pending hardware/software ecosystem maturity. The unified pipeline should, however, be architected in a way that adding
