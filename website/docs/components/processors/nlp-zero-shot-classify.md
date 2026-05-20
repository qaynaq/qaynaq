# NLP Zero-Shot Classify

Runs an ONNX Natural Language Inference (NLI) model to classify text against an arbitrary list of labels chosen at runtime. Unlike regular classification, the model is never trained on the specific labels - it's evaluating whether the input text *entails* each candidate label.

Slower than dedicated classifiers but flexible: change the labels per flow without retraining.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Name | string | required | Identifier for this pipeline instance. Must be unique within the flow |
| Path | string | required | Local path where the model lives. See [Path and download modes](#path-and-download-modes) below |
| Enable Download | boolean | `false` | Fetch the model from HuggingFace into Path at startup |
| Repository | string | - | HuggingFace repository to download from. Required when Enable Download is on |
| ONNX File Path | string | `model.onnx` | Path of the `.onnx` file inside the HuggingFace repository |
| Labels | list of strings | required | Candidate labels to classify against. At least one |
| Multi-Label | boolean | `false` | Score each label independently. Off means scores sum to 1 across labels |
| Hypothesis Template | string | `This example is {}.` | Template used to turn each label into an NLI hypothesis. Must contain `{}` where the label is inserted |

## Path and download modes

The Path field changes meaning based on Enable Download:

- **Download off** - Path points to an existing ONNX model on disk. Either a `.onnx` file directly, or a directory that contains `model.onnx` plus `tokenizer.json`. Qaynaq does not fetch anything; if the files are missing the flow fails at startup.
- **Download on** - Path is the *parent* directory where qaynaq will place downloaded models. Qaynaq creates a sub-directory inside it per repository (e.g. `./models/sentence-transformers_all-MiniLM-L6-v2/`). The parent directory is created if missing. The download runs once; subsequent restarts reuse what's already on disk.

### Sharing models across flows

Different flows that use the **same repository** can safely share the same Path - they both land in the same per-repository sub-directory and reuse the files. Different repositories under the same Path get their own sub-directories. `./models` is a fine default for most setups.

A worker also keeps a content-addressed cache in `~/.cache/huggingface/hub/`, so even unrelated Paths never re-download the same model from HuggingFace.

## Output

```json
{
  "sequence": "I am going to the park",
  "labels": ["fun", "boring", "dangerous"],
  "scores": [0.77, 0.15, 0.08]
}
```

Labels are returned sorted by score, highest first.

## Hypothesis template

The template wraps each candidate label into a sentence the NLI model can score. The default `"This example is {}."` works for adjective-style labels like `positive` or `boring`. For other shapes:

| Label style | Suggested template |
|-------------|--------------------|
| Adjectives (`busy`, `relaxed`) | `This person is {}.` |
| Topics (`sports`, `politics`) | `This text is about {}.` |
| Intents (`refund`, `complaint`) | `The user wants to {}.` |

## Suggested models

| Model | Notes |
|-------|-------|
| `KnightsAnalytics/deberta-v3-base-zeroshot-v1` | Strong general-purpose zero-shot |
| `MoritzLaurer/mDeBERTa-v3-base-mnli-xnli` | Multilingual |
