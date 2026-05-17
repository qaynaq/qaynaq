import { Switch } from "@/components/ui/switch";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";

interface Props extends Omit<FieldWrapperProps, "children"> {
  checked: boolean;
  onChange: (next: boolean) => void;
  disabled?: boolean;
}

export function SwitchField({
  checked,
  onChange,
  disabled,
  ...wrapper
}: Props) {
  return (
    <FieldWrapper {...wrapper}>
      <div className="flex items-center gap-2">
        <Switch
          checked={checked}
          onCheckedChange={onChange}
          disabled={disabled}
        />
        <span className="text-sm text-muted-foreground">
          {checked ? "Enabled" : "Disabled"}
        </span>
      </div>
    </FieldWrapper>
  );
}
