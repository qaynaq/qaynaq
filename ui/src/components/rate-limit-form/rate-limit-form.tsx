import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { componentSchemas, componentLists } from "@/lib/component-schemas";

interface RateLimitFormProps {
  initialData?: {
    label: string;
    component: string;
    config: any;
  };
  onSubmit: (data: { label: string; component: string; config: any }) => void;
  onCancel: () => void;
}

export function RateLimitForm({
  initialData,
  onSubmit,
  onCancel,
}: RateLimitFormProps) {
  const [label, setLabel] = useState(initialData?.label || "");
  const [selectedComponent, setSelectedComponent] = useState(
    initialData?.component || "",
  );
  const [config, setConfig] = useState<any>(initialData?.config || {});

  const rateLimitComponents = componentLists.rate_limit || [];
  const selectedSchema =
    selectedComponent && componentSchemas.rate_limit
      ? componentSchemas.rate_limit[
          selectedComponent as keyof typeof componentSchemas.rate_limit
        ]
      : null;

  useEffect(() => {
    if (selectedComponent && !initialData) {
      const schema =
        componentSchemas.rate_limit?.[
          selectedComponent as keyof typeof componentSchemas.rate_limit
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
        <CardTitle>
          {initialData ? "Edit Rate Limit" : "Add New Rate Limit"}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="rate-limit-label">
              Label <span className="text-red-500">*</span>
            </Label>
            <Input
              id="rate-limit-label"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              placeholder="shopify_api"
              required
            />
            <p className="text-sm text-muted-foreground">
              A unique identifier for this rate limit resource
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="rate-limit-component">
              Rate Limit Type <span className="text-red-500">*</span>
            </Label>
            <Select
              value={selectedComponent}
              onValueChange={setSelectedComponent}
              required
            >
              <SelectTrigger>
                <SelectValue placeholder="Select a rate limit type" />
              </SelectTrigger>
              <SelectContent>
                {rateLimitComponents.map((component) => (
                  <SelectItem key={component} value={component}>
                    {componentSchemas.rate_limit?.[
                      component as keyof typeof componentSchemas.rate_limit
                    ]?.title || component}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
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
              {initialData ? "Update Rate Limit" : "Create Rate Limit"}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
