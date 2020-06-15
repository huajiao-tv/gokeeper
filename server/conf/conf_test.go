package conf

import (
	"testing"
)

func TestSubscription(t *testing.T) {
	s := "abc.conf"
	expectType := SubscribeFile
	actualType := Subscription(s).Type()
	if actualType != expectType {
		t.Error(expectType, actualType, s)
	}

	expectFile := "abc.conf"
	actualFile := Subscription(s).File()
	if actualFile != expectFile {
		t.Error(expectFile, actualFile, s)
	}

	expectSection := ""
	actualSection := Subscription(s).Section()
	if actualSection != expectSection {
		t.Error(expectSection, actualSection, s)
	}

	//
	s = "abc.conf/section"
	expectType = SubscribeSection
	actualType = Subscription(s).Type()
	if actualType != expectType {
		t.Error(expectType, actualType, s)
	}

	expectFile = "abc.conf"
	actualFile = Subscription(s).File()
	if actualFile != expectFile {
		t.Error(expectFile, actualFile, s)
	}

	expectSection = "section"
	actualSection = Subscription(s).Section()
	if actualSection != expectSection {
		t.Error(expectSection, actualSection, s)
	}

	//
	s = "abc/dir"
	expectType = SubscribeDir
	actualType = Subscription(s).Type()
	if actualType != expectType {
		t.Error(expectType, actualType, s)
	}

	expectFile = "abc/dir"
	actualFile = Subscription(s).File()
	if actualFile != expectFile {
		t.Error(expectFile, actualFile, s)
	}

	expectSection = ""
	actualSection = Subscription(s).Section()
	if actualSection != expectSection {
		t.Error(expectSection, actualSection, s)
	}
}

func TestHasConfSuffix(t *testing.T) {
	s := "abc.conf"
	expect := true
	actual := HasConfSuffix(s)
	if actual != expect {
		t.Error(expect, actual)
	}

	s = "abcconf"
	expect = false
	actual = HasConfSuffix(s)
	if actual != expect {
		t.Error(expect, actual)
	}
}
