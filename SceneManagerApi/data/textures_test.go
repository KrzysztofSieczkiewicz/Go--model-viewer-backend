package data

import "testing"

// Example requests for api testing:
// PUT:
// curl -v http://localhost:9090/textures/FUCCNu--2Lru2QoKhR3zc -XPUT -d "{\"name\":\"NewTestName\", \"path\":\"NewPath/asd\", \"tags\":[]}"
//
// POST:
// curl -v http://localhost:9090/textures -XPOST -d "{\"name\":\"NewTestName\", \"path\":\"filepath/to/asset\"}"
//
// GET:
// curl -v http://localhost:9090/textures
//
// GET: (SINGLE)
// curl -v http://localhost:9090/textures/FUCCNu--2Lru2QoKhR3zcas




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