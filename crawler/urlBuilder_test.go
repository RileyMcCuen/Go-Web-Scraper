package crawler

import (
	"testing"
)

func TestRelNoDotURL(t *testing.T) {
	ret, err := nextURL("https://projectriley.co/index.php", "/upload.php")
	if err != nil {
		t.Errorf("Got this error and should not have: %s\n", err.Error())
	}
	if ret.String() != "https://projectriley.co/upload.php" {
		t.Errorf("Expected: %s, but got: %s", "https://projectriley.co/upload.php", ret.String())
	}
}

func TestRelNoDotLongerURL(t *testing.T) {
	ret, err := nextURL("https://projectriley.co/otherDir/index.php", "/upload.php")
	if err != nil {
		t.Errorf("Got this error and should not have: %s\n", err.Error())
	}
	if ret.String() != "https://projectriley.co/upload.php" {
		t.Errorf("Expected: %s, but got: %s", "https://projectriley.co/upload.php", ret.String())
	}
}

func TestRelDotURL(t *testing.T) {
	ret, err := nextURL("https://projectriley.co/index.php", "./upload.php")
	if err != nil {
		t.Errorf("Got this error and should not have: %s\n", err.Error())
	}
	if ret.String() != "https://projectriley.co/upload.php" {
		t.Errorf("Expected: %s, but got: %s", "https://projectriley.co/upload.php", ret.String())
	}
}

func TestRelDotLongerURL(t *testing.T) {
	ret, err := nextURL("https://projectriley.co/otherDir/index.php", "./upload.php")
	if err != nil {
		t.Errorf("Got this error and should not have: %s\n", err.Error())
	}
	if ret.String() != "https://projectriley.co/otherDir/upload.php" {
		t.Errorf("Expected: %s, but got: %s", "https://projectriley.co/otherDir/upload.php", ret.String())
	}
}
