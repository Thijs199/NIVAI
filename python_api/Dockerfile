# 1. Start from the python:3.13-slim base image
FROM python:3.13-slim

# 2. Install Poetry
RUN pip install poetry

# 3. Set the working directory to /app
WORKDIR /app

# 4. Copy pyproject.toml and poetry.lock files
COPY pyproject.toml poetry.lock /app/

# 5. Configure Poetry to not create virtual environment (we're in a container)
RUN poetry config virtualenvs.create false

# 6. Install only dependencies (not the project itself)
RUN poetry install --only main --no-interaction --no-ansi --no-root

# 7. Copy the src directory
COPY src /app/src/

# 8. Install the project
RUN poetry install --only main --no-interaction --no-ansi

# 9. Expose port 8081
EXPOSE 8081

# 10. Set the default command
CMD ["poetry", "run", "uvicorn", "src.api.main:app", "--host", "0.0.0.0", "--port", "8081"]
