package types

import (
	"fmt"
	"strconv"
	"testing"
)

func TestTruncateFloat64(t *testing.T) {
	in := "123.456789"
	for i := 5; i >= 1; i-- {
		t.Run(fmt.Sprintf("truncate float down to precision: %d", i), func(t *testing.T) {
			val := in[:4+i]
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				t.Fatalf("failed to parse float: %s", err)
			}

			want := TruncateFloat64(num, i)
			if want != num {
				t.Errorf("expected %f, got %f", num, want)
			}
		})
	}
}
