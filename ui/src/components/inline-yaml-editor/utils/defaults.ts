import { FieldSchema } from '../types';

export function getDefaultValue(fieldSchema: FieldSchema): any {
  if (fieldSchema.default !== undefined) {
    return fieldSchema.default;
  }
  
  switch (fieldSchema.type) {
    case "input":
      return "";
    case "number":
      return 0;
    case "bool":
      return false;
    case "select":
      return fieldSchema.options?.[0] || "";
    case "code":
      return "";
    case "key_value":
      return {};
    case "array":
      return [];
    case "object":
      return {};
    case "processor_list":
      return [];
    case "output_cases":
      return [];
    case "output_list":
      return [];
    case "input_list":
      return [];
    case "processor_cases":
      return [];
    default:
      return "";
  }
} 