// Root test: mapErrorStatus and domainError are unexported; tests/ can only reach the public API.
package editorial

import "testing"

func TestMapErrorStatus(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{nil, 200},
		{ErrNotFound, 404},
		{ErrAlreadyExists, 409},
		{ErrInvalidTransition, 400},
		{ErrReasonRequired, 400},
		{domainError("some other error"), 500},
	}

	for _, tc := range tests {
		if got := mapErrorStatus(tc.err); got != tc.code {
			t.Errorf("mapErrorStatus(%v) = %d; want %d", tc.err, got, tc.code)
		}
	}
}
