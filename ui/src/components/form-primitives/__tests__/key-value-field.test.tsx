import { describe, it, expect } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { z } from "zod";
import { KeyValueField } from "../key-value-field";
import {
  parseYaml,
  serializeYaml,
} from "@/components/flow-components/utils/yaml";

const schema = z.object({ headers: z.record(z.string(), z.string()) });
const defaults = { headers: {} };

// Mirrors the editor hosts: state is serialized YAML, re-parsed on every change.
function LossyYamlParent({ initialYaml = "" }: { initialYaml?: string }) {
  const [yamlStr, setYamlStr] = useState(initialYaml);
  const parsed = parseYaml(schema, yamlStr, defaults);
  return (
    <KeyValueField
      label="Headers"
      value={parsed.headers}
      onChange={(headers) => setYamlStr(serializeYaml({ headers }))}
    />
  );
}

describe("KeyValueField", () => {
  it("keeps a newly added row visible through a lossy YAML round-trip", () => {
    render(<LossyYamlParent />);

    fireEvent.click(screen.getByRole("button", { name: /add/i }));

    expect(screen.getAllByPlaceholderText("Key")).toHaveLength(1);
  });

  it("lets the user fill in a new entry end to end", () => {
    render(<LossyYamlParent />);

    fireEvent.click(screen.getByRole("button", { name: /add/i }));
    const keyInput = screen.getByPlaceholderText("Key");
    fireEvent.change(keyInput, { target: { value: "Content-Type" } });
    const valueInput = screen.getByPlaceholderText("Value");
    fireEvent.change(valueInput, { target: { value: "application/json" } });

    expect(screen.getByDisplayValue("Content-Type")).toBeInTheDocument();
    expect(screen.getByDisplayValue("application/json")).toBeInTheDocument();
  });

  it("does not collapse rows while a key transiently equals another key", () => {
    render(<LossyYamlParent initialYaml={"headers:\n  Auth: a\n"} />);

    fireEvent.click(screen.getByRole("button", { name: /add/i }));
    const keyInputs = screen.getAllByPlaceholderText("Key");
    fireEvent.change(keyInputs[1], { target: { value: "Auth" } });

    expect(screen.getAllByPlaceholderText("Key")).toHaveLength(2);
  });

  it("removes an entry", () => {
    render(<LossyYamlParent initialYaml={"headers:\n  Auth: a\n"} />);

    const trash = screen
      .getAllByRole("button")
      .find((b) => !b.textContent?.includes("Add"));
    fireEvent.click(trash!);

    expect(screen.queryByDisplayValue("Auth")).not.toBeInTheDocument();
  });

  it("re-syncs rows when the value prop changes externally", () => {
    const { rerender } = render(
      <KeyValueField label="Headers" value={{ a: "1" }} onChange={() => {}} />,
    );
    expect(screen.getByDisplayValue("a")).toBeInTheDocument();

    rerender(
      <KeyValueField label="Headers" value={{ b: "2" }} onChange={() => {}} />,
    );

    expect(screen.getByDisplayValue("b")).toBeInTheDocument();
    expect(screen.queryByDisplayValue("a")).not.toBeInTheDocument();
  });
});
