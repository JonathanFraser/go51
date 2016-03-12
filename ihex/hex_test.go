package ihex

import (
	"bytes"
	"testing"
)

func TestBlank(t *testing.T) {
	r := bytes.NewReader([]byte{})
	_, err := Parse(r)
	if err == nil {
		t.Error("blank file not considered an error")
	}
}

func TestJustEOF(t *testing.T) {
	r := bytes.NewBufferString(":00000001FF\n")
	_, err := Parse(r)
	if err != nil {
		t.Error(err)
	}
}

func TestExtraBlank(t *testing.T) {
	r := bytes.NewBufferString(":00000001FF\n\n")
	_, err := Parse(r)
	if err == nil {
		t.Error("failed to catch extra blank line")
	}
}
