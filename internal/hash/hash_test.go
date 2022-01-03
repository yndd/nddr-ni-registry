package hash

import (
	"fmt"
	"testing"
)

func TestInsertDelete(t *testing.T) {
	keys1 := []struct {
		regName string
		niName  string
	}{
		{regName: "1", niName: "prov"},
		{regName: "2", niName: "infra"},
		{regName: "3", niName: "multus2"},
		{regName: "4", niName: "multus1"},
	}

	keys2 := []struct {
		regName string
		niName  string
	}{
		{regName: "1", niName: "prov"},
		{regName: "2", niName: "infra"},
		{regName: "3", niName: "multus2"},
		{regName: "4", niName: "multus1"},
		{regName: "5", niName: "multus"},
	}

	h := New(10000)

	for _, key := range keys1 {
		idx := h.Insert(key.niName, key.regName, map[string]string{"vpc": "test"})
		fmt.Printf("Key: %s, Idx: %d\n", key, idx)
	}

	a, b := h.GetAllocated()
	fmt.Println(a)
	for _, bb := range b {
		fmt.Println(*bb)
	}

	for _, key := range keys2 {
		h.Delete(key.niName, key.regName, map[string]string{"vpc": "test"})
	}

	a, b = h.GetAllocated()
	fmt.Println(a)
	for _, bb := range b {
		fmt.Println(*bb)
	}

}
