# NLP Classify Text

Runs an ONNX text classification model against the message and replaces its content with the predicted label(s) and scores. Use for sentiment analysis, intent detection, topic classification, or grammatical correctness checks.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Name | string | required | Identifier for this pipeline instance. Must be unique within the flow |
| Path | string | required | Local path where the model lives. See [Path and download modes](#path-and-download-modes) below |
| Enable Download | boolean | `false` | Fetch the model from HuggingFace into Path at startup |
| Repository | string | - | HuggingFace repository to download from. Required when Enable Download is on |
| ONNX File Path | string | `model.onnx` | Path of the `.onnx` file inside the HuggingFace repository |
| Aggregation Function | enum | `SOFTMAX` | `SOFTMAX` for mutually exclusive labels (probabilities sum to 1). `SIGMOID` for independent labels |
| Multi-Label | boolean | `false` | Return scores for every label. When off, only the top label is emitted |

## Path and download modes

The Path field changes meaning based on Enable Download:

- **Download off** - Path points to an existing ONNX model on disk. Either a `.onnx` file directly, or a directory that contains `model.onnx` plus `tokenizer.json`. Qaynaq does not fetch anything; if the files are missing the flow fails at startup.
- **Download on** - Path is the *parent* directory where qaynaq will place downloaded models. Qaynaq creates a sub-directory inside it per repository (e.g. `./models/sentence-transformers_all-MiniLM-L6-v2/`). The parent directory is created if missing. The download runs once; subsequent restarts reuse what's already on disk.

### Sharing models across flows

Different flows that use the **same repository** can safely share the same Path - they both land in the same per-repository sub-directory and reuse the files. Different repositories under the same Path get their own sub-directories. `./models` is a fine default for most setups.

A worker also keeps a content-addressed cache in `~/.cache/huggingface/hub/`, so even unrelated Paths never re-download the same model from HuggingFace.

## Output

A JSON array of `{Label, Score}` objects. With Multi-Label off, the array contains one entry. With it on, the array contains every label.

## Suggested models

| Model | Task |
|-------|------|
| `KnightsAnalytics/distilbert-base-uncased-finetuned-sst-2-english` | Sentiment (positive/negative) |
| `Cohee/distilbert-base-uncased-go-emotions-onnx` | Fine-grained emotion |
| `KnightsAnalytics/roberta-base-go_emotions-onnx` | Multi-label emotion |
