package task

import "testing"

func TestStatusIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{
			name:   "incomplete",
			status: StatusIncomplete,
			want:   true,
		},
		{
			name:   "completed",
			status: StatusCompleted,
			want:   true,
		},
		{
			name:   "negative status",
			status: Status(-1),
			want:   false,
		},
		{
			name:   "status above valid range",
			status: Status(2),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()

			if got != tt.want {
				t.Errorf(
					"Status(%d).IsValid() = %t, want %t",
					tt.status,
					got,
					tt.want,
				)
			}
		})
	}
}
