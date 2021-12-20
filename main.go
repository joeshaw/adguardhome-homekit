package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hclog "github.com/brutella/hc/log"
)

const (
	statusEndpoint    = "/control/status"
	dnsConfigEndpoint = "/control/dns_config"
)

type Config struct {
	// Storage path for information about the HomeKit accessory.
	// Defaults to ~/.homecontrol
	StoragePath string `json:"storage_path"`

	// HomeKit PIN.  Defaults to 00102003
	HomekitPIN string `json:"homekit_pin"`

	// AdGuard Home URL
	URL string `json:"url"`

	// AdGuard Home username
	Username string `json:"username"`

	// AdGuard Home password
	Password string `json:"password"`
}

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", "config.json", "config file")
	flag.Parse()

	// Default values
	cfg := Config{
		StoragePath: filepath.Join(os.Getenv("HOME"), ".homecontrol"),
		HomekitPIN:  "00102003",
	}

	f, err := os.Open(configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	if cfg.URL == "" {
		log.Fatal("missing URL")
	}

	if cfg.Username == "" {
		log.Fatal("missing username")
	}

	if cfg.Password == "" {
		log.Fatal("missing password")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if x := os.Getenv("HC_DEBUG"); x != "" {
		hclog.Debug.Enable()
	}

	enabled, err := ProtectionEnabled(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to AdGuard Home, protection enabled: %t", enabled)

	info := accessory.Info{
		Name: "AdGuard Home",
	}

	sw := accessory.NewSwitch(info)
	sw.Switch.On.OnValueRemoteUpdate(func(on bool) {
		SetProtectionEnabled(ctx, &cfg, on)
	})

	hcConfig := hc.Config{
		Pin:         cfg.HomekitPIN,
		StoragePath: cfg.StoragePath,
	}
	t, err := hc.NewIPTransport(hcConfig, sw.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		cancel()
		<-t.Stop()
	})

	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				enabled, err := ProtectionEnabled(ctx, &cfg)
				if err != nil {
					log.Printf("error checking protection enabled: %v", err)
					continue
				}
				sw.Switch.On.SetValue(enabled)
			}
		}
	}()

	log.Println("Starting transport...")
	t.Start()
}

func ProtectionEnabled(ctx context.Context, cfg *Config) (bool, error) {
	req, err := http.NewRequest("GET", cfg.URL+statusEndpoint, nil)
	if err != nil {
		return false, err
	}

	req.SetBasicAuth(cfg.Username, cfg.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var status struct {
		ProtectionEnabled bool `json:"protection_enabled"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return false, err
	}

	return status.ProtectionEnabled, nil
}

func SetProtectionEnabled(ctx context.Context, cfg *Config, enabled bool) error {
	payload := fmt.Sprintf(`{"protection_enabled": %t}`, enabled)
	req, err := http.NewRequest(
		"POST",
		cfg.URL+dnsConfigEndpoint,
		strings.NewReader(payload),
	)
	if err != nil {
		return err
	}

	req.SetBasicAuth(cfg.Username, cfg.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
