package monkey

type patchAndExecOptions struct {
	envVarName, envVarValue string
	removePatched           bool
}

type PatchAndExecOption interface {
	apply(*patchAndExecOptions)
}

func (o *patchAndExecOptions) applyAll(opts ...PatchAndExecOption) {
	for _, option := range opts {
		option.apply(o)
	}
}

type optionFunc func(*patchAndExecOptions)

func (o optionFunc) apply(options *patchAndExecOptions) { o(options) }

// WithEnvVarName sets a custom environment variable name to determine if code runs inside patched executable or not.
func WithEnvVarName(name string) PatchAndExecOption {
	return optionFunc(func(options *patchAndExecOptions) {
		options.envVarName = name
	})
}

// WithEnvVarValue sets exact value for environment variable used to determine if code runs inside patched executable or not.
// Without this option only presence of variable will be checked.
func WithEnvVarValue(value string) PatchAndExecOption {
	return optionFunc(func(options *patchAndExecOptions) {
		options.envVarValue = value
	})
}

// RemovePatchedExecutable enables automatic removal of patched executable.
func RemovePatchedExecutable() PatchAndExecOption {
	return optionFunc(func(options *patchAndExecOptions) {
		options.removePatched = true
	})
}
