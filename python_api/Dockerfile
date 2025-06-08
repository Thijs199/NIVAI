# 1. Start from the python:3.9-slim base image
FROM python:3.9-slim

# 2. Install Poetry
RUN pip install poetry

# 3. Set the working directory to /app
WORKDIR /app

# 4. Copy pyproject.toml and poetry.lock files
COPY pyproject.toml poetry.lock /app/

# 5. Install dependencies
RUN poetry install --no-dev --no-interaction --no-ansi

# 6. Copy the src directory
COPY src /app/src/

# 7. Expose port 8081
EXPOSE 8081

# 8. Set the default command
CMD ["poetry", "run", "uvicorn", "src.api.main:app", "--host", "0.0.0.0", "--port", "8081"]
