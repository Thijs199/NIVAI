# AIFAA Football Analytics Platform

This document provides instructions on how to set up, configure, and run the various components of the AIFAA Football Analytics Platform locally for development and testing.

The platform consists of three main services:
1.  **Go Backend:** Handles API requests, user authentication, video/data uploads, and orchestrates calls to the Python API.
2.  **Python API:** Performs intensive physical statistics calculations on tracking and event data.
3.  **Next.js Frontend:** Provides the user interface for uploading data, viewing matches, and analyzing dashboards.

## System Architecture Overview

This section provides a high-level overview of the AIFAA Football Analytics Platform, detailing its components, how they interact, and the flow of data through the system.

The platform is designed to ingest football match data (videos, tracking information, event data), process it to extract valuable analytics, and present these insights to the user through a web interface.

### Components

The platform is comprised of three core components:

*   **Frontend:** A Next.js (React) web application serving as the primary user interface. It allows users to:
    *   Upload match-related data (videos, tracking files, event files).
    *   View a list of available matches and their current analytics processing status.
    *   Access detailed dashboards for individual matches, showcasing various statistics and visualizations.
    *   It interacts with the Go Backend via RESTful API calls.

*   **Go Backend:** A Go-based API built using the `gorilla/mux` router. Its main responsibilities include:
    *   Handling user authentication and authorization.
    *   Managing metadata for videos and matches (e.g., titles, descriptions, team information, file paths), which is typically stored in a PostgreSQL database.
    *   Processing file uploads from the frontend and storing them in a designated file storage solution (e.g., local filesystem, Azure Blob Storage).
    *   Orchestrating the analytics pipeline by triggering the Python API for processing and fetching analytics status/results.
    *   Serving as a proxy for analytics data requests from the frontend to the Python API.

*   **Python API:** A Python FastAPI application dedicated to computational tasks. Its key functions are:
    *   Exposing an internal API endpoint (`/process-match`) that the Go Backend calls to initiate the asynchronous analysis of tracking and event data for a specific match.
    *   Reading data files from the shared storage location based on paths provided by the Go Backend.
    *   Performing intensive calculations to generate physical statistics, player metrics, and team analytics.
    *   Providing endpoints for the Go Backend to query the status of processing (`/match/{id}/status`) and retrieve the computed analytics data (e.g., `/match/{id}/stats/summary`, `/match/{id}/player/{player_id}/details`).

### Data Flow

The following outlines the primary data flows within the AIFAA platform:

1.  **Video/Data Upload and Analytics Processing Trigger:**
    *   **User (Frontend):** Initiates an upload of match-related files (video is optional, tracking and event data files are typically required for analytics) along with any associated metadata (title, teams, etc.) through the web interface.
    *   **Frontend to Go Backend:** Sends a `POST` request (e.g., to `/api/v1/videos`) with the files and metadata.
    *   **Go Backend (`VideoController`):**
        *   Generates a unique ID for the match/video.
        *   Saves the uploaded files to the configured file storage (e.g., Azure Blob Storage, local disk). Paths are typically structured using the generated ID.
        *   Stores metadata (including file paths, title, user-provided details) in the PostgreSQL database.
        *   Makes a `POST` request to the **Python API**'s `/process-match` endpoint, providing the unique match ID and the storage paths to the tracking and event data files.
    *   **Python API:**
        *   Receives the request and queues the data for asynchronous processing.
        *   Acknowledges the request to the Go Backend (e.g., with a `202 Accepted` response).
        *   In the background, it reads the specified data files from the shared storage, performs complex calculations, and stores the resulting analytics.

2.  **Viewing Match List and Analytics Status:**
    *   **User (Frontend):** Navigates to the page displaying the list of matches (e.g., `/matches`).
    *   **Frontend to Go Backend:** Sends a `GET` request (e.g., to `/api/v1/matches`).
    *   **Go Backend (`MatchController`):**
        *   Fetches the list of matches/videos from the PostgreSQL database.
        *   For each match in the list, it makes a `GET` request to the **Python API**'s `/match/{match_id}/status` endpoint to retrieve the current analytics processing status.
        *   Compiles the list of matches along with their respective analytics statuses.
    *   **Go Backend to Frontend:** Returns the combined list (match details + analytics status) as a JSON response.
    *   **Frontend:** Displays the list of matches, often color-coding or indicating the status of analytics (e.g., Pending, Processing, Processed, Error).

3.  **Viewing Detailed Match Analytics:**
    *   **User (Frontend):** Selects a specific match from the list to view its detailed analytics dashboard (e.g., navigates to `/dashboard/{match_id}`).
    *   **Frontend to Go Backend:** Sends `GET` requests for specific analytics data, (e.g., to `/api/v1/analytics/matches/{match_id}` for summary stats, or `/api/v1/analytics/players/{player_id}?match_id={match_id}` for player-specific data).
    *   **Go Backend (`AnalyticsController`):**
        *   Acts as a proxy. Forwards these requests to the relevant endpoints on the **Python API** (e.g., `/match/{match_id}/stats/summary`, `/match/{match_id}/player/{player_id}/details`).
    *   **Python API:**
        *   Retrieves the requested pre-computed/processed analytics data.
        *   Returns the data in JSON format to the Go Backend.
    *   **Go Backend to Frontend:** Relays the JSON response from the Python API back to the frontend.
    *   **Frontend:** Renders the received data into charts, tables, and other visualizations on the match dashboard.

### Data Sources

The platform utilizes several data storage and configuration mechanisms:

*   **PostgreSQL Database:**
    *   **Usage:** Primarily managed by the Go Backend.
    *   **Content:** Stores metadata such as user accounts, video/match details (e.g., title, description, associated team names, competition, season), paths to files in the File Storage, and potentially other relational application data.
    *   **Interaction:** The Go Backend performs CRUD (Create, Read, Update, Delete) operations on this database.

*   **File Storage (e.g., Azure Blob Storage, Local Filesystem):**
    *   **Usage:** Stores large binary files.
    *   **Content:** Raw uploaded files, including video files (e.g., MP4, AVI), tracking data files (e.g., Parquet, GZIP compressed), and event data files.
    *   **Interaction:**
        *   The **Go Backend** writes files to this storage upon upload by the user.
        *   The **Python API** reads these files from this storage for analytics processing, using file paths provided by the Go Backend.
    *   **Critical Note:** For the system to function, the Python API must have read access to the locations where the Go Backend stores these files. In a distributed setup, this often involves shared network storage, cloud storage buckets with appropriate permissions, or persistent volumes in containerized environments. The `EXTERNAL_DATA_MOUNT` variable (or similar mechanism in Docker) facilitates this.

*   **Python API Internal Data Management:**
    *   **Usage:** The Python API processes raw data and generates analytical results.
    *   **Content:** The derived statistics, time-series data, and other analytical outputs.
    *   **Interaction:** The Python API serves this processed data via its own API endpoints. The exact internal storage mechanism (e.g., in-memory cache, file-based cache, Redis, or a dedicated database for analytics results) is an implementation detail of the Python API, but it's abstracted away from other components, which simply query its API.

*   **Configuration Files & Environment Variables:**
    *   **Go Backend:** Uses environment variables (potentially loaded from a `.env` file or `config.json`) for settings like database connection strings, file storage credentials (e.g., Azure Blob Storage account keys), the URL of the Python API (`PYTHON_API_URL`), and server port.
    *   **Python API:** Configured via its startup command (e.g., host/port for Uvicorn) and relies on the Go Backend to provide paths to data files. May have its own internal configurations for processing parameters.
    *   **Frontend:** Uses environment variables (e.g., `.env.local`) for settings like the Go Backend API URL (`NEXT_PUBLIC_API_BASE_URL`).

### Key Interactions & API Endpoints

While not exhaustive, this list highlights some of the primary API endpoints and interactions between services:

*   **Frontend -> Go Backend:**
    *   `POST /api/v1/auth/login`: User authentication.
    *   `POST /api/v1/videos`: Uploading video, tracking, and event data.
    *   `GET /api/v1/matches`: Fetching the list of available matches and their analytics statuses.
    *   `GET /api/v1/analytics/matches/{match_id}`: Fetching summary analytics for a specific match.
    *   `GET /api/v1/analytics/players/{player_id}?match_id={match_id}`: Fetching detailed time-series data for a player in a match.
    *   `GET /api/v1/analytics/teams/{team_id}?match_id={match_id}`: Fetching team summary data over time for a match.

*   **Go Backend -> Python API:**
    *   `POST /process-match`: (Called by Go Backend's `VideoController` after file upload)
        *   **Request Body:** `{ "tracking_data_path": "...", "event_data_path": "...", "match_id": "..." }`
        *   **Purpose:** To trigger asynchronous processing of the specified match data files.
    *   `GET /match/{match_id}/status`: (Called by Go Backend's `MatchController` when listing matches)
        *   **Purpose:** To get the current analytics processing status for a match.
        *   **Response Body (Example):** `{ "status": "processed", "match_id": "..." }`
    *   `GET /match/{match_id}/stats/summary`: (Called by Go Backend's `AnalyticsController` as a proxy)
        *   **Purpose:** To retrieve overall player and team statistics for a processed match.
    *   `GET /match/{match_id}/player/{player_id}/details`: (Called by Go Backend's `AnalyticsController` as a proxy)
        *   **Purpose:** To retrieve detailed time-series data for a specific player.
    *   `GET /match/{match_id}/team/{team_id}/summary-over-time`: (Called by Go Backend's `AnalyticsController` as a proxy)
        *   **Purpose:** To retrieve aggregated team metrics over time intervals.

*   **Internal Go Backend Interactions:**
    *   Interacts with PostgreSQL database for metadata management.
    *   Interacts with the chosen File Storage solution (e.g., Azure Blob Storage) for file I/O.

*   **Internal Python API Interactions:**
    *   Reads files from the shared File Storage.
    *   Manages its internal cache/storage for computed analytics.

### Architecture & Data Flow Diagram

The following diagram provides a simplified visual representation of the system architecture and the main data flows described above:

```text
+-----------------+      +---------------------+      +-------------------+
|   User Browser  |----->|  Frontend (Next.js) |      | File Storage      |
| (Web Interface) |      | (nginx/standalone)  |      | (Azure Blob / FS) |
+-----------------+      +----------^----------+      +---------^---------+
      ^      |                      |   ^                        |   ^
      |      | (HTML/JS/CSS)        |   | (JSON API)             |   | (Files)
      |      |                      |   |                        |   |
(Display)   (Upload Req)            |   |                        |   |
      |      |                      |   +------------------------+---+ (Read/Write Files)
      |      |                      |   | (API Calls)            |   |
      |      v                      v   |                        |   |
      +--------------------------+ |   |                        |   |
      |                          | |   |                        |   |
      |     Go Backend (Go API)  | |   |                        |   |
      | (Docker Container / Host)| |   |                        |   |
      |                          | |   |                        |   |
      |  - Auth                  <---+                            |   |
      |  - Video/Match Metadata  |                                |   |
      |  - File Upload Handler   |---->(Save Files)---------------+   |
      |  - Analytics Proxy       |                                   |
      |                          |                                   |
      |  (Interacts with DB)     |---(Python API Calls)------------+ | (Read Files for Processing)
      |  +-------------------+   |                                 | |
      |  | PostgreSQL DB     |<--+                                 | |
      |  +-------------------+   |                                 | |
      +--------------------------+                                 | |
                  ^                                                | |
                  | (JSON API - Process, Status, Data)             | |
                  |                                                | |
                  v                                                | |
      +--------------------------+                                 | |
      | Python API (FastAPI)     |                                 | |
      | (Docker Container / Host)|                                 | |
      |                          |<---(Read Source Data Files)------+
      |  - Data Processing       |
      |  - Statistics Engine     |
      |  - Serves Analytics Data |
      |  (Internal Cache/Store)  |
      +--------------------------+
```

**Diagram Legend & Notes:**

*   Arrows (`-->`, `<--`, `<-->`) indicate the direction of data flow or requests.
*   `User Browser`: Represents the end-user interacting with the system.
*   `Frontend`: The Next.js application rendering the UI and making API calls.
*   `Go Backend`: The core API service handling business logic, data management, and orchestration.
*   `Python API`: The specialized service for data processing and analytics.
*   `File Storage`: Represents where raw data files (video, tracking, events) are stored (e.g., Azure Blob Storage, a local/network file system).
*   `PostgreSQL DB`: The relational database for metadata.
*   `(API Calls)` / `(JSON API)`: Indicate typical RESTful API interactions.
*   The diagram simplifies some aspects for clarity, such as detailed network configurations or specific authentication flows within API calls.

*(As per user suggestion, a more detailed Mermaid diagram could be added here in the future if desired.)*

## Running the Platform with Docker Compose

This is the recommended way to run the entire AIFAA Football Analytics Platform for development and testing, as it simplifies setup and ensures consistency across services.

### Prerequisites for Docker Compose

*   **Docker Desktop:** Installed and running on your system. Download from [https://www.docker.com/products/docker-desktop/](https://www.docker.com/products/docker-desktop/).

### Configuration

1.  **Environment Variables:** A `.env` file in the project root can be used to set environment variables for `docker-compose`. An example is `COMPOSE_PROJECT_NAME=nivai_dashboard`. Service-specific variables like database credentials or cloud keys can also be placed here if not already defined in the `docker-compose.yml` or if you need to override them.
    The `docker-compose.yml` file is pre-configured for inter-service communication (e.g., frontend calling backend, backend calling python-api).

2.  **Shared Data:** The `docker-compose.yml` defines a named volume (`shared_data`) that is mounted into both the Go Backend (`/data/shared`) and the Python API (`/data/shared`). This volume is used for storing uploaded match files (tracking data, event data) so that the Python API can access files saved by the Go Backend. The `EXTERNAL_DATA_MOUNT` environment variable for the Go Backend is set to `/data/shared` within its container.

### Running the Application

1.  **Open a terminal** in the root directory of the project (where `docker-compose.yml` is located).
2.  **Build and start all services:**
    ```bash
    docker-compose up --build
    ```
    *   The `--build` flag ensures images are built if they don't exist or if Dockerfiles have changed.
    *   To run in detached mode (in the background), add the `-d` flag: `docker-compose up --build -d`.
3.  **Accessing Services:**
    *   **Frontend (Main Application):** `http://localhost:3000`
    *   **Go Backend API:** `http://localhost:8080`
    *   **Python API:** `http://localhost:8081` (primarily for internal use by the Go backend)

### Stopping the Application

1.  If running in the foreground (without `-d`), press `Ctrl+C` in the terminal.
2.  If running in detached mode, or to ensure all services and networks are removed:
    ```bash
    docker-compose down
    ```
    *   To also remove volumes (including the `shared_data` volume, which will delete all uploaded match data stored in it):
        ```bash
        docker-compose down -v
        ```

### Notes

*   The first time you run `docker-compose up --build`, it might take some time to download base images and build the application images.
*   Log output from all services will be aggregated in your terminal if running in the foreground. If detached, use `docker-compose logs -f` to follow logs, or `docker-compose logs <service_name>` for specific service logs.
*   The section "Manual Setup: Data Storage/Access Notes" below provides more context on how data is shared, which is handled by the Docker volume in this setup.

## Manual Setup: Prerequisites

These instructions are for setting up each service manually, without the primary Docker Compose environment.

Ensure the following software is installed on your development machine:

*   **Go:** Version 1.18 or higher.
*   **Node.js:** Version 18.x or higher.
    *   **npm:** Version 8.x or higher (usually comes with Node.js).
    *   (Alternatively, Yarn can be used if preferred for frontend package management).
*   **Python:** Version 3.9 or higher.
*   **Poetry:** Python dependency management tool. Follow installation instructions at [https://python-poetry.org/docs/](https://python-poetry.org/docs/).
*   **Docker & Docker Compose:** For manual setup, Docker is not strictly required unless you plan to containerize parts of it yourself. If you wish to use the primary Docker Compose setup, see the section "Running the Platform with Docker Compose".

## Manual Setup: Configuration

Each service may require environment variables for proper configuration. It's recommended to use `.env` files where supported (e.g., `backend/.env`, `frontend/.env.local`). Example environment files (`.env.example`) are provided in the respective service directories.

### 2.1. Go Backend (`backend/`)

*   **`PYTHON_API_URL`**: The full URL of the Python statistics API.
    *   Example: `PYTHON_API_URL=http://localhost:8081`
*   **`EXTERNAL_DATA_MOUNT`**: The base path on the host machine where uploaded files (videos, tracking data, event data) are stored. This path must be accessible by both the Go backend (for writing) and the Python API (for reading). See section "Manual Setup: Data Storage/Access Notes" for critical details.
    *   Example: `EXTERNAL_DATA_MOUNT=/path/to/your/nivai_data_storage`
    *   The Go service's `FileStorageService` will create subdirectories within this path (e.g., `/path/to/your/nivai_data_storage/videos/...`).
*   **`SERVER_PORT`**: Port on which the Go backend will run.
    *   Example: `SERVER_PORT=8000` (The Go backend is often referred to as running on port 8000 in frontend calls if not proxied by Next.js itself).
*   **Database Configuration:** (If a database like PostgreSQL is fully integrated for user management, video metadata, etc.)
    *   `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
*   An example configuration can be found in `.env.example` at the project root (which might be copied to `backend/.env`).

### 2.2. Python API (`python_api/`)

*   **Port & Host:** Configured directly in the `uvicorn` run command. Default is `0.0.0.0:8081`.
*   No specific environment variables are strictly required by the Python API itself for its core logic, as file paths are passed to it via API requests. However, it needs read access to the `EXTERNAL_DATA_MOUNT` location specified for the Go backend.

### 2.3. Next.js Frontend (`frontend/`)

*   **`NEXT_PUBLIC_API_BASE_URL`** (Recommended): The full base URL for the Go backend API.
    *   Example: `NEXT_PUBLIC_API_BASE_URL=http://localhost:8000/api/v1`
    *   If this is not set, frontend `fetch` calls to relative paths like `/api/v1/...` assume the Go backend is served on the same origin as the Next.js app (e.g., port 3000), or that Next.js rewrites/proxies are configured in `next.config.ts` to route these requests to the Go backend. The current implementation in frontend pages uses relative paths like `/api/v1/videos`, implying such a proxy or same-origin setup.
*   Create a `.env.local` file in the `frontend/` directory by copying from `.env.local.example` if it exists, and modify as needed.

## Manual Setup: Running Individual Services

Run each service in a separate terminal window.

### 3.1. Python API (`python_api/`)

1.  **Navigate to the directory:**
    ```bash
    cd python_api
    ```
2.  **Install dependencies:**
    ```bash
    poetry install
    ```
3.  **Activate virtual environment (optional but recommended):**
    ```bash
    poetry shell
    ```
    (If you don't use `poetry shell`, prefix commands with `poetry run`)
4.  **Run the API:**
    ```bash
    poetry run uvicorn src.api.main:app --host 0.0.0.0 --port 8081 --reload
    ```
    The API will be available at `http://localhost:8081`.

### 3.2. Go Backend (`backend/`)

1.  **Navigate to the directory:**
    ```bash
    cd backend
    ```
2.  **(Optional) Set environment variables:** Create a `.env` file in this directory or ensure they are set in your shell. See `../.env.example`.
    ```bash
    # Example:
    # export PYTHON_API_URL=http://localhost:8081
    # export EXTERNAL_DATA_MOUNT=/your/chosen/data/path
    # export SERVER_PORT=8000
    ```
3.  **Install dependencies (if necessary):**
    ```bash
    go mod tidy
    ```
4.  **Run the backend:**
    ```bash
    go run cmd/api/main.go
    ```
    The backend will typically run on the port specified by `SERVER_PORT` (e.g., `http://localhost:8000`).

### 3.3. Next.js Frontend (`frontend/`)

1.  **Navigate to the directory:**
    ```bash
    cd frontend
    ```
2.  **(Optional) Set environment variables:** Create a `.env.local` file if you need to override `NEXT_PUBLIC_API_BASE_URL` or other Next.js specific variables.
3.  **Install dependencies:**
    ```bash
    npm install
    ```
    (or `yarn install` if using Yarn)
4.  **Run the development server:**
    ```bash
    npm run dev
    ```
    (or `yarn dev`)
    The frontend will typically be available at `http://localhost:3000`.

## Manual Setup: Order of Startup

While services should ideally be resilient, the recommended startup order for a smooth experience is:

1.  **Python API:** Ensures it's ready to receive processing requests from the Go backend.
2.  **Go Backend:** Depends on the Python API for some functionalities (triggering processing, getting analytics status).
3.  **Next.js Frontend:** Depends on the Go Backend for API calls.

## Manual Setup: Accessing the Application

Once all services are running:

*   Open your web browser and navigate to the Next.js frontend URL: **`http://localhost:3000`**
*   Key pages to explore:
    *   **Upload Matches:** `http://localhost:3000/upload`
    *   **View Matches:** `http://localhost:3000/matches`
    *   **Match Dashboard:** (Navigate from "View Matches" page by clicking a match) e.g., `http://localhost:3000/dashboard/your_match_id`

## Manual Setup: Data Storage/Access Notes

For the system to function correctly, especially the analytics processing, both the Go Backend and the Python API must have access to the same file storage location for tracking and event data.

*   When files are uploaded via the frontend, the Go Backend saves them to a path derived from its `StorageService` configuration. This service uses the `EXTERNAL_DATA_MOUNT` environment variable as the root for storing these files (e.g., `EXTERNAL_DATA_MOUNT/videos/match_id_components/match_id/filename.ext`).
*   When the Go Backend calls the Python API's `/process-match` endpoint, it sends the *exact file paths* where it stored the tracking and event data.
*   **The Python API service must be able to read files from these exact paths.**
*   **Local Development Setup:**
    *   The simplest way to achieve this locally is to run both the Go Backend and Python API processes on the same machine.
    *   Set the `EXTERNAL_DATA_MOUNT` environment variable for the Go Backend to a directory on your local filesystem (e.g., `/Users/yourname/nivai_data` or `/mnt/nivai_data`).
    *   The Python API, running as a local process, will then be able to access these files directly using the paths provided by the Go backend.

Ensure that the user running the Python API process has read permissions for the directories and files created by the Go Backend process within the `EXTERNAL_DATA_MOUNT` location.

---

This README provides a starting point for running the AIFAA platform. Refer to individual service directories for more specific documentation if available.
