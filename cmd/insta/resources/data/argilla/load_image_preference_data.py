import argilla as rg
from argilla_base import init_argilla_client, create_dataset_if_not_exists, ARGILLA_WORKSPACE
import os

DATASET_NAME = "landscape-image-preference"
# No suggestion logic for this one yet, so no QUESTION_NAME or LABELS for suggestions needed here for now.

def define_dataset_settings() -> rg.Settings:
    return rg.Settings(
        fields=[
            rg.TextField(name="image_url_1", title="Image URL 1", use_markdown=False),
            rg.TextField(name="image_url_2", title="Image URL 2", use_markdown=False),
            rg.TextField(name="image_display_1", title="Image 1", use_markdown=True),
            rg.TextField(name="image_display_2", title="Image 2", use_markdown=True),
            rg.TextField(name="prompt", title="Optional Prompt/Context", use_markdown=True)
        ],
        questions=[
            rg.RankingQuestion(
                name="preference_ranking",
                title="Which image do you prefer for the given prompt? Or rank them.",
                values=["Image 1", "Image 2"],
                required=True
            ),
            rg.LabelQuestion(
                name="choice",
                title="Alternatively, which one do you choose? (Image 1 or Image 2)",
                labels={"img1": "Image 1", "img2": "Image 2", "neither": "Neither is good"},
                required=False # Making this optional, ranking is primary
            ),
            rg.TextQuestion(
                name="preference_reason",
                title="Why did you prefer this image (or rank them this way)?",
                required=False
            )
        ],
        guidelines=(
            "You will be shown two images (Image 1 and Image 2) and an optional prompt.\n\n"
            "1. **Ranking**: Drag and drop the images in order of preference (most preferred at the top).\n"
            "2. **Choice (Optional)**: If you have a clear single choice, select it.\n"
            "3. **Reason (Optional)**: Briefly explain your preference or ranking.\n\n"
            "Consider aspects like relevance to the prompt, image quality, aesthetics, and composition."
        )
    )

def prepare_records(client: rg.Argilla) -> list[rg.Record]:
    # Using picsum for varied placeholder landscape images
    records_data = [
        {
            "image_url_1": "https://picsum.photos/seed/landscape1a/400/300",
            "image_url_2": "https://picsum.photos/seed/landscape1b/400/300",
            "prompt": "A serene mountain lake at dawn."
        },
        {
            "image_url_1": "https://picsum.photos/seed/landscape2a/400/300",
            "image_url_2": "https://picsum.photos/seed/landscape2b/400/300",
            "prompt": "A bustling cityscape at night."
        },
        {
            "image_url_1": "https://picsum.photos/seed/landscape3a/400/300",
            "image_url_2": "https://picsum.photos/seed/landscape3b/400/300",
            "prompt": "A dense, misty forest trail."
        },
        {
            "image_url_1": "https://picsum.photos/seed/landscape4a/400/300",
            "image_url_2": "https://picsum.photos/seed/landscape4b/400/300",
            "prompt": "Abstract representation of tranquility."
        },
    ]

    records = []
    for data in records_data:
        fields_for_record = {
            "image_url_1": data["image_url_1"],
            "image_url_2": data["image_url_2"],
            "image_display_1": f"![Image 1: {data.get('prompt','')}]({data['image_url_1']})",
            "image_display_2": f"![Image 2: {data.get('prompt','')}]({data['image_url_2']})",
            "prompt": data.get("prompt", "N/A")
        }
        # No suggestions for preference tasks in this iteration
        records.append(rg.Record(fields=fields_for_record))
    return records

def main():
    print(f"Starting Image Preference data loading for dataset: {DATASET_NAME}")
    client = init_argilla_client()
    if not client:
        print("Failed to initialize Argilla client. Exiting.")
        return
    
    try:
        settings = define_dataset_settings()
        dataset = create_dataset_if_not_exists(client, DATASET_NAME, settings)

        if dataset:
            print(f"Preparing records for dataset '{DATASET_NAME}'...")
            records_to_add = prepare_records(client)
            if records_to_add:
                print(f"Adding {len(records_to_add)} records to dataset '{DATASET_NAME}'...")
                dataset.records.log(records=records_to_add)
                print(f"Successfully added records to '{DATASET_NAME}'.")
            else:
                print("No records prepared to add.")
    except Exception as e:
        print(f"Error during Image Preference data loading: {e}")

if __name__ == "__main__":
    main() 