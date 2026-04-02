export interface FieldSchema {
  type:
    | "input"
    | "number"
    | "bool"
    | "select"
    | "dynamic_select"
    | "code"
    | "key_value"
    | "array"
    | "object"
    | "processor_list"
    | "output_cases"
    | "output_list"
    | "input_list"
    | "processor_cases"
    | "file"
    | "property_list";
  title: string;
  description?: string;
  required?: boolean;
  default?: any;
  options?: string[];
  pattern?: string;
  patternMessage?: string;
  dataSource?: "caches" | "connections" | "secrets" | "rate_limits";
  properties?: Record<string, FieldSchema>;
}

export interface ComponentSchema {
  title?: string;
  flat?: boolean;
  properties?: Record<string, FieldSchema>;
  [key: string]: any;
}

export interface ProcessorComponentSchema {
  id: string;
  name: string;
  component: string;
  type: "input" | "processor" | "output";
  schema?: ComponentSchema;
}

export interface InlineYamlEditorProps {
  schema: ComponentSchema;
  value?: string;
  onChange: (yamlValue: string) => void;
  availableProcessors?: ProcessorComponentSchema[];
  availableInputs?: ProcessorComponentSchema[];
  availableOutputs?: ProcessorComponentSchema[];
  previewMode?: boolean;
}

export interface FieldState {
  enabled: boolean;
  value: any;
}

export interface EditorProps {
  value: any;
  updateValue: (value: any) => void;
  previewMode?: boolean;
  availableProcessors?: ProcessorComponentSchema[];
  availableInputs?: ProcessorComponentSchema[];
  availableOutputs?: ProcessorComponentSchema[];
}
