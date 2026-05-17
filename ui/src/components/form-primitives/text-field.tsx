import { Input } from "@/components/ui/input";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";
import { cn } from "@/lib/utils";

interface Props extends Omit<FieldWrapperProps, "children"> {
  value: string;
  onChange: (next: string) => void;
  placeholder?: string;
  disabled?: boolean;
  type?: "text" | "password" | "email" | "url";
}

export function TextField({
  value,
  onChange,
  placeholder,
  disabled,
  type = "text",
  size,
  ...wrapper
}: Props) {
  return (
    <FieldWrapper size={size} {...wrapper}>
      <Input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        className={cn(size === "sm" && "h-8 text-sm")}
      />
    </FieldWrapper>
  );
}
