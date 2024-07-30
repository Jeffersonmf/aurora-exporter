package entities

type environmentTargetValue string

var EnvironmentTarget = struct {
	Local  environmentTargetValue
	AURORA environmentTargetValue
}{Local: "local", AURORA: "aurora"}
