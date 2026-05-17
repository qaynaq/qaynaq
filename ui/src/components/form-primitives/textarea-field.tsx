import { Textarea } from "@/components/ui/textarea";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";

interface Props extends Omit<FieldWrapperProps, "children"> {
  value: string;
  onChange: (next: string) => void;
  placeholder?: string;
  disabled?: boolean;
  rows?: number;
}

export function TextAreaField({
  value,
  onChange,
  placeholder,
  disabled,
  rows = 3,
  ...wrapper
}: Props) {
  return (
    <FieldWrapper {...wrapper}>
      <Textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        rows={rows}
      />
    </FieldWrapper>
  );
}
