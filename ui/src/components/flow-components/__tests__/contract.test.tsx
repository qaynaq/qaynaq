import { describe, it, expect } from "vitest";
import { render } from "@testing-library/react";
import { Suspense } from "react";
import { listAll } from "../registry";

const components = listAll();

describe("registry", () => {
  it("loads without throwing", () => {
    expect(Array.isArray(components)).toBe(true);
  });
});

describe.each(components.map((c) => [c.category, c.id, c] as const))(
  "%s:%s",
  (_category, _id, component) => {
    it("default config conforms to its schema", () => {
      const result = component.configSchema.safeParse(component.defaultConfig);
      if (!result.success) {
        throw new Error(
          `default config failed validation: ${JSON.stringify(result.error.issues)}`,
        );
      }
    });

    it("serialize then parse returns the default config", () => {
      const serialized = component.serialize(component.defaultConfig);
      const parsed = component.parse(serialized);
      expect(parsed).toEqual(component.defaultConfig);
    });

    it("parse of empty string returns the default config", () => {
      expect(component.parse("")).toEqual(component.defaultConfig);
    });

    it("editor renders without throwing", () => {
      const Editor = component.Editor;
      const noop = () => {};
      expect(() =>
        render(
          <Suspense fallback={null}>
            <Editor value={component.defaultConfig} onChange={noop} />
          </Suspense>,
        ),
      ).not.toThrow();
    });
  },
);
