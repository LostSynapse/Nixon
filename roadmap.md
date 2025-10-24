# Nixon Development Roadmap

This document outlines the development plan for the Nixon project, combining initial goals with features from the "Nixon-Record: Software Design" note.

## Phase 1: Core Functionality (Complete)

This phase establishes the foundational features of the Nixon appliance.

* **Streaming Engine:**
    * \[x\] Low-latency SRT streaming via GStreamer.
    * \[x\] Icecast streaming via GStreamer.
* **Recording Engine:**
    * \[x\] Manual start/stop/split recording to FLAC format.
* **Backend:**
    * \[x\] Go-based backend server with modular structure.
    * \[x\] REST API for controlling streams and recordings.
    * \[x\] WebSocket for real-time UI updates.
    * \[x\] SQLite database for metadata storage.
* **Frontend:**
    * \[x\] React-based single-page application with component structure.
    * \[x\] Responsive UI for desktop and mobile.
    * \[x\] Real-time status indicators for all services.
    * \[x\] Full recording management (play, download, edit, protect, delete).
    * \[x\] Dynamic disk usage gauge.
* **Configuration:**
    * \[x\] External `config.json` for all major settings.
    * \[x\] Settings UI with tabs for System, Audio, and Icecast.
    * \[x\] Dynamic detection of audio hardware.
    * \[x\] UI controls for enabling/disabling stream widgets.

## Phase 2: Intelligent Recording & Automation

This phase focuses on making the recording process smarter and more automated.

* **Voice Activity Detection (VAD):**
    * \[x\] Implement a GStreamer VAD pipeline to monitor audio activity.
    * \[x\] **Auto-record:** Automatically start recording when audio is detected.
    * \[x\] **Auto-stop:** Automatically stop recording after a configurable period of silence.
    * \[x\] UI toggle for enabling/disabling auto-record mode.
* **Disk Management:**
    * \[ \] **Auto-purge:** If storage exceeds a configurable threshold (e.g., 97%), automatically delete the oldest unprotected recordings.
* **Pre-roll Buffer:**
    * \[ \] Implement an in-memory ring buffer that constantly captures the last 1-2 minutes of audio.
    * \[ \] When a recording starts (manually or via VAD), prepend this buffer to the file to "capture the magic" that just happened.

## Phase 3: User Management & Security

This phase introduces multi-user capabilities and secures the application.

* **Authentication:**
    * \[ \] Implement a user login system (e.g., username/password).
    * \[ \] Store hashed passwords securely.
* **Permissions:**
    * \[ \] **Admin Role:** Can change system settings, manage all recordings, and manage users.
    * \[ \] **User Role:** Can start/stop streams and recordings, manage their own recordings, but cannot change critical settings or delete others' recordings.
* **Ownership:**
    * \[ \] Associate recordings with the user who created them.
    * \[ \] Modify the recordings list to show only the logged-in user's files (or all files for an admin).

## Phase 4: Advanced Features & Usability

This phase adds advanced audio control and enhances the user experience.

* **Advanced Audio Configuration:**
    * \[ \] **Channel Mapping:** In the Audio Settings tab, implement backend logic to allow users to select which input channels from the audio device are used for the stereo stream (e.g., use inputs 3/4 instead of 1/2).
    * \[ \] **Bit Depth & Sample Rate:** Implement backend logic and GStreamer pipeline modifications to support the "Bit Depth" and "Channels" options that are currently placeholders in the UI.
* **Web Player Enhancements:**
    * \[ \] **Waveform Display:** Integrate a library to render a visual waveform of recordings for easier scrubbing and navigation.
    * \[ \] **Tagging/Marking:** Allow users to add time-stamped markers or tags to recordings during playback.
* **System Dashboard:**
    * \[ \] Add a system status page/modal showing CPU usage, memory usage, and GStreamer pipeline status for advanced diagnostics.
EOF
