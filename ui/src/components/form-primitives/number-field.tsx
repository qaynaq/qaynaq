import { Input } from "@/components/ui/input";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";
import { cn } from "@/lib/utils";

interface Props extends Omit<FieldWrapperProps, "children"> {
  value: number;
  onChange: (next: number) => void;
  min?: number;
  max?: number;
  step?: number;
  disabled?: boolean;
  placeholder?: string;
}

export function NumberField({
  value,
  onChange,
  min,
  max,
  step,
  disabled,
  placeholder,
  size,
  ...wrapper
}: Props) {
  return (
    <FieldWrapper size={size} {...wrapper}>
      <Input
        type="number"
        value={Number.isFinite(value) ? value : ""}
        min={min}
        max={max}
        step={step}
        disabled={disabled}
        placeholder={placeholder}
        className={cn(size === "sm" && "h-8 text-sm")}
        onChange={(e) => {
          const raw = e.target.value;
          if (raw === "") {
            onChange(0);
            return;
          }
          const n = Number(raw);
          if (Number.isFinite(n)) onChange(n);
        }}
      />
    </FieldWrapper>
  );
}
