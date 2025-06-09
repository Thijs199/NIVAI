# AIFAA Football Analytics Platform

This document provides instructions on how to set up, configure, and run the various components of the AIFAA Football Analytics Platform locally for development and testing.

The platform consists of three main services:
1.  **Go Backend:** Handles API requests, user authentication, video/data uploads, and orchestrates calls to the Python API.
2.  **Python API:** Performs intensive physical statistics calculations on tracking and event data.
3.  **Next.js Frontend:** Provides the user interface for uploading data, viewing matches, and analyzing dashboards.

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
