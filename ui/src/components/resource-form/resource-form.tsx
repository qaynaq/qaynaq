import { Suspense, useEffect, useMemo, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  TextField,
  SelectField,
  type SelectOption,
} from "@/components/form-primitives";
import {
  getComponent,
  listComponents,
} from "@/components/flow-components/registry";
import { flattenZodErrors } from "@/components/flow-components/utils/errors";
import type { ComponentCategory } from "@/components/flow-components/types";

export interface ResourceFormSubmit {
  label: string;
  component: string;
  config: unknown;
}

interface ResourceFormProps {
  category: ComponentCategory;
  resourceLabel: string;
  initialData?: { label: string; component: string; configYaml: string };
  onSubmit: (data: ResourceFormSubmit) => void;
  onCancel: () => void;
}

export function ResourceForm({
  category,
  resourceLabel,
  initialData,
  onSubmit,
  onCancel,
}: ResourceFormProps) {
  const choices = useMemo(() => listComponents(category), [category]);
  const componentOptions: SelectOption[] = choices.map((c) => ({
    value: c.id,
    label: c.name,
  }));

  const [label, setLabel] = useState(initialData?.label ?? "");
  const [selectedId, setSelectedId] = useState(initialData?.component ?? "");
  const [configYaml, setConfigYaml] = useState(initialData?.configYaml ?? "");

  const component = selectedId
    ? getComponent(category, selectedId)
    : undefined;

  useEffect(() => {
    if (!selectedId) return;
    if (initialData && initialData.component === selectedId) return;
    const c = getComponent(category, selectedId);
    if (!c) return;
    setConfigYaml(c.serialize(c.defaultConfig));
  }, [category, selectedId, initialData]);

  const parsed = useMemo(() => {
    if (!component) return null;
    try {
      return component.parse(configYaml);
    } catch {
      return component.defaultConfig;
    }
  }, [component, configYaml]);

  const errors = useMemo(() => {
    if (!component || parsed === null) return undefined;
    const result = component.configSchema.safeParse(parsed);
    return result.success ? undefined : flattenZodErrors(result.error);
  }, [component, parsed]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!component || parsed === null) return;
    onSubmit({ label, component: selectedId, config: parsed });
  };

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>
          {initialData ? `Edit ${resourceLabel}` : `Add New ${resourceLabel}`}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          <TextField
            label="Label"
            description={`A unique identifier for this ${resourceLabel.toLowerCase()} resource`}
            required
            value={label}
            onChange={setLabel}
            placeholder={`my_${category}`}
          />

          <SelectField
            label={`${resourceLabel} Type`}
            required
            value={selectedId}
            onChange={setSelectedId}
            options={componentOptions}
            placeholder={`Select a ${resourceLabel.toLowerCase()} type`}
          />

          {component && parsed !== null && (
            <div className="border rounded-lg p-4 space-y-4">
              <h3 className="font-semibold text-lg">Configuration</h3>
              <Suspense
                fallback={
                  <p className="text-sm text-muted-foreground">
                    Loading editor...
                  </p>
                }
              >
                <component.Editor
                  value={parsed as never}
                  onChange={(next: unknown) =>
                    setConfigYaml(component.serialize(next as never))
                  }
                  errors={errors}
                />
              </Suspense>
            </div>
          )}

          <div className="flex justify-end space-x-2 pt-4">
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled={!label || !selectedId}>
              {initialData ? `Update ${resourceLabel}` : `Create ${resourceLabel}`}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
