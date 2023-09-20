package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
	"github.com/stretchr/testify/assert"
)

func TestHTTP_CreateCapsule(t *testing.T) {
	body := fmt.Sprintf(`{
		"message": "test message",
		"openAt": "%s"
	}`, time.Now().UTC().AddDate(1, 0, 0).Format(time.RFC3339))

	Test(t,
		Description("Create Capsule Success"),
		Post(basePath+"/capsules"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add(fmt.Sprintf("Bearer %s", token)),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusCreated),
		Expect().Body().JSON().JQ(".message").Equal("test message"),
	)
}

func TestHTTP_GetCapsules(t *testing.T) {
	Test(t,
		Description("Get Capsules Success"),
		Get(basePath+"/capsules"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".[0].message").Equal("test message"),
		Store().Response().Body().JSON().JQ(".[0].id").In(&capsuleID),
	)
}

func TestHTTP_GetCapsule(t *testing.T) {
	Test(t,
		Description("Get Capsule Success"),
		Get(basePath+"/capsules/"+capsuleID),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".message").Equal("test message"),
	)
}

// Framework does not support a PATCH method =(
func TestHTTP_UpdateCapsule(t *testing.T) {
	reqURL := basePath + "/capsules/" + capsuleID
	body := `{
		"message": "new test message"
	}`

	req, err := http.NewRequest(http.MethodPatch, reqURL, bytes.NewBufferString(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusNoContent)

	Test(t,
		Description("Get Updated Capsule Success"),
		Get(reqURL),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().JSON().JQ(".message").Equal("new test message"),
	)
}

func TestHTTP_DeleteCapsule(t *testing.T) {
	reqURL := basePath + "/capsules/" + capsuleID

	Test(t,
		Description("Get Capsule Success"),
		Delete(reqURL),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
		Expect().Status().Equal(http.StatusNoContent),
	)

	Test(t,
		Description("Get Deleted Capsule Success"),
		Get(reqURL),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add("Bearer "+token),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().JQ(".message").Equal("not found"),
	)
}
