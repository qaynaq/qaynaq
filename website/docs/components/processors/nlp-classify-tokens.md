# NLP Classify Tokens

Runs an ONNX token classification model against the message and replaces its content with the list of detected token-level labels. The most common use is named entity recognition (people, organizations, locations).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Name | string | required | Identifier for this pipeline instance. Must be unique within the flow |
| Path | string | required | Local path where the model lives. See [Path and download modes](#path-and-download-modes) below |
| Enable Download | boolean | `false` | Fetch the model from HuggingFace into Path at startup |
| Repository | string | - | HuggingFace repository to download from. Required when Enable Download is on |
| ONNX File Path | string | `model.onnx` | Path of the `.onnx` file inside the HuggingFace repository |
| Aggregation Strategy | enum | `SIMPLE` | `SIMPLE` merges adjacent tokens sharing an entity label. `NONE` keeps every token separate |
| Ignore Labels | list of strings | `[]` | Labels to drop from output. Common values are `O` (the BIO scheme's outside tag) and `MISC` |

## Path and download modes

The Path field changes meaning based on Enable Download:

- **Download off** - Path points to an existing ONNX model on disk. Either a `.onnx` file directly, or a directory that contains `model.onnx` plus `tokenizer.json`. Qaynaq does not fetch anything; if the files are missing the flow fails at startup.
- **Download on** - Path is the *parent* directory where qaynaq will place downloaded models. Qaynaq creates a sub-directory inside it per repository (e.g. `./models/sentence-transformers_all-MiniLM-L6-v2/`). The parent directory is created if missing. The download runs once; subsequent restarts reuse what's already on disk.

### Sharing models across flows

Different flows that use the **same repository** can safely share the same Path - they both land in the same per-repository sub-directory and reuse the files. Different repositories under the same Path get their own sub-directories. `./models` is a fine default for most setups.

A worker also keeps a content-addressed cache in `~/.cache/huggingface/hub/`, so even unrelated Paths never re-download the same model from HuggingFace.

## Output

A JSON array of entity objects. With `SIMPLE` aggregation:

```json
[
  {"Entity": "PER", "Score": 0.997, "Word": "John", "Start": 0, "End": 4},
  {"Entity": "ORG", "Score": 0.985, "Word": "Apple Inc.", "Start": 14, "End": 24}
]
```

With `NONE`, you get one entry per token (including BIO prefixes like `B-PER`, `I-PER`).

## Suggested models

| Model | Tags |
|-------|------|
| `KnightsAnalytics/distilbert-NER` | PER, ORG, LOC, MISC |
| `Davlan/bert-base-multilingual-cased-ner-hrl` | Multilingual NER |
