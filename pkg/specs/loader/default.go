package loader

var DefaultLoader SpecsLoader

func init() {
	yamlLoader := NewYamlLoader()
	DefaultLoader = CurrentWorkingDirectoryLoader(
		AscendingFileLoader(
			PriorityFileLoader(
				AppendFileLoader(yamlLoader, "incredible.yml"),
				AppendFileLoader(yamlLoader, ".incredible.yml"),
				AppendFileLoader(yamlLoader, "incredible.json"),
				AppendFileLoader(yamlLoader, ".incredible.json"),
			),
		),
	)
}
