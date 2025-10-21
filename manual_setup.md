This guide provides a step-by-step process to manually compile, install, and run the Nixon appliance on your AutoRecorder.

### Step 1: Install Build Dependencies

We need Go, Node.js/NPM, SQLite, a C compiler (`gcc`), `pnpm`, and a new GStreamer plugin for the auto-recording feature.

```bash
sudo apt-get update
sudo apt-get install -y golang-go nodejs npm git sqlite3 build-essential gstreamer1.0-plugins-good
sudo npm install -g pnpm
```

### Step 2: Set Up Project Directory

Create the complete directory structure for the project in your home directory.

```bash
# Clean start
rm -rf ~/nixon
mkdir -p ~/nixon/cmd/nixon
mkdir -p ~/nixon/internal/api
mkdir -p ~/nixon/internal/config
mkdir -p ~/nixon/internal/db
mkdir -p ~/nixon/internal/gstreamer
mkdir -p ~/nixon/internal/state
mkdir -p ~/nixon/internal/websocket
mkdir -p ~/nixon/web/public
mkdir -p ~/nixon/web/src/components
```

### Step 3: Place All Project Files

Place all the files from your repository into their corresponding locations within the `~/nixon` directory. The correct structure is:

```
nixon/
+-- cmd/nixon/
¦   +-- main.go
+-- go.mod
+-- config.json
+-- internal/
¦   +-- api/
¦   ¦   +-- handlers.go
¦   ¦   +-- router.go
¦   ¦   +-- tasks.go
¦   +-- config/
¦   ¦   +-- config.go
¦   +-- db/
¦   ¦   +-- db.go
¦   +-- gstreamer/
¦   ¦   +-- gstreamer.go
¦   +-- state/
¦   ¦   +-- state.go
¦   +-- websocket/
¦       +-- websocket.go
+-- README.md
+-- web/
    +-- package.json
    +-- index.html
    +-- tailwind.config.js
    +-- postcss.config.js
    +-- public/
    ¦   +-- nixon_logo.svg
    +-- src/
        +-- App.jsx
        +-- main.jsx
        +-- index.css
        +-- components/
            +-- DiskUsage.jsx
            +-- Modals.jsx
            +-- RecordingControl.jsx
            +-- RecordingsList.jsx
            +-- StreamControl.jsx
```

### Step 4: Build the Go Backend

Navigate to the project root (`~/nixon`) and build the Go application.

```bash
cd ~/nixon

# This command downloads dependencies and creates the go.sum file.
go mod tidy

# Compile the application with CGO enabled for the SQLite driver.
CGO_ENABLED=1 go build -o nixon ./cmd/nixon
```

### Step 5: Build the React Frontend

This step uses `pnpm` and has been updated to resolve the native dependency issue.

```bash
# This single block of commands will handle the entire frontend build process.
# Run it from the ~/nixon directory.
cd web && \
rm -rf node_modules pnpm-lock.yaml && \
pnpm install --no-optional && \
pnpm run build && \
cd .. && \
rm -rf web/assets web/index.html web/nixon_logo.svg && \
mv web/dist/* web/ && \
rm -rf web/dist web/src web/node_modules web/pnpm-lock.yaml web/package.json web/tailwind.config.js web/postcss.config.js web/public
```

### Step 6: Test the Application

Run the appliance directly to test it.

```bash
# From ~/nixon
./nixon
```
Access the web interface at `http://<your-AutoRecorder-ip>:8080`. Press `CTRL+C` to stop the application when you are done testing.

### Step 7: Create the systemd Service File

To make the application run as a background service, create a `systemd` service file.

Create and open the file for editing:
```bash
sudo nano /etc/systemd/system/nixon.service
```

Paste the following content into the file. **Important:** Replace `your_user` with your actual username on the AutoRecorder (e.g., `rock`, `pi`, `ubuntu`).

```ini
[Unit]
Description=Nixon AutoRecorder Service
After=network.target

[Service]
Type=simple
User=your_user
WorkingDirectory=/home/your_user/nixon
ExecStart=/home/your_user/nixon/nixon
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Save the file and exit the editor (in `nano`, press `CTRL+X`, then `Y`, then `Enter`).

### Step 8: Enable and Start the Service

Reload the `systemd` daemon to make it aware of the new service, then enable and start it.

```bash
# Reload the systemd manager configuration
sudo systemctl daemon-reload

# Enable the service to start on boot
sudo systemctl enable nixon.service

# Start the service immediately
sudo systemctl start nixon.service

# Check the status of the service
sudo systemctl status nixon.service