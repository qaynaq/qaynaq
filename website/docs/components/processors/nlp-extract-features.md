# NLP Extract Features

Runs an ONNX feature extraction model against the message and replaces its content with the resulting embedding vector. This is the embedding step in a RAG pipeline: combine it with the [RAG Chunker](/docs/components/scanners/rag-chunker) scanner and a vector store output.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Name | string | required | Identifier for this pipeline instance. Must be unique within the flow |
| Path | string | required | Local path where the model lives. See [Path and download modes](#path-and-download-modes) below |
| Enable Download | boolean | `false` | Fetch the model from HuggingFace into Path at startup |
| Repository | string | - | HuggingFace repository to download from. Required when Enable Download is on |
| ONNX File Path | string | `model.onnx` | Path of the `.onnx` file inside the HuggingFace repository (not on your disk). Change only when the repo publishes multiple variants like `onnx/model_quantized.onnx` |
| Normalization | boolean | `false` | L2-normalize the output vector. Required by most cosine-similarity vector stores |

## Path and download modes

The Path field changes meaning based on Enable Download:

- **Download off** - Path points to an existing ONNX model on disk. Either a `.onnx` file directly, or a directory that contains `model.onnx` plus `tokenizer.json`. Qaynaq does not fetch anything; if the files are missing the flow fails at startup.
- **Download on** - Path is the *parent* directory where qaynaq will place downloaded models. Qaynaq creates a sub-directory inside it per repository (e.g. `./models/sentence-transformers_all-MiniLM-L6-v2/`). The parent directory is created if missing. The download runs once; subsequent restarts reuse what's already on disk.

### Sharing models across flows

Different flows that use the **same repository** can safely share the same Path - they both land in the same per-repository sub-directory and reuse the files. Different repositories under the same Path get their own sub-directories. `./models` is a fine default for most setups.

A worker also keeps a content-addressed cache in `~/.cache/huggingface/hub/`, so even unrelated Paths never re-download the same model from HuggingFace.

## Output

The message body is replaced with a float64 vector of the model's output dimension. For `all-MiniLM-L6-v2` this is a 384-element array.

Metadata added to each message:

| Field | Description |
|-------|-------------|
| `pipeline_name` | The Name field above |
| `output_type` | Type of the model output tensor |
| `output_shape` | Shape of the output tensor |

## Suggested models

| Model | Dimensions | Use case |
|-------|------------|----------|
| `sentence-transformers/all-MiniLM-L6-v2` | 384 | Small, fast, general-purpose |
| `sentence-transformers/all-mpnet-base-v2` | 768 | Higher quality, slower |
| `BAAI/bge-small-en-v1.5` | 384 | Strong on retrieval benchmarks |
