package pluginsdk

import (
	"testing"
	"time"
)

func TestE2EClhClient(t *testing.T) {
	config := PluginConfig{
		Uuid:        "test-uuid",
		Name:        "Test Plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Capabilities: []PluginCapability{
			CapabilityWsjtxMessage,
			CapabilityRigData,
		},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	// Test Connect
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Test WaitMessage
	msg, err := client.WaitMessage()
	if err != nil {
		t.Fatalf("WaitMessage failed: %v", err)
	}
	t.Logf("WaitMessage: %v", msg)

	msg, err = client.WaitMessage()
	if err != nil {
		t.Fatalf("WaitMessage failed: %v", err)
	}
	t.Logf("WaitMessage: %v", msg)

	msg, err = client.WaitMessage()
	if err != nil {
		t.Fatalf("WaitMessage failed: %v", err)
	}
	t.Logf("WaitMessage: %v", msg)

	// Test Close
	if err := client.Close(); err != nil {
		t.Errorf("First Close failed: %v", err)
	}

	// Test idempotent Close
	if err := client.Close(); err != nil {
		t.Errorf("Second Close should be safe, but got: %v", err)
	}

	// Test Connect on closed client
	if err := client.Connect(); err == nil {
		t.Error("Expected error when connecting closed client, got nil")
	} else if err.Error() != "you are not allowd to call connect on a closed client. please create a new client instead" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestWithHeartbeatInterval(t *testing.T) {
	config := PluginConfig{
		Uuid:        "test-uuid",
		Name:        "Test",
		Version:     "1.0",
		Description: "test",
	}

	client, err := NewClient(config, WithHeartbeatInterval(3*time.Second))
	if err != nil {
		t.Fatalf("NewClient with option failed: %v", err)
	}
	if client.heartbeatInterval != 3*time.Second {
		t.Errorf("Expected 3s heartbeat, got %v", client.heartbeatInterval)
	}

	// Test invalid interval
	_, err = NewClient(config, WithHeartbeatInterval(15*time.Second))
	if err == nil {
		t.Error("Expected error for invalid heartbeat interval")
	}
}

func TestNewClientValidation(t *testing.T) {
	testCases := []struct {
		name    string
		config  PluginConfig
		wantErr bool
	}{
		{
			name:    "missing uuid",
			config:  PluginConfig{Name: "x", Version: "1", Description: "d"},
			wantErr: true,
		},
		{
			name:    "missing name",
			config:  PluginConfig{Uuid: "u", Version: "1", Description: "d"},
			wantErr: true,
		},
		{
			name:    "missing version",
			config:  PluginConfig{Uuid: "u", Name: "n", Description: "d"},
			wantErr: true,
		},
		{
			name:    "missing description",
			config:  PluginConfig{Uuid: "u", Name: "n", Version: "1"},
			wantErr: true,
		},
		{
			name:    "valid",
			config:  PluginConfig{Uuid: "u", Name: "n", Version: "1", Description: "d"},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewClient(tc.config)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
