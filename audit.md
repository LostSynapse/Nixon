### 1. Roadmap Understanding

The roadmap is clear on the planned feature set:

* **Phase 1 (Complete):** Core Go/GStreamer audio engine (SRT/Icecast) with SQLite database and a React/WebSocket real-time UI for basic control and recording management.
* **Phase 2 (Automation):** Focus on making the core recorder intelligent via VAD, pre-roll buffering (crucial feature for audio), and auto-purge disk management.
* **Phase 3 (Security):** Implementing user roles, authentication, and ownership (requires updates to the \`db\` package and API handlers).
* **Phase 4 (Advanced):** Advanced audio settings (channel mapping), Web Player enhancements (waveform), and a system dashboard.

### 2. Primary Architectural Conflict: Modularity

The single largest deviation from our agreed-upon goal is the tightly-coupled architecture. To achieve true modularity and prepare for AES70/OSC plugins, a significant refactoring is required to establish the necessary abstraction layers.

| Requirement | Current Status | Recommendation for Alignment |
| :--- | :--- | :--- |
| **Stream Pluginization** | The SRT and Icecast streaming protocols are **hardcoded** into \`internal/config/config.go\`, \`internal/api/router.go\`, and \`internal/gstreamer/gstreamer.go\`. | **Implement a \`Plugin\` interface and \`StreamService\` abstraction.** Refactor \`gstreamer.go\`'s \`buildPipelineString\` to accept a list of registered \`StreamPlugin\`s that dynamically contribute their GStreamer branches and configuration to the main pipeline. |
| **Control Abstraction** | **Missing** (Not in Roadmap) | **Implement a \`ControlAbstractionLayer\` in Go.** This layer will define a common API (e.g., \`StartStream(name string)\`, \`GetStatus()\`) that the UI, AES70, and OSC will call, decoupling them from the specific GStreamer implementation. |
| **Dynamic UI** | The React components (\`App.jsx\`, \`Modals.jsx\`) **statically import** and render tabs/controls for SRT and Icecast. | The Go backend must provide an API endpoint (\`/api/plugins\`) that returns **JSON schemas and UI metadata** for all loaded plugins. The React frontend must use a **Schema-Driven Forms** library to render the Settings tabs dynamically. |

### 3. Code Audit Findings & Efficiency

The Go code is generally well-structured for a non-modular system, but several issues conflict with the **Google Go Style Guide**, **efficiency**, and **redundancy avoidance**.

| File | Finding | Standard Violated / Recommendation |
| :--- | :--- | :--- |
| \`internal/config/config.go\` | **Redundant Fields:** Top-level fields \`SRTEnabled\` and \`IcecastEnabled\` are duplicated and synchronized with their nested struct fields (\`SrtSettings.SrtEnabled\`, etc.). | **Redundancy:** Remove top-level \`SRTEnabled\` and \`IcecastEnabled\`. Rely solely on the nested struct fields, as they are the source of truth. |
| \`internal/config/config.go\` | **Legacy/Unused Fields:** Structs still contain legacy fields (e.g., \`AudioSettings.Bitrate\`, \`AudioSettings.Channels\`, \`AutoRecord.TimeoutSeconds\`). | **Efficiency/Clarity:** Remove unused or legacy fields from all config structs to simplify configuration and reduce surface area for errors. |
| \`cmd/nixon/main.go\` | **Graceful Shutdown:** The \`StopPipelineGraceful\` call in the shutdown sequence is not defined in \`gstreamer.go\`. Only \`StopPipeline\` exists. | **Errors/Maintainability:** Rename the function call in \`main.go\` to the existing \`manager.StopPipeline()\` or implement a dedicated graceful stop sequence in \`gstreamer.go\`. |
| \`internal/db/db.go\` | **GORM Logging:** Uses the GORM \`logger.New\` with \`os.Stdout\`, which is not ideal for high-performance logging, although acceptable in this context. | **Efficiency:** Consider logging only to a file or using a more performance-tuned logger (like \`slog\` in Go 1.21+) for production appliances, especially when verbose GORM logs are enabled. |
| \`internal/gstreamer/gstreamer.go\` | **Mutex Usage:** The \`managerMutex\` for accessing the global manager is a standard \`sync.Mutex\`, but the logic for checking if \`manager == nil\` is repeated in multiple public functions. | **Efficiency/Clarity:** Consolidate manager retrieval into a single, reliable getter function that handles the nil check and locks appropriately. |
| \`internal/api/router.go\` | **GORM/API Inconsistency:** The \`UpdateRecordingHandler\` request body uses **pointer types** (\`*string\`) for optional fields, while the \`db.Recording\` model uses **non-pointer strings**. | **Best Practice/Consistency:** Modify the request body struct to match the \`db.Recording\` model (use non-pointer strings), simplifying the update logic and adhering to best practices for data consistency. |
| \`internal/api/tasks.go\` | **Temporary Error Handling:** The \`monitorIcecastListeners\` function does not reset listener count on temporary HTTP errors. | **Robustness:** While keeping the last state is good for transient errors, a more robust system would include a counter to reset the status to '0' after several consecutive failures, to accurately reflect stream outages. |

