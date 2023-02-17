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
		currentSubnets: &map[string]string{"test1": "10.116.4.0/22", "test3": "10.116.0.0/24", "test4": "10.116.2.0/25", "test5": "10.116.3.0/24", "test6": "10.119.0.0/16", "foo": "10.116.8.0/24"},
		prefixLength:   25,
		baseCidrRange:  "10.116.0.0/14",
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
	if netmask != "10.116.1.128/25" {
		t.Fatalf("Unexpected value for next netmask %s", netmask)
	}
}

func TestCorrectNextCidr16StillFits(t *testing.T) {
	testData := initTestData()
	testData.prefixLength = 16
	cidrCalculator, err := New(testData.currentSubnets, testData.prefixLength, testData.baseCidrRange)
	if err != nil {
		t.Fatal(err)
	}
	netmask, err := cidrCalculator.GetNextNetmask()
	if err != nil {
		t.Fatal(err)
	}
	if netmask != "10.118.0.0/16" {
		t.Fatalf("Unexpected value for next netmask %s", netmask)
	}
}

func TestBaseCidrExhausted(t *testing.T) {
	testData := initTestData()
	expected := fmt.Sprintf("baseCidrRange %s is exhausted!", testData.baseCidrRange)
	testData.prefixLength = 15
	theCidrCalculator, err := New(testData.currentSubnets, testData.prefixLength, testData.baseCidrRange)
	if err != nil {
		t.Fatal(err)
	}
	next := ""
	next, err = theCidrCalculator.GetNextNetmask()
	println(next)
	if err == nil {
		t.Fatal("There should be an error when Cidr Range is exhausted!!!")
	}
	if err.Error() != expected {
		t.Fatalf("The error message does not match: Expected %s, Got %s", expected, err.Error())
	}
}
