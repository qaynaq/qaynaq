import { Suspense, lazy } from "react";
import { FieldWrapper, type FieldWrapperProps } from "./field-wrapper";

const MonacoInner = lazy(() => import("./code-field.inner"));

interface Props extends Omit<FieldWrapperProps, "children"> {
  value: string;
  onChange: (next: string) => void;
  language?: string;
  height?: string;
}

export function CodeField({
  value,
  onChange,
  language = "bloblang",
  height = "150px",
  ...wrapper
}: Props) {
  return (
    <FieldWrapper {...wrapper}>
      <Suspense
        fallback={
          <div className="h-[150px] rounded-md border bg-muted/30 flex items-center justify-center text-xs text-muted-foreground">
            Loading editor...
          </div>
        }
      >
        <MonacoInner
          value={value}
          onChange={onChange}
          language={language}
          height={height}
        />
      </Suspense>
    </FieldWrapper>
  );
}
