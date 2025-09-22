# Templating

**frank** supports Jinja templating and HCL for dynamic Kubernetes manifest generation. This allows you to create flexible, reusable templates that adapt to different environments and configurations.

## Supported File Extensions

**frank** automatically detects and processes Jinja templates with these extensions:

- `.jinja` - Standard Jinja template files
- `.j2` - Alternative Jinja extension
- `.hcl` - Standard HCL files
- `.tf` - Alternative HCL extension
