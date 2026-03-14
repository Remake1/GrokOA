package ws

import "testing"

func TestNormalizeServerURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "websocket url", input: "ws://172.16.3.88", want: "ws://172.16.3.88"},
		{name: "secure websocket url", input: "wss://example.com/base/", want: "wss://example.com/base"},
		{name: "http becomes websocket", input: "http://172.16.3.88", want: "ws://172.16.3.88"},
		{name: "https becomes secure websocket", input: "https://example.com/app/", want: "wss://example.com/app"},
		{name: "bare host becomes websocket", input: "172.16.3.88:80", want: "ws://172.16.3.88:80"},
		{name: "query and fragment are removed", input: "http://example.com/path/?a=1#frag", want: "ws://example.com/path"},
		{name: "invalid empty", input: "", wantErr: true},
		{name: "invalid scheme", input: "ftp://example.com", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeServerURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
