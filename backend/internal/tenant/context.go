package tenant

import "context"

type contextKey string

const schemaKey contextKey = "tenantSchema"

func ContextWithSchema(ctx context.Context, schema string) context.Context {
	return context.WithValue(ctx, schemaKey, schema)
}

func SchemaFromContext(ctx context.Context) string {
	if schema, ok := ctx.Value(schemaKey).(string); ok {
		return schema
	}
	return ""
}
