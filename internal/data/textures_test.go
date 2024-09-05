package data

import "testing"

func TestValidationNegative(t *testing.T) {
	texture := &Texture{}

	err := texture.Validate()

	if err == nil {
		t.Fatal(err)
	}
}

func TestValidationPositive(t *testing.T) {
	texture := &Texture{
		ID:       "UT_UID_ASDFADF",
		Name: 	  "UT_Name_Falafel",
		FilePath: "files/assets/unitTests.obj",
	}

	err := texture.Validate()

	if err != nil {
		t.Fatal(err)
	}
}