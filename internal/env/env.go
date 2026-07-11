package env

func SetEnvVar(name, value string) error {
	return setEnvVarOS(name, value)
}

func GetEnvVar(name string) (string, error) {
	return getEnvVarOS(name)
}

func RemoveEnvVar(name string) error {
	return removeEnvVarOS(name)
}

func AddToPath(pathEntry string) error {
	return addToPathOS(pathEntry)
}

func RemoveFromPath(pathEntry string) error {
	return removeFromPathOS(pathEntry)
}

func SetEnvVars(jdkHome string) error {
	return setEnvVarsOS(jdkHome)
}
