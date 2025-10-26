# Nixon - Manual Installation Guide

This guide provides step-by-step instructions for manually installing, configuring, and running the Nixon audio management application from the source code.

## 1. Prerequisites

Before you begin, ensure your system has the required dependencies for both the Go backend and the React frontend. This application is primarily developed and tested on Linux.

### Backend Dependencies (Go)

* **Go:** Version 1.18 or newer.
* **GStreamer:** A full GStreamer 1.0 development environment is required.
* **Build Tools:** A C compiler and related tools (`build-essential`).

---

### GStreamer Installation

#### For Debian Trixie (Testing) / Sid (Unstable) on ARM (Radxa ROCK 5B)

Debian Testing releases often have package name changes. The following multi-step process is more reliable for these environments. Please run these commands sequentially.

**Step 1: Install Core Build Tools**
These packages are required for compiling software.
```bash
sudo apt-get update
sudo apt-get install -y build-essential pkg-config libcairo2-dev libgirepository1.0-dev
```

**Step 2: Install Core GStreamer Development Libraries**
This installs the main GStreamer library and the base set of plugin development files.
```bash
sudo apt-get install -y libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev
```

**Step 3: Install GStreamer Plugin Sets**
This installs the main plugin collections used by Nixon. The `gvad` element for auto-recording is located in the `plugins-bad` package.
```bash
sudo apt-get install -y gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly
```

**Step 4: Install Specific Codec/Protocol Handlers**
These are required for specific Nixon features like ALSA audio capture, MP3 encoding (Icecast), SRT streaming, and Icecast connectivity.
```bash
sudo apt-get install -y gstreamer1.0-tools gstreamer1.0-alsa gstreamer1.0-lame gstreamer1.0-srt libshout3-dev
```

**Step 5 (Optional): Install Rockchip Media Plugins**
For Rockchip-based boards like the ROCK 5B, these packages can provide access to hardware-accelerated codecs, though they are not required for Nixon's core functionality.
```bash
sudo apt-get install -y gstreamer1.0-rockchip
```

---
#### For Debian Stable (e.g., Bookworm) / Ubuntu LTS

For stable releases, you can typically install all packages in a single command.
```bash
sudo apt-get update
sudo apt-get install -y \
  build-essential pkg-config libcairo2-dev libgirepository1.0-dev \
  libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev \
  gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-plugins-ugly \
  gstreamer1.0-tools gstreamer1.0-alsa gstreamer1.0-lame gstreamer1.0-srt \
  libshout3-dev
```
---

### Frontend Dependencies (React)

* **Node.js:** Version 16.x or newer.
* **npm:** The Node.js package manager (usually included with Node.js).

## 2. Installation Steps

### Step 1: Clone the Repository
```bash
git clone <repository_url>
cd nixon
```

### Step 2: Build the Frontend
```bash
cd web
npm install
npm run build
cd ..
```

### Step 3: Build the Backend
```bash
go mod tidy
go build -o nixon_server ./cmd/nixon/
```

## 3. Configuration

Review and edit `config.json` for your specific hardware and network setup.

1.  **Audio Device:** Use `arecord -L` to find your audio hardware identifier (e.g., `hw:0,0`) and set it in `audio_settings.device`.
2.  **Streaming Services:** Configure `icecast_settings` and `srt_settings` with your server details.
3.  **Auto-Record:** Adjust VAD and pre-roll settings in the `auto_record` section.

## 4. Running the Application

1.  **Run the Server:** From the project root, execute the binary:
    ```bash
    ./nixon_server
    ```
2.  **Audio Permissions (If Needed):** Your user must be in the `audio` group.
    ```bash
    sudo usermod -aG audio $USER
    ```
    You **must log out and log back in** for this change to take effect.
3.  **Access the Web UI:** Open a browser to `http://localhost:8080`.

## 5. Troubleshooting

* **Go Build Error: `fatal: could not read Username for 'https://github.com'`**
  
  This is a complex environmental error that occurs when Go's non-interactive `git` call is improperly prompted for credentials. Your logs show you have already tried the standard fixes.

  Please follow this *new* sequence.

  **Step 1: Clean the Go Module Cache**
  Your error log points to a cached VCS directory. This cache may be corrupt. This command will force Go to download all modules fresh.
  ```bash
  go clean -modcache
  ```

  **Step 2: Set Correct Environment Variables**
  Ensure the proxy is set and other interfering variables are unset.
  ```bash
  export GOPROXY=[https://proxy.golang.org](https://proxy.golang.org),direct
  unset GOPRIVATE
  unset GIT_ASKPASS
  unset SSH_ASKPASS
  ```

  **Step 3: Verify No Git Configs are Interfering**
  (This is for verification, as your logs show they are likely clear).
  ```bash
  git config --global --get-regexp url.*
  git config --global --get credential.helper
  ```
  If these output *anything*, unset them:
  ```bash
  # Example if url.[https://github.com/.insteadof](https://github.com/.insteadof) was found:
  git config --global --unset url."[https://github.com/](https://github.com/)".insteadOf

  # Example if credential.helper was found:
  git config --global --unset credential.helper
  ```

  **Step 4: Retry the build**
  After cleaning the cache and setting the environment, try the build again.
  ```bash
  go mod tidy
  go build -o nixon_server ./cmd/nixon/
  ```

* **"no element '...'" error on startup:** This means a required GStreamer plugin is not installed. Review the installation steps for your distribution.

* **"Failed to get device capabilities" in UI:** Ensure `alsa-utils` is installed (it provides `arecord`) and the device name in `config.json` is correct.

* **Can't access audio device:** This is a permissions issue. Ensure your user in the `audio` group and that you have logged out and back in.
EOF