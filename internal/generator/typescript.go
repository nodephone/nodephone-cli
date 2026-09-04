package generator

import (
	"fmt"
	"sort"

	"strings"
)

type GeneratedFiles map[string]string

// GenerateTypeScript parses OpenAPISpec and produces a map of TS files
func GenerateTypeScript(spec *OpenAPISpec) (GeneratedFiles, error) {
	files := make(GeneratedFiles)

	// Domain type buckets
	authTypes := make([]string, 0)
	dbTypes := make([]string, 0)
	storageTypes := make([]string, 0)
	functionsTypes := make([]string, 0)
	apiTypes := make([]string, 0)

	// Sort schema names for deterministic generation
	schemaNames := make([]string, 0, len(spec.Components.Schemas))
	for name := range spec.Components.Schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	for _, name := range schemaNames {
		schema := spec.Components.Schemas[name]
		tsCode := renderInterface(name, schema)

		lower := strings.ToLower(name)
		switch {
		case strings.Contains(lower, "auth") || strings.Contains(lower, "login") || strings.Contains(lower, "user") || strings.Contains(lower, "token"):
			authTypes = append(authTypes, tsCode)
		case strings.Contains(lower, "db") || strings.Contains(lower, "schema") || strings.Contains(lower, "table") || strings.Contains(lower, "migration"):
			dbTypes = append(dbTypes, tsCode)
		case strings.Contains(lower, "storage") || strings.Contains(lower, "bucket") || strings.Contains(lower, "file"):
			storageTypes = append(storageTypes, tsCode)
		case strings.Contains(lower, "function") || strings.Contains(lower, "event") || strings.Contains(lower, "handler"):
			functionsTypes = append(functionsTypes, tsCode)
		default:
			apiTypes = append(apiTypes, tsCode)
		}
	}

	// 1. auth.ts
	files["auth.ts"] = wrapFileHeader("Authentication Types", authTypes)

	// 2. database.ts
	files["database.ts"] = wrapFileHeader("Database Types", dbTypes)

	// 3. storage.ts
	files["storage.ts"] = wrapFileHeader("Storage Types", storageTypes)

	// 4. functions.ts
	files["functions.ts"] = wrapFileHeader("Serverless Function Types", functionsTypes)

	// 5. api.ts
	apiRouteTypes := renderRoutes(spec)
	allApiContent := append(apiTypes, apiRouteTypes)
	files["api.ts"] = wrapFileHeader("API Models & Route Endpoints", allApiContent)

	// 6. index.ts
	files["index.ts"] = `/**
 * NodePhone Auto-Generated TypeScript Definitions
 * DO NOT EDIT MANUALLY - Synchronized via 'nodephone gen types'
 */

export * from './auth';
export * from './database';
export * from './storage';
export * from './functions';
export * from './api';
`

	return files, nil
}

func wrapFileHeader(title string, typeSnippets []string) string {
	var sb strings.Builder
	sb.WriteString("/**\n")
	sb.WriteString(fmt.Sprintf(" * NodePhone Generated %s\n", title))
	sb.WriteString(" * DO NOT EDIT MANUALLY\n")
	sb.WriteString(" */\n\n")

	if len(typeSnippets) == 0 {
		sb.WriteString("// No domain schemas found\nexport {};\n")
	} else {
		sb.WriteString(strings.Join(typeSnippets, "\n\n"))
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderInterface(name string, schema Schema) string {
	var sb strings.Builder

	if schema.Description != "" {
		sb.WriteString(fmt.Sprintf("/** %s */\n", schema.Description))
	}

	// Check if enum
	if len(schema.Enum) > 0 {
		sb.WriteString(fmt.Sprintf("export type %s = ", name))
		enumVals := make([]string, 0, len(schema.Enum))
		for _, v := range schema.Enum {
			enumVals = append(enumVals, fmt.Sprintf("%q", v))
		}
		sb.WriteString(strings.Join(enumVals, " | "))
		sb.WriteString(";")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("export interface %s {\n", name))

	requiredSet := make(map[string]bool)
	for _, req := range schema.Required {
		requiredSet[req] = true
	}

	// Sort property keys for deterministic output
	propNames := make([]string, 0, len(schema.Properties))
	for prop := range schema.Properties {
		propNames = append(propNames, prop)
	}
	sort.Strings(propNames)

	for _, prop := range propNames {
		propSchema := schema.Properties[prop]
		optional := "?"
		if requiredSet[prop] {
			optional = ""
		}

		tsType := openAPITypeToTS(propSchema)
		if propSchema.Description != "" {
			sb.WriteString(fmt.Sprintf("  /** %s */\n", propSchema.Description))
		}
		sb.WriteString(fmt.Sprintf("  %s%s: %s;\n", prop, optional, tsType))
	}

	sb.WriteString("}")
	return sb.String()
}

func openAPITypeToTS(s Schema) string {
	if s.Ref != "" {
		return ExtractRefName(s.Ref)
	}

	switch s.Type {
	case "string":
		return "string"
	case "integer", "number":
		return "number"
	case "boolean":
		return "boolean"
	case "array":
		if s.Items != nil {
			return openAPITypeToTS(*s.Items) + "[]"
		}
		return "any[]"
	case "object":
		if len(s.Properties) > 0 {
			props := make([]string, 0, len(s.Properties))
			for k, v := range s.Properties {
				props = append(props, fmt.Sprintf("%s: %s", k, openAPITypeToTS(v)))
			}
			return "{ " + strings.Join(props, "; ") + " }"
		}
		return "Record<string, any>"
	default:
		return "any"
	}
}

func renderRoutes(spec *OpenAPISpec) string {
	var sb strings.Builder
	sb.WriteString("export interface APIRoutes {\n")

	paths := make([]string, 0, len(spec.Paths))
	for p := range spec.Paths {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, path := range paths {
		item := spec.Paths[path]
		sb.WriteString(fmt.Sprintf("  %q: {\n", path))

		if item.Get != nil {
			sb.WriteString(fmt.Sprintf("    GET: { summary: %q };\n", item.Get.Summary))
		}
		if item.Post != nil {
			sb.WriteString(fmt.Sprintf("    POST: { summary: %q };\n", item.Post.Summary))
		}
		if item.Put != nil {
			sb.WriteString(fmt.Sprintf("    PUT: { summary: %q };\n", item.Put.Summary))
		}
		if item.Delete != nil {
			sb.WriteString(fmt.Sprintf("    DELETE: { summary: %q };\n", item.Delete.Summary))
		}
		sb.WriteString("  };\n")
	}

	sb.WriteString("}")
	return sb.String()
}
