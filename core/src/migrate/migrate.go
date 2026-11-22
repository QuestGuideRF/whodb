package migrate
var DeprecatedConfigKey bool
var DeprecatedOpenAICompatibleEnv bool
func Warnings() []string {
	var warnings []string
	if DeprecatedConfigKey {
		warnings = append(warnings, `Deprecated: database connection profile uses "config" key — please rename to "advanced". The "config" key will be removed in a future release.`)
	}
	if DeprecatedOpenAICompatibleEnv {
		warnings = append(warnings, `Deprecated: WHODB_OPENAI_COMPATIBLE_ENDPOINT, WHODB_OPENAI_COMPATIBLE_API_KEY, and WHODB_CUSTOM_MODELS are deprecated and no longer have any effect. Use WHODB_AI_GENERIC_<ID>_* environment variables instead.`)
	}
	return warnings
}