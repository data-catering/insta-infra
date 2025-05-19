import argilla as rg
from argilla_base import init_argilla_client, create_dataset_if_not_exists, ARGILLA_WORKSPACE
import os

try:
    from transformers import pipeline, AutoImageProcessor # AutoModelForImageClassification is often implicitly handled by pipeline
    from PIL import Image
    import requests
    from io import BytesIO
except ImportError:
    print("WARN: Transformers, Pillow, or Requests library not installed. Image classification suggestions will be skipped.")
    pipeline = None
    AutoImageProcessor = None # Though pipeline might load it, good to have for explicit model loading if ever needed
    Image = None
    requests = None
    BytesIO = None

DATASET_NAME = "animal-image-classification"
QUESTION_NAME = "animal_type"
LABELS = {"cat": "Cat", "dog": "Dog", "bird": "Bird", "other": "Other"}
SUGGESTION_AGENT = "deit-tiny-patch16-224-v1"

LOADED_PIPELINE_CACHE = {}

def define_dataset_settings() -> rg.Settings:
    return rg.Settings(
        fields=[
            rg.TextField(name="image_url", title="Image URL", use_markdown=False),
            rg.TextField(name="image_display", title="Image", use_markdown=True)
        ],
        questions=[
            rg.LabelQuestion(
                name=QUESTION_NAME,
                title="What type of animal is in the image?",
                labels=LABELS,
                required=True,
            )
        ],
        guidelines=(
            "Classify the animal in the image.\n\n"
            "- **Cat**: Any feline animal.\n"
            "- **Dog**: Any canine animal.\n"
            "- **Bird**: Any avian animal.\n"
            "- **Other**: Any other animal or if unsure."
        )
    )

def get_image_classification_suggestions(record_data: dict) -> list[rg.Suggestion] | None:
    if not pipeline or not Image or not requests or not BytesIO:
        return None
    try:
        pipe_key = "image-classification-animals"
        if pipe_key not in LOADED_PIPELINE_CACHE:
            print(f"Loading {pipe_key} model (facebook/deit-tiny-patch16-224)...Once per script run.")
            # AutoImageProcessor can be specified if needed, but pipeline often infers it.
            LOADED_PIPELINE_CACHE[pipe_key] = pipeline(
                "image-classification", 
                model="facebook/deit-tiny-patch16-224", 
                # image_processor="facebook/deit-tiny-patch16-224", # often optional
                top_k=None
            )

        image_url = record_data.get("image_url")
        if not image_url: return None

        try:
            response = requests.get(image_url, stream=True, timeout=10)
            response.raise_for_status()
            image = Image.open(BytesIO(response.content))
        except Exception as e:
            print(f"Could not load image from {image_url}: {e}")
            return None
        
        raw_prediction_result = LOADED_PIPELINE_CACHE[pipe_key](image)
        top_prediction = None

        if raw_prediction_result: # Ensure it's not None or empty list
            first_element_of_result = raw_prediction_result[0]
            if isinstance(first_element_of_result, list): # Expected for top_k=None or top_k > 1 (result is List[List[Dict]])
                if first_element_of_result: # Ensure inner list is not empty
                    top_prediction = first_element_of_result[0]
            elif isinstance(first_element_of_result, dict): # Handles case if pipeline returns List[Dict] (e.g. top_k=1)
                top_prediction = first_element_of_result
        
        if not top_prediction: # If no valid prediction could be extracted
            # print(f"No valid top_prediction found for {image_url}") # Optional debug logging
            return None

        model_label_text = top_prediction['label'].lower()
        predicted_value = None

        # Simple keyword mapping to our dataset labels
        if ("cat" in model_label_text or "kitten" in model_label_text) and "cat" in LABELS:
            predicted_value = "cat"
        elif ("dog" in model_label_text or "puppy" in model_label_text or "hound" in model_label_text) and "dog" in LABELS:
            predicted_value = "dog"
        elif ("bird" in model_label_text) and "bird" in LABELS:
            predicted_value = "bird"
        
        if not predicted_value and "other" in LABELS: # Fallback to other
            predicted_value = "other"
        
        if predicted_value:
            return [rg.Suggestion(question_name=QUESTION_NAME, value=predicted_value, score=top_prediction['score'], agent=SUGGESTION_AGENT)]
        return None

    except Exception as e:
        print(f"Error during image classification suggestion: {type(e).__name__} - {str(e)}")
        return None

def prepare_records(client: rg.Argilla) -> list[rg.Record]:
    records_data = [
        {"image_url": "https://picsum.photos/seed/cat1/300/200", "caption": "A fluffy cat."},
        {"image_url": "https://picsum.photos/seed/dog1/300/200", "caption": "A playful dog."},
        {"image_url": "https://picsum.photos/seed/bird1/300/200", "caption": "A colorful bird."},
        {"image_url": "https://picsum.photos/seed/cat2/300/200", "caption": "Another cat sleeping."},
        {"image_url": "https://picsum.photos/seed/animal_other/300/200", "caption": "A generic animal (placeholder for other)."},
        {"image_url": "https://picsum.photos/seed/dog2/300/200", "caption": "A dog running."},
        {"image_url": "https://picsum.photos/seed/cat_dog/300/200", "caption": "A cat and a dog together."},
        {"image_url": "https://picsum.photos/seed/bird_flying/300/200", "caption": "A bird in flight."},
    ]

    records = []
    for data in records_data:
        fields_for_record = {
            "image_url": data["image_url"], 
            "image_display": f"![{data.get('caption', 'Image')}]({data['image_url']})"
        }
        suggestions = get_image_classification_suggestions(data) # Pass original data dict for image_url
        records.append(
            rg.Record(
                fields=fields_for_record,
                suggestions=suggestions if suggestions else [],
            )
        )
    return records

def main():
    print(f"Starting Image Classification data loading for dataset: {DATASET_NAME}")
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
                dataset.records.log(records=records_to_add) # Corrected method
                print(f"Successfully added records to '{DATASET_NAME}'.")
            else:
                print("No records prepared to add.")
    except Exception as e:
        print(f"Error during Image Classification data loading: {e}")

if __name__ == "__main__":
    main() 