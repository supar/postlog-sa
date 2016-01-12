package main

import (
	"reflect"
	"testing"
)

func TestStrEmpty_Empty(t *testing.T) {
	if v := StrEmpty("   ", "TestStrEmpty_Empty"); v != "TestStrEmpty_Empty" {
		t.Fatalf("Expected value: TestStrEmpty_Empty, but got '%s'", v)
	}
}

func TestStrEmpty_SomeValue(t *testing.T) {
	if v := StrEmpty("TestStrEmpty_SomeValue", "TestStrEmpty_Empty"); v != "TestStrEmpty_SomeValue" {
		t.Fatalf("Expected value: TestStrEmpty_SomeValue, but got '%s'", v)
	}
}

func TestStrEmpty_SomeValue_Trim(t *testing.T) {
	if v := StrEmpty(" TestStrEmpty_SomeValue ", "TestStrEmpty_Empty"); v != "TestStrEmpty_SomeValue" {
		t.Fatalf("Expected value: TestStrEmpty_SomeValue, but got '%s'", v)
	}
}

func TestInt2Str_IntValue(t *testing.T) {
	if v := Int2Str(23434); reflect.TypeOf(v).Kind() != reflect.String && v != "23434" {
		t.Fatalf("Expected int convertion to the string, value 23434")
	}
}

func TestInt2Str_StringAsIs(t *testing.T) {
	if v := Int2Str("23434"); reflect.TypeOf(v).Kind() != reflect.String && v != "23434" {
		t.Fatalf("Expected int convertion to the string, value 23434")
	}
}

func TestStr2Int_StringValue(t *testing.T) {
	if v := Str2Int("85746465"); reflect.TypeOf(v).Kind() != reflect.Int && v != 85746465 {
		t.Fatalf("Expected string convertion to the int, value 85746465")
	}
}

func TestStr2Int_IntAsIs(t *testing.T) {
	if v := Str2Int(85746465); reflect.TypeOf(v).Kind() != reflect.Int && v != 85746465 {
		t.Fatalf("Expected string convertion to the int, value 85746465")
	}
}

func TestStr2Bool_StringValue(t *testing.T) {
	if v := Str2Bool("1"); reflect.TypeOf(v).Kind() != reflect.Bool || v != true {
		t.Fatalf("Expected true value, but got %v", v)
	}
}

func TestStr2Bool_IntValue(t *testing.T) {
	if v := Str2Bool(-1); reflect.TypeOf(v).Kind() != reflect.Bool || v != false {
		t.Fatalf("Expected false value, but got %v", v)
	}
}

func TestStr2Bool_BoolTrueValue(t *testing.T) {
	if v := Str2Bool(true); reflect.TypeOf(v).Kind() != reflect.Bool || v != true {
		t.Fatalf("Expected true value, but got %v", v)
	}
}

func TestStr2Bool_BoolFalseValue(t *testing.T) {
	if v := Str2Bool(false); reflect.TypeOf(v).Kind() != reflect.Bool || v != false {
		t.Fatalf("Expected false value, but got %v", v)
	}
}
