# Inline YAML Editor - Refactored Structure

## Structure

```
inline-yaml-editor/
├── index.ts                    # Main exports
├── types.ts                    # TypeScript interfaces and types
├── inline-yaml-editor.tsx      # Main component (simplified)
├── utils/
│   └── defaults.ts             # Default value utilities
└── components/
    ├── index.ts                # Component exports
    ├── text-input-field.tsx    # Text input with hover tooltips
    ├── code-editor-field.tsx   # Code editor dialog
    ├── key-value-editor.tsx    # Key-value pair editor
    ├── array-editor.tsx        # Array editor
    ├── object-editor.tsx       # Object editor with nested properties
    ├── nested-property-input.tsx # Input component for object properties
    ├── processor-list-editor.tsx # Processor list with configurations
    ├── output-cases-editor.tsx  # Output cases with conditional logic
    ├── output-list-editor.tsx   # Output list editor
    ├── input-list-editor.tsx    # Input list editor
    ├── processor-cases-editor.tsx # Processor cases with switch logic
    └── nested-processor.tsx     # Individual processor configuration
```

## Completed Components

✅ **Basic Field Types**:
- `TextInputField` - Text input with hover tooltips for long values
- `CodeEditorField` - Modal dialog for code editing with syntax highlighting
- `KeyValueEditor` - Dynamic key-value pair management
- `ArrayEditor` - Array item management with add/remove functionality

✅ **Complex Field Types**:
- `ObjectEditor` - Object editor with nested property support
- `NestedPropertyInput` - Handles different input types for object properties
- `ProcessorListEditor` - Manages lists of processors with configurations
- `OutputCasesEditor` - Conditional output routing with Bloblang queries
- `OutputListEditor` - Multiple output destinations management
- `InputListEditor` - Multiple input sources management  
- `ProcessorCasesEditor` - Switch-like processor routing with conditions
- `NestedProcessor` - Individual processor configuration with flat/structured support

## Advanced Features

- **Preview Mode**: All editors respect the `previewMode` flag for read-only display
- **Internal State Management**: Complex editors maintain internal state for UI operations
- **Bidirectional YAML Conversion**: Seamless conversion between API and UI formats
- **Validation**: Components validate configurations before generating YAML
- **Tooltips**: Long text values show hover tooltips with full content
- **Responsive Design**: Components adapt to different container sizes
- **Error Handling**: Graceful fallbacks for invalid configurations
- **Flat Component Support**: Handles both structured and flat component schemas

## Usage

The component maintains the same API as before:

```tsx
import { InlineYamlEditor } from '@/components/inline-yaml-editor';

<InlineYamlEditor
  schema={schema}
  value={yamlValue}
  onChange={handleChange}
  availableProcessors={processors}
  previewMode={false}
/>
```

All existing imports will continue to work due to the re-export structure. 