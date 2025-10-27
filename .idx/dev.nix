# To learn more about how to use Nix to configure your environment
# see: https://firebase.google.com/docs/studio/customize-workspace
{ pkgs, ... }: {
  # Which nixpkgs channel to use.
  channel = "stable-24.05"; # or "unstable"

  # Use https://search.nixos.org/packages to find packages
  packages = [
     pkgs.go
   # pkgs.python311
   # pkgs.python311Packages.pip
    pkgs.nodejs_20
    pkgs.golangci-lint
   # Build tools required for CGo & GStreamer
    pkgs.pkg-config
    pkgs.gcc
   # pkgs.nodePackages.nodemon
    pkgs.pipewire
    pkgs.gst_all_1.gstreamer
    pkgs.gst_all_1.gst-plugins-base # Core plugins (incl. ALSA)
    pkgs.gst_all_1.gst-plugins-good # Good quality plugins and codecs
    pkgs.gst_all_1.gst-plugins-bad  # Less common plugins (incl. SRT)
    pkgs.gst_all_1.gst-plugins-ugly # Good plugins with potential licensing issues (incl. LAME for MP3)
    pkgs.libshout
    pkgs.nodePackages.npm
 ];

  # Sets environment variables in the workspace
  env = {
    NIXON_ENV = "development";
  };

  idx = {
    # Search for the extensions you want on https://open-vsx.org/ and use "publisher.id"
    extensions = [
      "golang.go"
      # "vscodevim.vim"
    ];

    # Enable previews
    previews = {
      enable = true;
      previews = {
        web = {
          # Example: run "npm run dev" with PORT set to IDX's defined port for previews,
          # and show it in IDX's web preview panel
          command = ["sh" "-c" "cd web && npm run dev"];
          manager = "web";
          env = {
            # Environment variables to set for your server
            PORT = "$PORT";
          };
        };
        backend = {
          # Run the Go backend application
          command = ["go" "run" "cmd/nixon/main.go"];
          manager = "web"; # Use 'web' to get a preview URL and expose the port
          env = {
            # Environment variables for the Go backend
            WEB_LISTENADDRESS = "0.0.0.0:$PORT"; # Make Go backend listen on IDX-provided port
            DATABASE_PATH = "/tmp/nixon.db"; # A default database path for development
            # Add other relevant environment variables as needed, e.g.:
            # NIXON_AUDIO_DEVICENAME = "default";
            # NIXON_WEBS_SECRET = "your-secret";
          };
        };
      };
    };

    # Workspace lifecycle hooks
    workspace = {
      # Runs when a workspace is first created
      onCreate = {
        # Install JS dependencies for the web UI
        npm-install-web = "npm install --prefix web";
        # Download Go modules for the backend
        go-mod-download = "go mod download";
      };
      # Runs when the workspace is (re)started
      onStart = {
        # The previews configured above will handle starting the applications
      };
    };
  };
}