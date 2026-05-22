package rules

func AllRules() []Rule {
	return []Rule{
		NewUndefinedReceiverRule(),
		NewUndefinedProcessorRule(),
		NewUndefinedExporterRule(),
		NewUndefinedExtensionRule(),
		NewUnusedReceiverRule(),
		NewUnusedProcessorRule(),
		NewUnusedExporterRule(),
		NewEmptyPipelineRule(),
	}
}
