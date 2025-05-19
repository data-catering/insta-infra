import argilla as rg
from argilla_base import init_argilla_client, create_dataset_if_not_exists, ARGILLA_WORKSPACE
import os

# Try to import ML libraries for this specific script
try:
    from transformers import pipeline
except ImportError:
    print("WARN: Transformers library not installed. Text classification suggestions will be skipped.")
    pipeline = None

DATASET_NAME = "sentiment-reviews-text-classification"
QUESTION_NAME = "sentiment"
LABELS = {"positive": "Positive", "negative": "Negative", "neutral": "Neutral"}
SUGGESTION_AGENT = "distilbert-sst2-english-v1"

# Cache for the model pipeline specific to this script
LOADED_PIPELINE_CACHE = {}

def define_dataset_settings() -> rg.Settings:
    return rg.Settings(
        fields=[
            rg.TextField(name="text", title="Review Text", use_markdown=True)
        ],
        questions=[
            rg.LabelQuestion(
                name=QUESTION_NAME,
                title="What is the sentiment of the text?",
                labels=LABELS,
                required=True,
            )
        ],
        guidelines=(
            "Please classify the sentiment of the review as positive, negative, or neutral.\n"
            "- **Positive**: The review expresses a favorable opinion.\n"
            "- **Negative**: The review expresses an unfavorable opinion.\n"
            "- **Neutral**: The review is objective or does not express a strong opinion."
        )
    )

def get_text_classification_suggestions(record_data: dict) -> list[rg.Suggestion] | None:
    if not pipeline: # If transformers is not installed
        return None
    try:
        pipe_key = "text-classification-sentiment"
        if pipe_key not in LOADED_PIPELINE_CACHE:
            print(f"Loading {pipe_key} model (distilbert-base-uncased-finetuned-sst-2-english)...Once per script run.")
            LOADED_PIPELINE_CACHE[pipe_key] = pipeline(
                "sentiment-analysis", 
                model="distilbert-base-uncased-finetuned-sst-2-english", 
                tokenizer="distilbert-base-uncased-finetuned-sst-2-english", 
                top_k=None # Get all scores
            )
        
        text_to_classify = record_data.get("text", "")
        if not text_to_classify: return None

        raw_predictions = LOADED_PIPELINE_CACHE[pipe_key](text_to_classify)
        
        # Model output is like [[{'label': 'POSITIVE', 'score': 0.999}, {'label': 'NEGATIVE', 'score': 0.001}]]
        # Map model output to our dataset labels
        # The `labels` for the Question is a dict: {"positive": "Positive", ...}
        # The suggestion value should be the key (e.g., "positive")
        if raw_predictions and raw_predictions[0]:
            top_prediction = raw_predictions[0][0] # Highest scoring prediction from the model
            model_label = top_prediction['label'].upper()
            # Map to our internal label keys
            if model_label == "POSITIVE" and "positive" in LABELS:
                predicted_value = "positive"
            elif model_label == "NEGATIVE" and "negative" in LABELS:
                predicted_value = "negative"
            else:
                # Fallback or if model produces labels not directly in our simple map
                # For this model, it's usually POSITIVE/NEGATIVE. We map to neutral if not directly matched and neutral exists.
                predicted_value = "neutral" if "neutral" in LABELS else None 

            if predicted_value:
                return [rg.Suggestion(question_name=QUESTION_NAME, value=predicted_value, score=top_prediction['score'], agent=SUGGESTION_AGENT)]
        return None

    except Exception as e:
        print(f"Error during text classification suggestion: {e}")
        return None

def prepare_records(client: rg.Argilla) -> list[rg.Record]:
    records_data = [
        {"text": "I love this product! It exceeded all my expectations."},
        {"text": "The quality is not as good as I hoped. Quite disappointing."},
        {"text": "This is an average product. It does the job, nothing more, nothing less."},
        {"text": "Absolutely fantastic! I would recommend this to everyone."},
        {"text": "I had a terrible experience with their customer service."},
        {"text": "The product arrived on time and was well-packaged."},
        {"text": "It is okay, but I have seen better for this price range."},
        {"text": "What a wonderful purchase! So happy with it."},
        {"text": "Completely useless, a waste of money."},
        {"text": "The features are exactly as described in the manual."}
    ]

    records = []
    for data in records_data:
        suggestions = get_text_classification_suggestions(data)
        records.append(
            rg.Record(
                fields=data,
                suggestions=suggestions if suggestions else [],
            )
        )
    return records

def main():
    print(f"Starting Text Classification data loading for dataset: {DATASET_NAME}")
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
        print(f"Error during Text Classification data loading: {e}")
        # Consider if exit(1) is appropriate or if the main script (run_all_argilla_loaders.sh) should continue

if __name__ == "__main__":
    main() 