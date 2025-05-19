import os
import time
import requests
from doccano_client import DoccanoClient
from doccano_client.models import Project, Label

DOCCANO_BASE_URL = os.getenv("DOCCANO_URL", "http://doccano:8000")
DOCCANO_ADMIN_USER = os.getenv("DOCCANO_ADMIN_USERNAME", "admin")
DOCCANO_ADMIN_PASSWORD = os.getenv("DOCCANO_ADMIN_PASSWORD", "admin")

SAMPLE_DATA_PATH = "/data/sample_data.jsonl"
PROJECT_NAME = "Sample Sentiment Analysis"
PROJECT_DESCRIPTION = "A sample project for text sentiment analysis."
PROJECT_TYPE = "SequenceLabeling" # For text classification
LABELS = [
    Label(text="positive", background_color="#209cee", text_color="#ffffff"),
    Label(text="negative", background_color="#ff3860", text_color="#ffffff"),
    Label(text="neutral", background_color="#f5f5f5", text_color="#0a0a0a"),
]
MAX_RETRIES = 20
RETRY_DELAY = 15

def wait_for_doccano():
    print(f"Waiting for Doccano to be available at {DOCCANO_BASE_URL}...")
    for i in range(MAX_RETRIES):
        try:
            response = requests.get(f"{DOCCANO_BASE_URL}/v1/health", timeout=5)
            if response.status_code == 200:
                print("Doccano is up!")
                return True
        except requests.exceptions.ConnectionError:
            pass
        except requests.exceptions.Timeout:
            print(f"Timeout connecting to Doccano health check (attempt {i+1}/{MAX_RETRIES})")
        print(f"Doccano not ready yet. Retrying in {RETRY_DELAY}s... (Attempt {i+1}/{MAX_RETRIES})")
        time.sleep(RETRY_DELAY)
    print("Max retries reached. Doccano did not become available.")
    return False

def main():
    if not wait_for_doccano():
        exit(1)

    try:
        print(f"Connecting to Doccano at {DOCCANO_BASE_URL} as user {DOCCANO_ADMIN_USER}")
        client = DoccanoClient(DOCCANO_BASE_URL)
        client.login(username=DOCCANO_ADMIN_USER, password=DOCCANO_ADMIN_PASSWORD)
        print("Successfully logged into Doccano.")

        # Check if project exists
        existing_projects = client.list_projects()
        project = next((p for p in existing_projects if p.name == PROJECT_NAME), None)

        if project:
            print(f"Project '{PROJECT_NAME}' already exists with ID: {project.id}. Reusing it.")
            project_id = project.id
            # Optionally, clear existing labels or data if needed, or just add more.
            # For simplicity, we'll assume if project exists, we might not need to re-add labels/data
            # or that re-adding labels is idempotent (depends on doccano-client behavior).
        else:
            print(f"Creating project: {PROJECT_NAME}")
            project = client.create_project(
                Project(name=PROJECT_NAME, description=PROJECT_DESCRIPTION, project_type=PROJECT_TYPE, allow_overlapping=False)
            )
            project_id = project.id
            print(f"Project '{PROJECT_NAME}' created with ID: {project_id}")

            print(f"Creating labels for project ID: {project_id}")
            for label_obj in LABELS:
                try:
                    # Check if label already exists by name for this project
                    # This functionality might not be direct in older client versions, 
                    # so we'll attempt to create and catch potential errors if it exists.
                    client.create_label(project_id=project_id, label=label_obj)
                    print(f"  Label '{label_obj.text}' created.")
                except Exception as le: # Catch a generic exception if label creation fails (e.g. already exists)
                    print(f"  Could not create label '{label_obj.text}' (it might already exist): {le}")

        # Upload data
        # Check if data already exists is harder; Doccano doesn't easily report on uploaded file names.
        # For this script, we'll just attempt to upload. If it's re-run, data might be duplicated unless Doccano handles it.
        print(f"Uploading data from {SAMPLE_DATA_PATH} to project ID: {project_id}")
        try:
            # The `upload` method in doccano-client might vary. Some versions expect file_paths as a list.
            # Some versions might have `doccano_client.beta_api.upload_dataset`
            # Assuming a recent enough version that supports client.upload directly with task type
            client.upload(
                project_id=project_id, 
                file_paths=[SAMPLE_DATA_PATH], 
                format="JSONL", 
                task=PROJECT_TYPE # SequenceLabeling for this project type
            )
            print("Sample data uploaded successfully.")
        except Exception as ue:
            print(f"Error uploading data: {ue}")
            print("Please check if the data format is correct and the file path is accessible within the container.")

    except Exception as e:
        print(f"An error occurred: {e}")
        exit(1)

if __name__ == "__main__":
    main() 