# Comprehensive Reference: Managed Appliance Upgrade

This document formalizes the mandatory requirements and methodology for deploying in-place code upgrades to the Go-based audio appliance. The design guarantees data integrity, configuration persistence, and controlled orchestration in both **Networked** and **Air-Gapped (Isolated)** environments.

The foundation of this solution is **atomicity** (the process either succeeds completely or reverts fully to the last valid state).

> **Note:** This reference is prescriptive. Additional elements may be added, but all listed requirements are mandatory.

## I. Foundational Requirements and Tooling (R1–R10)

All development and implementation must adhere strictly to the **Google Go Style Guide (R8)**.

| ID | Requirement | Standard & Implementation | 
 | ----- | ----- | ----- | 
| **R1** | **Go Migration Tooling** | **Industry Standard:** All schema and configuration changes must be executed using a dedicated, mature Go migration package (e.g., `golang-migrate/migrate`). | 
| **R2** | **Configuration Version** | The `config.json` must contain a mandatory `"ConfigVersion": "X.Y.Z"` field. | 
| **R3** | **Non-Destructive Merging** | **Tool:** Use **`imdario/mergo`** to preserve existing user values and secrets over new defaults during configuration upgrades. | 
| **R4** | **Atomic DB Control** | **Integrity:** Requires an external **file lock** on the SQLite database + internal **atomic SQL transaction** for all schema changes. | 
| **R5** | **Integrity Guard** | The main application must **refuse to start** if a "migration in progress" lock file is detected (Anticipation Point F). | 
| **R6** | **APT Upgrade Lock** | Use `apt-mark hold` to prevent uncontrolled, unattended upgrades (Anticipation Point B). | 
| **R7** | **Companion Executable** | **Tool:** Create a privileged Go binary, **`nixon-migrator`**, using the **`klauspost/compress`** package for secure OS commands (`apt-mark`, `dpkg -i`). | 
| **R8** | **Coding Standard** | All code must adhere strictly to the **Google Go Style Guide**. | 
| **R9** | **Secure Local Distribution** | The Master must use **AES70 file transfer** (or SCP/SFTP over SSH as fallback) for encrypted, high-performance package distribution to all Slaves. | 
| **R10** | **Backup/Restore Feature (Migration)** | Implement a Web UI to export/import a version-stamped archive (`.zip`) containing `config.json` and `nixon.db`. **The restore process must use the `nixon-migrator` to perform a version-based data and configuration migration** against the currently running binary. | 

## II. Dual-Path Upgrade Methodology

The upgrade is orchestrated by the **Master Unit** and its **`nixon-migrator` (R7)** utility.

### Orchestration and Failure Anticipation (Anticipation Points)

| ID | Point | Purpose | 
 | ----- | ----- | ----- | 
| **A** | **Atomic Revert on Failure** | Guarantees the system restores the **last known working state** (data and config) if the migration fails. | 
| **F** | **Unattended Migration Recovery** | The system must detect an interrupted migration lock (R5) and automatically execute the restore procedure upon reboot. | 
| **G** | **Multi-Device Orchestration** | The Master unit directs all Slaves to update sequentially, ensuring **Slave data/config is ignored** and Slaves enter a "Ready for Adoption" state. | 
| **H** | **Pre-Migration Readiness Check** | A mandatory dry run executed by the Master, verifying permissions, locks, and Slave readiness before committing to the live upgrade. | 

### A. Pre-Migration Phase (Guaranteed Revert Point)

This phase guarantees an atomic restore point and is triggered by the `nixon-migrator` (R7).

1. **Stop Service & Lock DB (R4):** `nixon-migrator` stops the main service and applies a mandatory file lock on the database.

2. **Backup (A):** Creates a secure, version-stamped backup of the active `config.json` (R2) and the entire database file (`nixon.db`).

3. **Readiness Check (H):** Executes a dry run, verifying permissions, locks, and confirming the Slave units are ready.

4. **Slave Lock (R6, G):** Master commands all Slaves (via AES70) to execute `apt-mark hold` locally.

### B. Execution and Migration Phase (The Dual Path)

| Path | Convenience (Networked) | Security (Isolated/Air-Gapped) | 
 | ----- | ----- | ----- | 
| **1. Package Staging** | Master executes `apt update` (R7). | User uploads the `.deb` file via the Master's UI (R10). Master verifies the signature and stages the file. | 
| **2. Distribution (R9)** | Slaves are commanded to execute `apt install` locally. | Master uses **AES70/SCP** (R9) to push the staged `.deb` file to all Slaves' local storage. | 
| **3. Migration (R3, R4)** | `nixon-migrator` performs **Config Merge (R3)** and **DB Migration (R1, R4)** on the Master unit's local data. | `nixon-migrator` performs **Config Merge (R3)** and **DB Migration (R1, R4)** on the Master unit's local data. | 
| **4. Slave Execution (G)** | Master commands Slaves to execute `apt upgrade` locally. | Master commands Slaves to execute **`dpkg -i <filename>`** locally. | 
| **5. Final Cleanup (R5):** If successful across all units, the temporary lock file is deleted, and all services are restarted. |  |  | 

### C. Failure and Recovery (Atomic Rollback)

1. The `nixon-migrator` (R7) is triggered by the package manager's failure hook (`postrm`).

2. The executable restores the backup `config.json` and `nixon.db` files to the active directories (A).

3. The system is restored to its exact, uncorrupted, pre-upgrade state, guaranteeing data integrity.
