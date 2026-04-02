import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { componentSchemas } from "@/lib/component-schemas";

interface BufferFormProps {
  initialData?: {
    label: string;
    component: string;
    config: any;
  };
  onSubmit: (data: { label: string; component: string; config: any }) => void;
  onCancel: () => void;
}

export function BufferForm({
  initialData,
  onSubmit,
  onCancel,
}: BufferFormProps) {
  const [label, setLabel] = useState(initialData?.label || "");
  const [selectedComponent, setSelectedComponent] = useState(
    initialData?.component || "",
  );
  const [config, setConfig] = useState<any>(initialData?.config || {});

  const bufferGroups = [
    {
      label: "Windowing",
      components: ["system_window"],
    },
    {
      label: "Utility",
      components: ["memory", "sqlite"],
    },
  ];
  const selectedSchema =
    selectedComponent && componentSchemas.buffer
      ? componentSchemas.buffer[
          selectedComponent as keyof typeof componentSchemas.buffer
        ]
      : null;

  useEffect(() => {
    if (selectedComponent && !initialData) {
      const schema =
        componentSchemas.buffer?.[
          selectedComponent as keyof typeof componentSchemas.buffer
        ];
      if (schema?.properties) {
        const defaultConfig: any = {};
        Object.entries(schema.properties).forEach(
          ([key, prop]: [string, any]) => {
            if (prop.default !== undefined) {
              defaultConfig[key] = prop.default;
            }
          },
        );
        setConfig(defaultConfig);
      }
    }
  }, [selectedComponent, initialData]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      label,
      component: selectedComponent,
      config,
    });
  };

  const renderField = (key: string, field: any) => {
    const value = config[key];

    const updateValue = (newValue: any) => {
      setConfig({ ...config, [key]: newValue });
    };

    switch (field.type) {
      case "number":
        return (
          <div key={key} className="space-y-2 mb-4">
            <Label htmlFor={key}>
              {field.title || key}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {field.description && (
              <p className="text-sm text-muted-foreground">
                {field.description}
              </p>
            )}
            <Input
              id={key}
              type="number"
              min={field.min || 0}
              value={value !== undefined ? value : field.default || ""}
              onChange={(e) => updateValue(Number(e.target.value))}
            />
          </div>
        );

      case "select":
        return (
          <div key={key} className="space-y-2 mb-4">
            <Label htmlFor={key}>
              {field.title || key}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {field.description && (
              <p className="text-sm text-muted-foreground">
                {field.description}
              </p>
            )}
            <Select
              value={value || field.default || ""}
              onValueChange={(val) => updateValue(val)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select an option" />
              </SelectTrigger>
              <SelectContent>
                {field.options?.map((option: string) => (
                  <SelectItem key={option} value={option}>
                    {option}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        );

      case "textarea":
        return (
          <div key={key} className="space-y-2 mb-4">
            <Label htmlFor={key}>
              {field.title || key}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {field.description && (
              <p className="text-sm text-muted-foreground">
                {field.description}
              </p>
            )}
            <Textarea
              id={key}
              rows={4}
              value={value !== undefined ? value : field.default || ""}
              onChange={(e) => updateValue(e.target.value)}
              className="font-mono text-sm"
            />
          </div>
        );

      case "object":
        return (
          <div key={key} className="space-y-2 mb-4">
            <Label>
              {field.title || key}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {field.description && (
              <p className="text-sm text-muted-foreground mb-2">
                {field.description}
              </p>
            )}
            <div className="border rounded-lg p-4 space-y-4 bg-muted/30">
              {field.properties &&
                Object.entries(field.properties).map(
                  ([nestedKey, nestedField]: [string, any]) => {
                    const nestedValue = value?.[nestedKey];
                    const updateNestedValue = (newValue: any) => {
                      setConfig({
                        ...config,
                        [key]: {
                          ...(config[key] || {}),
                          [nestedKey]: newValue,
                        },
                      });
                    };

                    if (nestedField.type === "number") {
                      return (
                        <div key={nestedKey} className="space-y-2">
                          <Label htmlFor={`${key}-${nestedKey}`}>
                            {nestedField.title || nestedKey}
                          </Label>
                          {nestedField.description && (
                            <p className="text-sm text-muted-foreground">
                              {nestedField.description}
                            </p>
                          )}
                          <Input
                            id={`${key}-${nestedKey}`}
                            type="number"
                            min={nestedField.min || 0}
                            value={
                              nestedValue !== undefined
                                ? nestedValue
                                : nestedField.default || ""
                            }
                            onChange={(e) =>
                              updateNestedValue(Number(e.target.value))
                            }
                          />
                        </div>
                      );
                    } else if (nestedField.type === "select") {
                      return (
                        <div key={nestedKey} className="space-y-2">
                          <Label htmlFor={`${key}-${nestedKey}`}>
                            {nestedField.title || nestedKey}
                          </Label>
                          {nestedField.description && (
                            <p className="text-sm text-muted-foreground">
                              {nestedField.description}
                            </p>
                          )}
                          <Select
                            value={nestedValue || nestedField.default || ""}
                            onValueChange={(val) => updateNestedValue(val)}
                          >
                            <SelectTrigger>
                              <SelectValue placeholder="Select an option" />
                            </SelectTrigger>
                            <SelectContent>
                              {nestedField.options?.map((option: string) => (
                                <SelectItem key={option} value={option}>
                                  {option}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                      );
                    } else if (nestedField.type === "textarea") {
                      return (
                        <div key={nestedKey} className="space-y-2">
                          <Label htmlFor={`${key}-${nestedKey}`}>
                            {nestedField.title || nestedKey}
                          </Label>
                          {nestedField.description && (
                            <p className="text-sm text-muted-foreground">
                              {nestedField.description}
                            </p>
                          )}
                          <Textarea
                            id={`${key}-${nestedKey}`}
                            rows={3}
                            value={
                              nestedValue !== undefined
                                ? nestedValue
                                : nestedField.default || ""
                            }
                            onChange={(e) => updateNestedValue(e.target.value)}
                            className="font-mono text-sm"
                          />
                        </div>
                      );
                    } else {
                      return (
                        <div key={nestedKey} className="space-y-2">
                          <Label htmlFor={`${key}-${nestedKey}`}>
                            {nestedField.title || nestedKey}
                          </Label>
                          {nestedField.description && (
                            <p className="text-sm text-muted-foreground">
                              {nestedField.description}
                            </p>
                          )}
                          <Input
                            id={`${key}-${nestedKey}`}
                            type="text"
                            value={
                              nestedValue !== undefined
                                ? nestedValue
                                : nestedField.default || ""
                            }
                            onChange={(e) => updateNestedValue(e.target.value)}
                          />
                        </div>
                      );
                    }
                  },
                )}
            </div>
          </div>
        );

      case "string":
      default:
        if (field.options) {
          return (
            <div key={key} className="space-y-2 mb-4">
              <Label htmlFor={key}>
                {field.title || key}
                {field.required && <span className="text-red-500 ml-1">*</span>}
              </Label>
              {field.description && (
                <p className="text-sm text-muted-foreground">
                  {field.description}
                </p>
              )}
              <Select
                value={value || field.default || ""}
                onValueChange={(val) => updateValue(val)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select an option" />
                </SelectTrigger>
                <SelectContent>
                  {field.options.map((option: string) => (
                    <SelectItem key={option} value={option}>
                      {option}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          );
        }

        return (
          <div key={key} className="space-y-2 mb-4">
            <Label htmlFor={key}>
              {field.title || key}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {field.description && (
              <p className="text-sm text-muted-foreground">
                {field.description}
              </p>
            )}
            <Input
              id={key}
              type="text"
              value={value !== undefined ? value : field.default || ""}
              onChange={(e) => updateValue(e.target.value)}
            />
          </div>
        );
    }
  };

  return (
    <Card className="w-full max-w-2xl">
      <CardHeader>
        <CardTitle>{initialData ? "Edit Buffer" : "Add New Buffer"}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="buffer-label">
              Label <span className="text-red-500">*</span>
            </Label>
            <Input
              id="buffer-label"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              placeholder="my_buffer"
              required
            />
            <p className="text-sm text-muted-foreground">
              A unique identifier for this buffer resource
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="buffer-component">
              Buffer Type <span className="text-red-500">*</span>
            </Label>
            <Select
              value={selectedComponent}
              onValueChange={setSelectedComponent}
              required
            >
              <SelectTrigger>
                <SelectValue placeholder="Select a buffer type" />
              </SelectTrigger>
              <SelectContent>
                {bufferGroups.map((group) => (
                  <SelectGroup key={group.label}>
                    <SelectLabel>{group.label}</SelectLabel>
                    {group.components.map((component) => (
                      <SelectItem key={component} value={component}>
                        {componentSchemas.buffer?.[
                          component as keyof typeof componentSchemas.buffer
                        ]?.title || component}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                ))}
              </SelectContent>
            </Select>
            {selectedSchema?.description && (
              <p className="text-sm text-muted-foreground">
                {selectedSchema.description}
              </p>
            )}
          </div>

          {selectedComponent && selectedSchema && selectedSchema.properties && (
            <div className="border rounded-lg p-4 space-y-4">
              <h3 className="font-semibold text-lg">Configuration</h3>
              {Object.entries(selectedSchema.properties).map(([key, field]) =>
                renderField(key, field),
              )}
            </div>
          )}

          <div className="flex justify-end space-x-2 pt-4">
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled={!label || !selectedComponent}>
              {initialData ? "Update Buffer" : "Create Buffer"}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
