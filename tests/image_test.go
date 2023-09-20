package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
	"github.com/stretchr/testify/assert"
)

var fileBytes []byte

func TestHTTP_AddImage(t *testing.T) {
	body := fmt.Sprintf(`{
		"message": "test message",
		"openAt": "%s"
	}`, time.Now().UTC().AddDate(1, 0, 0).Format(time.RFC3339))

	Test(t,
		Description("Add Image Create Capsule Success"),
		Post(basePath+"/capsules"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add(fmt.Sprintf("Bearer %s", token)),
		Send().Body().String(body),
		Expect().Status().Equal(http.StatusCreated),
		Expect().Body().JSON().JQ(".message").Equal("test message"),
		Store().Response().Body().JSON().JQ(".id").In(&capsuleID),
	)

	reqURL := basePath + "/capsules/" + capsuleID + "/images"

	file, err := os.Open("./fixtures/images/ok.jpg")
	assert.NoError(t, err)
	defer file.Close()

	stat, err := file.Stat()
	assert.NoError(t, err)

	fileBytes = make([]byte, stat.Size())
	_, err = file.Read(fileBytes)
	assert.NoError(t, err)

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("image", "./fixtures/images/ok.jpg")
	assert.NoError(t, err)

	_, err = part.Write(fileBytes)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, reqURL, buf)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusCreated)

	var output map[string]any
	err = json.NewDecoder(resp.Body).Decode(&output)
	assert.NoError(t, err)

	if imgID, ok := output["name"]; !ok {
		t.Fatal("expected name in response body")
	} else {
		imageID = imgID.(string)
	}

	if _, ok := output["size"]; !ok {
		t.Fatal("expected size in response body")
	}
}

func TestHTTP_GetImage(t *testing.T) {
	Test(t,
		Description("Get Image Success"),
		Get(basePath+"/capsules/"+capsuleID+"/images/"+imageID),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add(fmt.Sprintf("Bearer %s", token)),
		Expect().Status().Equal(http.StatusOK),
		Expect().Body().Bytes().Equal(fileBytes),
	)
}

func TestHTTP_RemoveImage(t *testing.T) {
	Test(t,
		Description("Delete Image Success"),
		Delete(basePath+"/capsules/"+capsuleID+"/images/"+imageID),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add(fmt.Sprintf("Bearer %s", token)),
		Expect().Status().Equal(http.StatusNoContent),
	)

	Test(t,
		Description("Get Deleted Image Success"),
		Get(basePath+"/capsules/"+capsuleID+"/images/"+imageID),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Headers("Authorization").Add(fmt.Sprintf("Bearer %s", token)),
		Expect().Status().Equal(http.StatusNotFound),
		Expect().Body().JSON().JQ(".message").Equal("not found"),
	)
}
