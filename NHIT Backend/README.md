# NHIT Ecosystem

This documentation covers the complete architecture of the NHIT system. The system is designed as a **distributed microservices architecture** that handles everything from user management and organization hierarchy to complex approval workflows (Green Notes) and asynchronous notifications.

## 🌍 Ecosystem Overview

The NHIT system is built to be modular. Instead of a single monolithic application, we have splits responsibilities across three main repositories:

1.  **NHIT Backend (`/NHIT Backend`)**:
    *   **Role**: The "Core Foundation".
    *   **Responsibilities**: It manages the static entities of the system—Users, Organizations, Departments, Projects, and Vendors. It also handles the **API Gateway**, which is the "front door" for all traffic.
    *   **Key Services**: User Service, Auth Service, Project Service, Vendor Service.

2.  **GreenNote Service (`/Nhit-Note`)**:
    *   **Role**: The "Business Domain".
    *   **Responsibilities**: This is a specialized service dedicated to the "Green Note" approval process. It handles the creation of notes, tracks their approval status through various levels, and ensures compliance with project budgets. It connects back to the Core Backend to verify users and projects.

3.  **Notification Service (`/notification-service`)**:
    *   **Role**: The "Async Worker".
    *   **Responsibilities**: It listens for events (like "Password Reset Requested" or "Green Note Approved") and sends out emails. It doesn't handle HTTP requests directly; it sits in the background processing messages from Kafka.

### 🏗️ Global Architecture

The diagram below visualizes how a request travels through the system. Notice how the **API Gateway** shields the internal services from the outside world, and how **Kafka** decouples the email sending from the main logic.

```mermaid
graph TD
    Client[Client Apps (Web/Mobile)] -->|1. HTTPS Request| Gateway[API Gateway :8080]
    
    subgraph "Synchronous Flow (gRPC)"
        Gateway -->|Route| Auth[Auth Service :50052]
        Gateway -->|Route| User[User Service :50051]
        Gateway -->|Route| Org[Organization Service :50053]
        Gateway -->|Route| Dept[Department Service :50054]
        Gateway -->|Route| Desig[Designation Service :50055]
        Gateway -->|Route| Project[Project Service :50057]
        Gateway -->|Route| Vendor[Vendor Service :50058]
        Gateway -->|Route| GreenNote[GreenNote Service :50059]
        
        %% Inter-service Communication
        GreenNote -.->|Validate| Project
        GreenNote -.->|Validate| Vendor
        GreenNote -.->|Validate| Dept
        Auth -.->|Verify| User
    end

    subgraph "Asynchronous Flow (Kafka)"
        Auth -->|Pub: PRODUCER| Kafka{Kafka Broker :9092}
        GreenNote -->|Pub: PRODUCER| Kafka
        
        Kafka -->|Sub: CONSUMER| Notification[Notification Worker]
    end
    
    subgraph "Storage Layer"
        Auth --> DB[(PostgreSQL :5433)]
        User --> DB
        Org --> DB
        Dept --> DB
        Desig --> DB
        Project --> DB
        Vendor --> DB
        GreenNote --> DB
        Notification --> DB
    end
```

## 🔄 Key Workflows Explained

To understand the system better, let's walk through two common scenarios:

### 1. User Login Flow
*   **Step 1**: The client sends a POST request with credentials to the **API Gateway**.
*   **Step 2**: The Gateway routes this to the **Auth Service** via gRPC.
*   **Step 3**: The Auth Service verifies the password against the database.
*   **Step 4**: If valid, it generates a **JWT (JSON Web Token)**.
*   **Step 5**: The JWT is returned to the user. Future requests must include this token in the header. The API Gateway validates this token before allowing access to any other service.

### 2. Green Note Creation & Approval
*   **Step 1 (Creation)**: A user submits a Green Note form. The **GreenNote Service** receives this.
*   **Step 2 (Validation)**: The service calls the **Project Service** (to check budget) and **Vendor Service** (to check supplier details) internally using gRPC.
*   **Step 3 (Persistence)**: If all checks pass, the note is saved to the database with status `PENDING`.
*   **Step 4 (Notification)**: The GreenNote service publishes a `green_note_created` event to **Kafka**.
*   **Step 5 (Email)**: The **Notification Service** picks up this event and instantly sends an email to the approver.

## 📂 Project Directory Structure

Here is how the project files are organized on your disk:

```text
d:\Nhit\
├── NHIT Backend/               # [CORE REPO]
│   ├── services/               # Contains the source code for all core microservices
│   │   ├── api-gateway/        # The HTTP Server that routes traffic
│   │   ├── auth-service/       # Login & Security logic
│   │   ├── user-service/       # User profiles & Role management
│   │   ├── ...                 # (Org, Dept, Designation, Project, Vendor)
│   ├── pkg/                    # Shared code used by all services (Database connections, Middleware)
│   ├── api/                    # Protobuf files (.proto) defining the data contracts
│   └── docker-compose.yml      # Config to run the whole stack locally
│
├── Nhit-Note/                  # [GREENNOTE REPO]
│   ├── services/
│   │   └── greennote-service/  # The specific logic for Green Notes
│   └── ...
│
└── notification-service/       # [NOTIFICATION REPO]
    └── cmd/
        └── service/            # The main worker program that listens to Kafka
```

## 🚀 Service Registry

| Service | Port | Description |
| :--- | :--- | :--- |
| **API Gateway** | `8080` | **The Traffic Cop.** It authenticates every request and directs it to the right service. |
| **User Service** | `50051` | **The Address Book.** Stores names, emails, and roles. |
| **Auth Service** | `50052` | **The Bouncer.** Handles passwords and issues security tokens. |
| **Organization Service** | `50053` | **The Structure.** Manages the company hierarchy (Head Office vs. Branch). |
| **Project Service** | `50057` | **The Planner.** Tracks active projects and their budgets. |
| **Vendor Service** | `50058` | **The Supplier List.** Manages contractor details and banking info. |
| **GreenNote Service** | `50059` | **The Workflow Engine.** Manages the lifecycle of approval notes. |
| **Notification Service** | *N/A* | **The Messenger.** Sends emails. It doesn't have a port because it listens to Kafka. |

## 🚦 Quickstart Guide

### 1. Start Infrastructure
The infrastructure (Database and Kafka) is shared. Run it from the core backend folder:
```bash
cd "d:\Nhit\NHIT Backend"
docker-compose up postgres kafka -d
```

### 2. Run the Core Logic
You need to run the services you are working on. For example, to run the **User Service**:
```bash
cd "d:\Nhit\NHIT Backend\services\user-service"
go run ./cmd/server/main.go
```

### 3. Run the GreenNote System
If you are working on approvals, you must also run the **GreenNote Service**:
```bash
cd "d:\Nhit\Nhit-Note\services\greennote-service"
go run ./cmd/server/main.go
```

## 🛠️ Technology Stack
*   **Go (Golang)**: Chosen for its speed and native support for concurrency is perfect for microservices.
*   **gRPC**: A high-performance RPC framework. It's faster than REST because it uses binary data (Protobuf) instead of text (JSON).
*   **PostgreSQL**: A robust relational database for safe data storage.
*   **Kafka**: A streaming platform used for "fire and forget" tasks like sending emails, ensuring the main app doesn't slow down waiting for email servers.
