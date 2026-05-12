package logging

import "testing"

func TestParseLevel(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		want    Level
		wantErr bool
	}{
		{name: "empty", value: "", want: LevelUnspecified},
		{name: "debug", value: "debug", want: LevelDebug},
		{name: "info", value: "info", want: LevelInfo},
		{name: "warn", value: "warn", want: LevelWarn},
		{name: "warning", value: "warning", want: LevelWarn},
		{name: "error", value: "error", want: LevelError},
		{name: "dpanic", value: "dpanic", want: LevelDPanic},
		{name: "panic", value: "panic", want: LevelPanic},
		{name: "fatal", value: "fatal", want: LevelFatal},
		{name: "invalid", value: "nope", want: LevelUnspecified, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseLevel(tc.value)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected level: got=%v want=%v", got, tc.want)
			}
		})
	}
}
