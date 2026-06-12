# Templates

This directory contains Qaynaq's built-in templates. Each `*.yaml` file is one template: a bundle of flows that covers a use case end to end. The files are embedded into the Qaynaq binary at build time and served through the template catalog API, which powers the Templates wizard in the UI.

## Manifest format

A template is a single YAML file with this structure:

```yaml
id: shopify                      # unique, lowercase, [a-z0-9_], stable forever
name: Shopify
description: MCP tools for accessing your Shopify store data
version: 0.1.0
variables:
  - key: shop_name               # referenced in flows as {{ shop_name }}
    title: Store Name
    description: Your store name (without .myshopify.com)
    type: string                 # string | secret | connection
    required: true
    placeholder: mystore
  - key: api_access_token
    title: Admin API Access Token
    type: secret
    required: true
flows:
  - name: shopify_list_orders    # unique within the template, becomes the flow name
    kind: tool                   # tool (exposed via /mcp) | automation
    description: List recent orders from your Shopify store
    input:
      component: mcp_tool
      config:
        name: shopify_list_orders
        description: List recent orders from your Shopify store
        input_schema:
          - { name: limit, type: string, required: false, description: Max orders to return }
    processors:
      - label: shopify
        component: shopify
        config:
          action: list_orders
          shop_name: "{{ shop_name }}"
          api_access_token: "{{ api_access_token }}"
          limit: ${!this.limit.or("50")}
    output:
      component: sync_response
```

## Rules

- **Flows use Qaynaq's component model**, the same `component` + `config` structure the flow builder produces. Never raw engine configuration. Each `component` must be a registered Qaynaq component, and configs are validated against the component schemas at load and at install time.
- **`{{ key }}` placeholders** are substituted at install time with the values the user enters in the wizard. Every placeholder must match a declared variable; unknown placeholders fail validation. Substitution only touches `{{ }}` markers, so Bento-style runtime interpolation like `${!this.limit.or("50")}` passes through untouched.
- **Variable types decide what the user sees and what gets substituted:**
  - `string` - plain text input, substituted as-is.
  - `secret` - secret picker in the wizard. The substituted value is a `${SECRET_KEY}` reference, never the raw value. Installed flows are safe to export.
  - `connection` - OAuth connection picker. The substituted value is a `${QAYNAQ_CONN_<name>}` reference resolved at runtime.
- **`id` and flow `name`s are stable identifiers.** Installed flows are tagged with the template id (`managed_by`), and redeploys match by flow name. Renaming either orphans existing deployments.
- **One file per template.** The `file:` key on flow entries is reserved for future multi-file templates and is rejected by the loader.
- **`version`** is required. Bump it when you change a template.

## Validation

The loader validates every manifest at startup and in tests:

```
go test ./internal/templates/ ./internal/api/coordinator/
```

`TestEmbeddedTemplatesPassBentoValidation` renders every flow of every template with representative values and runs it through the same per-component linting that gates flow activation. If your template passes that test, it will install cleanly.
