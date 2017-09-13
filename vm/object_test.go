package vm

import "testing"

func TestObjectInt(t *testing.T) {
	obj := ObjectInt{200}

	if obj.String() != "200" {
		t.Errorf("Expected '200', got '%s'", obj.String())
	}

	obj.isObject()
}
