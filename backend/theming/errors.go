package theming

type ThemeError interface {
	error
	Code() string
	Path() string
}

type BaseThemeError struct {
	code string
	path string
	msg  string
}

func (e BaseThemeError) Error() string {
	return e.msg
}

func (e BaseThemeError) Code() string {
	return e.code
}

func (e BaseThemeError) Path() string {
	return e.path
}

func ErrUnknownToken(id string) ThemeError {
	return BaseThemeError{
		code: "unknown_token",
		path: "tokens." + id,
		msg:  "unknown token: " + id,
	}
}

func ErrCycle(path string) ThemeError {
	return BaseThemeError{
		code: "cycle_detected",
		path: path,
		msg:  "cycle detected at " + path,
	}
}

func ErrInvalidValue(path, msg string) ThemeError {
	return BaseThemeError{
		code: "invalid_value",
		path: path,
		msg:  msg,
	}
}
