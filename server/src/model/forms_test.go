package model

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/mailing_list_signup"
	"github.com/stretchr/testify/assert"
)

type interestFormTest struct {
	server         *httptest.Server
	receivedSignup *mailing_list_signup.Signup
}

func newInterestFormTest(t *testing.T) *interestFormTest {
	test := &interestFormTest{}
	test.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var signup mailing_list_signup.Signup
		err := json.NewDecoder(r.Body).Decode(&signup)
		assert.NoError(t, err)
		test.receivedSignup = &signup
		w.WriteHeader(http.StatusOK)
	}))
	return test
}

func (test *interestFormTest) Close() {
	test.server.Close()
}

func TestSubmitInterestForm_SendsCorrectSignupServiceRequest(t *testing.T) {
	test := newInterestFormTest(t)
	config.SignupURI = test.server.URL + "/foo"
	config.SignupAPIKey = "foo"

	db := newTestDB()
	defer db.Close()

	form := InterestFormData{
		ChapterId: 42,
		Form:      "interest",
		Name:      "Test User",
		Email:     "test@example.com",
		Zip:       "12345",
		Phone:     "510-555-5555",
	}

	err := SubmitInterestForm(db, form)
	assert.NoError(t, err)
	receivedSignup := test.receivedSignup

	if receivedSignup == nil {
		t.Fatal("No signup request found")
	}

	assert.Equal(t, 42, receivedSignup.SourceChapterId, "ChapterId should propagate to Enqueue")
	assert.Equal(t, "adb-interest-form", receivedSignup.Source)
	assert.Equal(t, "Test User", receivedSignup.Name)
	assert.Equal(t, "test@example.com", receivedSignup.Email)
	assert.Equal(t, "510-555-5555", receivedSignup.Phone)
}
