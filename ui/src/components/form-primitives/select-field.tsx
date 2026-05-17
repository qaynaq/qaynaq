import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";
import { cn } from "@/lib/utils";

export interface SelectOption {
  value: string;
  label?: string;
}

interface Props extends Omit<FieldWrapperProps, "children"> {
  value: string;
  onChange: (next: string) => void;
  options: ReadonlyArray<string | SelectOption>;
  placeholder?: string;
  disabled?: boolean;
}

export function SelectField({
  value,
  onChange,
  options,
  placeholder,
  disabled,
  size,
  ...wrapper
}: Props) {
  const items = options.map((o) =>
    typeof o === "string" ? { value: o, label: o } : { ...o, label: o.label ?? o.value },
  );
  return (
    <FieldWrapper size={size} {...wrapper}>
      <Select value={value} onValueChange={onChange} disabled={disabled}>
        <SelectTrigger className={cn(size === "sm" && "h-8 text-sm")}>
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent>
          {items.map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </FieldWrapper>
  );
}
