This is the rough process for governing the factory reset states/procedure.
This was designed prior to adoption of the Google Go Style Guide and codebase refactoring.
Specific variables and files mentioned may not be accurate. the general concepts and procedures shall apply.

### **Detailed State Machine Logic**

You then asked for the specific steps for each state. The following is the detailed logic from the "unambiguous query" document we created, which was based on your `devplan.txt` and `audit.md` files:

**1. Implement First-Boot Setup & Initialization Logic**

* **`systeminitialize` State Machine:** The backend must read a `systeminitialize` flag from `config.json`.
    * **State `3` (OEM First Boot):** On detecting this state, the backend must:
        1.  Generate a unique hostname (format: `nixon-[12-hex-chars]`).
        2.  Save this hostname to a new, separate config file named `nixonhost`.
        3.  Set the system's actual hostname using `hostnamectl`.
        4.  Update `config.json` to set `systeminitialize = 2`.
        5.  Trigger a full system reboot.

    * **State `2` (Factory Reset):** On detecting this state, the backend must:
        1.  Delete `studio.db` and all files in the `./recordings` directory.
        2.  Restore the factory default `config.json` (which must have `systeminitialize = 1`).
        3.  Trigger a graceful restart of the application service.

    * **State `1` (User Setup):** On detecting this state, the web server must serve a dedicated "Setup Wizard" UI instead of the main application.

* **Setup Wizard UI:** This UI must guide the user through:
    1.  Selecting an expertise level (Standard, Power User, Professional) to set the UI mode.
    2.  Setting a custom hostname (pre-filled with the generated one).
    3.  Creating an optional administrator account (display a persistent warning on the main UI if skipped).
    4.  Selecting the primary audio device.
    
    Upon completion or skipping, it must call an API endpoint that sets `systeminitialize = 0`.