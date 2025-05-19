import argilla as rg
import os
import time

# Configuration from environment variables
ARGILLA_API_URL = os.getenv("ARGILLA_API_URL", "http://argilla:6900")
ARGILLA_API_KEY = os.getenv("ARGILLA_API_KEY", "argilla.apikey") # Defaulting to user's set API key
ARGILLA_WORKSPACE = os.getenv("ARGILLA_WORKSPACE", "default")

MAX_RETRIES = int(os.getenv("ARGILLA_CONNECT_MAX_RETRIES", "10"))
RETRY_DELAY_SECONDS = int(os.getenv("ARGILLA_CONNECT_RETRY_DELAY", "10"))

def init_argilla_client() -> rg.Argilla | None:
    """Initializes and returns an Argilla client, retrying on failure."""
    for attempt in range(MAX_RETRIES):
        try:
            print(f"Attempting to connect to Argilla: {ARGILLA_API_URL}, attempt {attempt + 1}/{MAX_RETRIES}")
            client = rg.Argilla(api_url=ARGILLA_API_URL, api_key=ARGILLA_API_KEY)
            client.workspaces(ARGILLA_WORKSPACE) 
            print(f"Successfully connected to Argilla and accessed workspace '{ARGILLA_WORKSPACE}'.")
            return client
        except Exception as e:
            print(f"Connection or workspace access failed: {e}")
            if attempt < MAX_RETRIES - 1:
                print(f"Retrying in {RETRY_DELAY_SECONDS} seconds...")
                time.sleep(RETRY_DELAY_SECONDS)
            else:
                print("Max retries reached. Could not connect to Argilla or access workspace.")
    return None

def create_dataset_if_not_exists(client: rg.Argilla, dataset_name: str, settings: rg.Settings) -> rg.Dataset:
    """Creates a dataset with the given name and settings if it doesn't already exist in the workspace."""
    try:
        workspace = client.workspaces(ARGILLA_WORKSPACE)
        for ds_in_workspace in workspace.datasets:
            if ds_in_workspace.name == dataset_name:
                print(f"Dataset '{dataset_name}' already exists in workspace '{ARGILLA_WORKSPACE}'. Returning existing dataset.")
                return rg.Dataset.from_name(name=dataset_name, workspace=ARGILLA_WORKSPACE)

        print(f"Creating dataset: '{dataset_name}' in workspace '{ARGILLA_WORKSPACE}'")
        dataset = rg.Dataset(
            name=dataset_name,
            settings=settings,
            workspace=workspace
        )
        dataset.create()
        print(f"Dataset '{dataset_name}' created successfully in workspace '{ARGILLA_WORKSPACE}'.")
        return dataset
    except Exception as e:
        print(f"Error during dataset check or creation for '{dataset_name}': {e}")
        raise 