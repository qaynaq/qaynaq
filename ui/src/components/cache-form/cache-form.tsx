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
import { Checkbox } from "@/components/ui/checkbox";
import { Textarea } from "@/components/ui/textarea";
import { componentSchemas, componentLists } from "@/lib/component-schemas";
import { ChevronDown, ChevronRight } from "lucide-react";

interface CacheFormProps {
  initialData?: {
    label: string;
    component: string;
    config: any;
  };
  onSubmit: (data: { label: string; component: string; config: any }) => void;
  onCancel: () => void;
}

export function CacheForm({ initialData, onSubmit, onCancel }: CacheFormProps) {
  const [label, setLabel] = useState(initialData?.label || "");
  const [selectedComponent, setSelectedComponent] = useState(
    initialData?.component || "",
  );
  const [config, setConfig] = useState<any>(initialData?.config || {});
  const [expandedSections, setExpandedSections] = useState<Set<string>>(
    new Set(),
  );

  const cacheComponents = componentLists.cache || [];
  const selectedSchema =
    selectedComponent && componentSchemas.cache
      ? componentSchemas.cache[
          selectedComponent as keyof typeof componentSchemas.cache
        ]
      : null;

  useEffect(() => {
    if (selectedComponent && !initialData) {
      // Initialize config with default values when component changes
      const schema =
        componentSchemas.cache?.[
          selectedComponent as keyof typeof componentSchemas.cache
        ];
      if (schema?.properties) {
        const defaultConfig: any = {};
        Object.entries(schema.properties).forEach(
          ([key, prop]: [string, any]) => {
            if (prop.default !== undefined) {
              if (prop.type === "object" && typeof prop.default === "string") {
                try {
                  defaultConfig[key] = JSON.parse(prop.default);
                } catch {
                  defaultConfig[key] = prop.default;
                }
              } else if (
                prop.type === "array" &&
                typeof prop.default === "string"
              ) {
                try {
                  defaultConfig[key] = JSON.parse(prop.default);
                } catch {
                  defaultConfig[key] = [];
                }
              } else {
                defaultConfig[key] = prop.default;
              }
            }
          },
        );
        setConfig(defaultConfig);
      }
    }
  }, [selectedComponent, initialData]);

  const toggleSection = (key: string) => {
    const newExpanded = new Set(expandedSections);
    if (newExpanded.has(key)) {
      newExpanded.delete(key);
    } else {
      newExpanded.add(key);
    }
    setExpandedSections(newExpanded);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      label,
      component: selectedComponent,
      config,
    });
  };

  const renderField = (key: string, field: any, path: string[] = []) => {
    const fullPath = [...path, key].join(".");
    const value =
      path.length === 0 ? config[key] : getNestedValue(config, [...path, key]);

    const updateValue = (newValue: any) => {
      const newConfig = { ...config };
      if (path.length === 0) {
        newConfig[key] = newValue;
      } else {
        setNestedValue(newConfig, [...path, key], newValue);
      }
      setConfig(newConfig);
    };

    // Handle nested objects
    if (field.type === "object" && field.properties) {
      const isExpanded = expandedSections.has(fullPath);
      return (
        <div
          key={fullPath}
          className="space-y-2 border-l-2 border-gray-200 pl-4 mb-4"
        >
          <div
            className="flex items-center cursor-pointer"
            onClick={() => toggleSection(fullPath)}
          >
            {isExpanded ? (
              <ChevronDown className="h-4 w-4 mr-2" />
            ) : (
              <ChevronRight className="h-4 w-4 mr-2" />
            )}
            <Label className="font-semibold cursor-pointer">
              {field.title || key}
            </Label>
          </div>
          {field.description && (
            <p className="text-sm text-muted-foreground">{field.description}</p>
          )}
          {isExpanded && (
            <div className="space-y-3 mt-2">
              {Object.entries(field.properties).map(
                ([nestedKey, nestedField]) =>
                  renderField(nestedKey, nestedField, [...path, key]),
              )}
            </div>
          )}
        </div>
      );
    }

    // Handle different field types
    switch (field.type) {
      case "boolean":
        return (
          <div key={fullPath} className="flex items-center space-x-2 mb-4">
            <Checkbox
              id={fullPath}
              checked={value || false}
              onCheckedChange={(checked) => updateValue(checked)}
            />
            <Label htmlFor={fullPath} className="cursor-pointer">
              {field.title || key}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {field.description && (
              <p className="text-sm text-muted-foreground ml-6">
                {field.description}
              </p>
            )}
          </div>
        );

      case "number":
        return (
          <div key={fullPath} className="space-y-2 mb-4">
            <Label htmlFor={fullPath}>
              {field.title || key}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {field.description && (
              <p className="text-sm text-muted-foreground">
                {field.description}
              </p>
            )}
            <Input
              id={fullPath}
              type="number"
              value={value !== undefined ? value : field.default || ""}
              onChange={(e) => updateValue(Number(e.target.value))}
            />
          </div>
        );

      case "array":
        return (
          <div key={fullPath} className="space-y-2 mb-4">
            <Label htmlFor={fullPath}>
              {field.title || key}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {field.description && (
              <p className="text-sm text-muted-foreground">
                {field.description}
              </p>
            )}
            <Textarea
              id={fullPath}
              value={
                value !== undefined
                  ? typeof value === "string"
                    ? value
                    : JSON.stringify(value, null, 2)
                  : field.default || "[]"
              }
              onChange={(e) => {
                try {
                  const parsed = JSON.parse(e.target.value);
                  updateValue(parsed);
                } catch {
                  updateValue(e.target.value);
                }
              }}
              placeholder="[]"
              rows={3}
            />
          </div>
        );

      case "string":
      default:
        if (field.options) {
          return (
            <div key={fullPath} className="space-y-2 mb-4">
              <Label htmlFor={fullPath}>
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
          <div key={fullPath} className="space-y-2 mb-4">
            <Label htmlFor={fullPath}>
              {field.title || key}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </Label>
            {field.description && (
              <p className="text-sm text-muted-foreground">
                {field.description}
              </p>
            )}
            {field.type === "object" && !field.properties ? (
              <Textarea
                id={fullPath}
                value={
                  value !== undefined
                    ? typeof value === "string"
                      ? value
                      : JSON.stringify(value, null, 2)
                    : field.default || "{}"
                }
                onChange={(e) => {
                  try {
                    const parsed = JSON.parse(e.target.value);
                    updateValue(parsed);
                  } catch {
                    updateValue(e.target.value);
                  }
                }}
                placeholder="{}"
                rows={3}
              />
            ) : (
              <Input
                id={fullPath}
                type="text"
                value={value !== undefined ? value : field.default || ""}
                onChange={(e) => updateValue(e.target.value)}
              />
            )}
          </div>
        );
    }
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>{initialData ? "Edit Cache" : "Add New Cache"}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="cache-label">
              Label <span className="text-red-500">*</span>
            </Label>
            <Input
              id="cache-label"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              placeholder="my_cache"
              required
            />
            <p className="text-sm text-muted-foreground">
              A unique identifier for this cache resource
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="cache-component">
              Cache Type <span className="text-red-500">*</span>
            </Label>
            <Select
              value={selectedComponent}
              onValueChange={setSelectedComponent}
              required
            >
              <SelectTrigger>
                <SelectValue placeholder="Select a cache type" />
              </SelectTrigger>
              <SelectContent>
                {cacheComponents.map((component) => (
                  <SelectItem key={component} value={component}>
                    {componentSchemas.cache?.[
                      component as keyof typeof componentSchemas.cache
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
              {initialData ? "Update Cache" : "Create Cache"}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

// Helper functions for nested object operations
function getNestedValue(obj: any, path: string[]): any {
  return path.reduce((current, key) => current?.[key], obj);
}

function setNestedValue(obj: any, path: string[], value: any): void {
  const lastKey = path[path.length - 1];
  const parent = path.slice(0, -1).reduce((current, key) => {
    if (!current[key]) {
      current[key] = {};
    }
    return current[key];
  }, obj);
  parent[lastKey] = value;
}
