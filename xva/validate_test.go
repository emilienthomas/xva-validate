package xva

import (
	"testing"
)

func TestValidateOK(t *testing.T) {

	isValid, validationStatus, err := Validate("../test/small.xva", 2)
	if err != nil {
		t.Fatalf("small.xva should not return error, returned %v", err)
	}
	if !isValid {
		t.Fatal("small.xva should return valid, returned not valid")
	}
	if validationStatus != "" {
		t.Fatalf("small.xva should return empty validation status, return '%s'", validationStatus)
	}
}

func TestValidateKO(t *testing.T) {

	isValid, validationStatus, err := Validate("../test/smallko.xva", 2)
	if err != nil {
		t.Fatalf("smallko.xva should not return error, returned %v", err)
	}
	if isValid {
		t.Fatal("smallko.xva should return not valid, returned valid")
	}
	if validationStatus == "" {
		t.Fatal("smallko.xva should not return empty validation status")
	}
}
