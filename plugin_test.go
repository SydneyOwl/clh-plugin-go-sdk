package pluginsdk

import (
	"fmt"
	"github.com/google/uuid"
	"testing"
)

func TestDupeClose(t *testing.T) {
	s := uuid.New().String()
	client, err := NewClient(PluginConfig{
		Uuid:        s,
		Name:        "TestPlugin",
		Version:     "1.0.0",
		Description: "Test",
		Capabilities: []PluginCapability{
			CapabilityWsjtxMessage,
			CapabilityRigData,
		},
	}, nil)

	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	err = client.Close()
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	err = client.Close()
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}
}

func TestDupeConnect(t *testing.T) {
	s := uuid.New().String()
	client, err := NewClient(PluginConfig{
		Uuid:        s,
		Name:        "TestPlugin",
		Version:     "1.0.0",
		Description: "Test",
		Capabilities: []PluginCapability{
			CapabilityWsjtxMessage,
			CapabilityRigData,
		},
	}, nil)

	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}
}

func TestNewClient(t *testing.T) {
	s := uuid.New().String()
	client, err := NewClient(PluginConfig{
		Uuid:        s,
		Name:        "TestPlugin",
		Version:     "1.0.0",
		Description: "Test",
		Capabilities: []PluginCapability{
			CapabilityWsjtxMessage,
			CapabilityRigData,
		},
	}, nil)

	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}

	err = client.Connect()
	if err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	for {
		message, err := client.WaitMessage()
		if err != nil {
			t.Fatalf("Error waiting for message: %s", err)
		}
		fmt.Printf("%v\n", message)
	}
}
