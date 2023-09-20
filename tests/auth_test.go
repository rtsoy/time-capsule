package tests

import (
	"net/http"
	"testing"

	. "github.com/Eun/go-hit"
)

func TestHTTP_SignUp(t *testing.T) {
	body := `{
		"username": "username123",
		"email": "foo@example.com",
		"password": "Qwerty123"
	}`

	Test(t,
		Description("SignUp Success"),
		Post(basePath+"/sign-up"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusCreated),
		Expect().Body().JSON().JQ(".username").Equal("username123"),
		Expect().Body().JSON().JQ(".email").Equal("foo@example.com"),
		Expect().Body().JSON().NotContains("password"),
	)
}

func TestHTTP_SignIn(t *testing.T) {
	body := `{
		"email": "foo@example.com",
		"password": "Qwerty123"
	}`

	Test(t,
		Description("SignIn Success"),
		Post(basePath+"/sign-in"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().Contains("token"),
		Store().Response().Body().JSON().JQ(".token").In(&token),
	)

	headers = CombineSteps(
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
	)
}
