package contextargs

// GenerateVariables from context args
func GenerateVariables(ctx *Context) map[string]any {
	vars := map[string]any{
		"ip": ctx.MetaInput.CustomIP,
	}
	return vars
}
