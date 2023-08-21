package run

import (
	"shelob/auth"
	"shelob/cliArgs"
	"shelob/openapi"
	"shelob/request"
	"shelob/response"
	"time"
)

func Run() {
	Start := time.Now()

	Spec, TargetURL, UserName, Password, ApiKey, Token, OutputDir, Detailed, Duration, ExtraArgs := cliArgs.ParseCliArgs()

	AuthCookies := auth.CreateUser(UserName, Password, TargetURL)

	Context, OpenapiData, Router := openapi.ParseOpenapiSpec(Spec)

	for time.Since(Start) < Duration {
		Requests, RequestsValidationInput, RequestsValidationError := request.CreateRequest(*Context, OpenapiData, Router, TargetURL, AuthCookies, UserName, Password, ApiKey, Token, ExtraArgs)
		response.ParseResponse(*Context, Requests, RequestsValidationInput, RequestsValidationError, OutputDir, Detailed)
	}
}
