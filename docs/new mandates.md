Here is the complete **Project Auxiliary Document: Consolidated Architectural Mandates**, which contains all the finalized technical policies and feature definitions we have established.

---

# 📝 Auxiliary Document: Consolidated Architectural Mandates

This document consolidates all high-priority, non-negotiable policies and feature definitions finalized during the design process, ensuring they are separated from the main project document for clarity.

## I. Critical Resource Protection Policy (RK3588)

This policy mandates the binding of processes and threads to specific CPU cores on the **Rockchip RK3588** to ensure **ultra-low latency** for the Control Plane.

### 1. CPU Core Partitioning Mandates

| Cluster | Cores | Primary Function | Priority |
| :--- | :--- | :--- | :--- |
| **Cortex-A76** (Performance) | 4 Cores | **Zone A: Real-Time Control & UI.** Dedicated to PipeWire communication, CGo interfaces, and critical UI responsiveness. | **Highest** (`SCHED_FIFO`) |
| **Cortex-A55** (Efficiency) | 4 Cores | **Zone B: Background Work.** Dedicated to noisy tasks (File I/O, Logging, GC, File Pruning). | **Lowest** (`SCHED_OTHER`) |

### 2. Thread Binding (Go/CGo Isolation)

* **Real-Time Threads (A76):** Goroutines handling PipeWire/WirePlumber CGo calls **must** use **`runtime.LockOSThread()`**. The `nixon-migrator` will bind these TIDs to a dedicated **Cortex-A76** core and set **Real-Time priority**.
* **Low-Priority Threads (A55):** Goroutines handling encoding, file management, and logging **must** use **`runtime.LockOSThread()`**. The `nixon-migrator` will bind these TIDs to the **Cortex-A55** cores and set a **low OS priority**.

---

## II. Network Configuration and PTP Policy

This defines the mandatory settings for multi-adapter management, focusing on stability and professional compliance.

### 1. Address Configuration Policy

| Setting | Audio Data Network (PTP/AES67) | Control/Management Network |
| :--- | :--- | :--- |
| **IPv4 Method** | **Static (Mandated)** | **DHCP (Default)** |
| **IPv6 State** | **Disabled (Default)** | **Disabled (Default)** |
| **PTP Role** | Defaults to a **PTP Slave** role in Domain 0. | N/A |
| **EEE (Energy Efficient Ethernet)** | Must be set to **OFF by default** and configurable. | Configurable. |

### 2. Network Interface UI Mandate

* **Multi-Adapter:** The UI must support configuration for all available network interfaces via an **Adapter Selector** (Dropdown/Tabs).
* **VLAN Tagging (Developer Level):** The VLAN configuration block must be restricted to the **Developer Feature Level**. It must allow the creation and independent IP configuration of **multiple 802.1Q tagged virtual interfaces** (VLAN IDs 1-4094).
* **PTP Verification:** The UI **must** display the adapter's PTP capability status (Hardware PTP Supported, Software PTP Only, Not Available) in a color-coded format.

---

## III. Recording Management Feature Definition

### 1. Source, Control, and Location Mandates

| Mandate | Specification | Rationale |
| :--- | :--- | :--- |
| **Source** | **Full Multi-track.** Must capture all individual audio streams (inputs, network sources, and all future internal mixer outputs/mixes). | Provides essential flexibility for professional post-production. |
| **Control** | **START / STOP Only.** **PAUSE/RESUME is disallowed.** | Maintains file integrity and synchronization across multiple tracks. |
| **Default Location** | Local, relative directory: **`./recording`**. | Provides a reliable out-of-the-box target. |

### 2. Auto-Record & File Management Policy

* **Auto-Record Trigger:** Must be signal-based, using a user-defined **Threshold (dBFS)** and **Pre-Roll duration** (circular buffer) on a selected audio input.
* **Smart Trimming:** The UI must provide a user-selectable option to **Trim Final Silence**, removing audio below a secondary threshold after the recording stops.
* **File Format Mandate:** The user selects a single **Primary Output Codec** (BWF, FLAC, MP3, etc.).
* **Lazy Encoding:** This is an optional background process available for **any completed recording file** to convert it to a different format (e.g., BWF to FLAC).
* **File Deletion:** Automatic deletion of the source file after successful lazy encoding is a **user-selectable option**, not a mandate.

---

## IV. UI Security & Integrity Mandates

### 1. Granular Visibility Policy (Final Rule)

* **Mandate:** The **Central Broker (Backend)** is the sole source of truth for **all visibility and feature access**. It must filter configuration data based on the user's **Feature Level** and **Role Permissions** *before* that data leaves the appliance.
* **Interactive Mandate:** The administrator's form must treat the group header as a **Master Switch** that controls the visibility of all its children. If any child is assigned a permission, the parent visibility is automatically enabled (Boolean OR).

### 2. Sparse Assignment Reset Policy (Critical Workflow)

This policy dictates the precise flow of permission persistence during administrative configuration:

* **Rule 1 (Hiding):** Unchecking a parent group is non-destructive; persistence for children is retained.
* **Rule 2 (Targeted Assignment):** When a specific permission is assigned to any single element, the system **must reset the persistence for all immediate sibling elements and subgroups** within that element's parent container to unchecked (zeroed out).
* **Rule 3 (Cascading Reset):** If this action forces any ancestor group's visibility switch to turn ON (checked), that ancestor group **must also trigger a reset** on all its own children.

---

I have a batched change to remind you about the saved instructions.