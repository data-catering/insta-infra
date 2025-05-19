import argilla as rg
from argilla_base import init_argilla_client, create_dataset_if_not_exists, ARGILLA_WORKSPACE
import os

try:
    from transformers import pipeline, AutoTokenizer, AutoModelForTokenClassification
except ImportError:
    print("WARN: Transformers library not installed. Token classification suggestions will be skipped.")
    pipeline = None
    AutoTokenizer = None
    AutoModelForTokenClassification = None

DATASET_NAME = "ner-example-token-classification"
QUESTION_NAME = "ner_tags"
LABELS = {"PERSON": "Person", "ORG": "Organization", "LOC": "Location", "MISC": "Miscellaneous"}
SUGGESTION_AGENT = "distilbert-conll03-english-v1"

LOADED_PIPELINE_CACHE = {}

def define_dataset_settings() -> rg.Settings:
    return rg.Settings(
        fields=[
            rg.TextField(name="text", title="Text for NER")
        ],
        questions=[
            rg.SpanQuestion(
                name=QUESTION_NAME,
                title="Identify entities in the text.",
                labels=list(LABELS.keys()),
                field="text",
                required=True,
                allow_overlapping=False
            )
        ],
        guidelines=(
            "Annotate the entities in the text by selecting the token(s) and assigning the correct label.\n\n"
            "- **PERSON**: Names of people (e.g., John Doe).\n"
            "- **ORG**: Names of organizations (e.g., Example Corp).\n"
            "- **LOC**: Names of locations (e.g., New York).\n"
            "- **MISC**: Miscellaneous entities (e.g., event names, product names)."
        )
    )

def get_token_classification_suggestions(record_data: dict) -> list[rg.Suggestion] | None:
    if not pipeline or not AutoTokenizer or not AutoModelForTokenClassification:
        return None
    try:
        pipe_key = "token-classification-ner"
        if pipe_key not in LOADED_PIPELINE_CACHE:
            print(f"Loading {pipe_key} model (elastic/distilbert-base-cased-finetuned-conll03-english)...Once per script run.")
            tokenizer = AutoTokenizer.from_pretrained("elastic/distilbert-base-cased-finetuned-conll03-english")
            model = AutoModelForTokenClassification.from_pretrained("elastic/distilbert-base-cased-finetuned-conll03-english")
            LOADED_PIPELINE_CACHE[pipe_key] = pipeline("ner", model=model, tokenizer=tokenizer, grouped_entities=True)
        
        text_to_tag = record_data.get("text", "")
        if not text_to_tag: return None

        raw_predictions = LOADED_PIPELINE_CACHE[pipe_key](text_to_tag)
        suggestions_formatted = []
        model_to_dataset_label_map = {"PER": "PERSON", "ORG": "ORG", "LOC": "LOC", "MISC": "MISC"}

        for entity in raw_predictions:
            model_label_group = entity['entity_group']
            dataset_label = model_to_dataset_label_map.get(model_label_group)
            if dataset_label in LABELS:
                suggestions_formatted.append({
                    "label": dataset_label,
                    "start": entity['start'],
                    "end": entity['end']
                })
        
        if not suggestions_formatted: return None
        return [rg.Suggestion(question_name=QUESTION_NAME, value=suggestions_formatted, agent=SUGGESTION_AGENT)]

    except Exception as e:
        print(f"Error during token classification suggestion: {type(e).__name__} - {str(e)}")
        return None

def prepare_records(client: rg.Argilla) -> list[rg.Record]:
    records_data = [
        {"text": "Elon Musk founded SpaceX in Hawthorne, California."},
        {"text": "Apple Inc. is headquartered in Cupertino."},
        {"text": "The Eiffel Tower is a famous landmark in Paris, France."},
        {"text": "Germany won the World Cup in 2014."},
        {"text": "Dr. Jane Goodall is a renowned primatologist."},
        {"text": "The United Nations has its main office in New York City."},
        {"text": "Microsoft announced a new version of Windows."},
        {"text": "The Amazon River flows through Brazil and Peru."},
        {"text": "Mount Everest is the highest peak in the world."},
        {"text": "The Olympic Games are held every four years."}
    ]

    records = []
    for data in records_data:
        suggestions = get_token_classification_suggestions(data)
        records.append(
            rg.Record(
                fields=data,
                suggestions=suggestions if suggestions else [],
            )
        )
    return records

def main():
    print(f"Starting Token Classification data loading for dataset: {DATASET_NAME}")
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
        print(f"Error during Token Classification data loading: {e}")

if __name__ == "__main__":
    main() 