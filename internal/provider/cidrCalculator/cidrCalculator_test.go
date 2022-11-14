package cidrCalculator

import (
	"fmt"
	"testing"
)

type TestData struct {
	currentSubnets *map[string]string
	prefixLength   int8
	baseCidrRange  string
}

func initTestData() *TestData {
	return &TestData{
		currentSubnets: &map[string]string{"test1": "10.5.1.0/24", "test2": "10.5.0.0/26"},
		prefixLength:   26,
		baseCidrRange:  "10.5.0.0/16",
	}
}

func TestCorrectNextCidr(t *testing.T) {
	testData := initTestData()
	cidrCalculator, err := New(testData.currentSubnets, testData.prefixLength, testData.baseCidrRange)
	if err != nil {
		t.Fatal(err)
	}
	netmask, err := cidrCalculator.GetNextNetmask()
	if err != nil {
		t.Fatal(err)
	}
	if netmask != "10.5.0.64/26" {
		t.Fatalf("Unexpected value for next netmask %s", netmask)
	}
}

func TestBaseCidrExhausted(t *testing.T) {
	testData := initTestData()
	expected := fmt.Sprintf("baseCidrRange %s is exhausted!", testData.baseCidrRange)
	(*testData.currentSubnets)["testExhausted"] = "10.5.128.0/17"
	testData.prefixLength = 17
	theCidrCalculator, err := New(testData.currentSubnets, testData.prefixLength, testData.baseCidrRange)
	if err != nil {
		t.Fatal(err)
	}
	_, err = theCidrCalculator.GetNextNetmask()
	if err == nil {
		t.Fatal("There should be an error when Cidr Range is exhausted!!!")
	}
	if err.Error() != expected {
		t.Fatalf("The error message does not match: Expected %s, Got %s", expected, err.Error())
	}
}
