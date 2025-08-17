package internal

import (
	"context"
	"fmt"
	"slices"
	"time"
)

// Device handles all device operations
type Device struct {
	client  *NanoleafClient
	config  Config
	effects []string
}

func NewDevice() *Device {
	return &Device{
		client: newClient(),
	}
}

func (d *Device) LoadConfig() error {
	if !configExists() {
		return fmt.Errorf("no config found")
	}
	config, err := loadConfig()
	if err != nil {
		return err
	}
	d.config = config
	return nil
}

func (d *Device) IsDeviceReady(ctx context.Context) bool {
	if d.config.IP == "" || d.config.Token == "" {
		return false
	}
	_, err := d.client.getInfo(ctx, d.config.IP, d.config.Token)
	if err == nil {
		d.loadEffects(ctx)
		return true
	}
	return false
}

func (d *Device) loadEffects(ctx context.Context) {
	if effects, err := d.client.listEffects(ctx, d.config.IP, d.config.Token); err == nil {
		d.effects = effects
	}
}

func (d *Device) ListEffects(ctx context.Context) ([]string, error) {
	return d.client.listEffects(ctx, d.config.IP, d.config.Token)
}

func (d *Device) ScanForDevices(ctx context.Context) ([]string, error) {
	return scanForDevices(ctx)
}

func (d *Device) SetDevice(ip string) {
	d.config.IP = ip
}

func (d *Device) PairDevice(ctx context.Context) error {
	if d.config.IP == "" {
		return fmt.Errorf("no device IP set")
	}

	token, err := d.client.pair(ctx, d.config.IP)
	if err != nil {
		return err
	}

	d.config.Token = token
	return saveConfig(d.config.IP, d.config.Token)
}

func (d *Device) TurnOn(ctx context.Context) error {
	return d.client.setPower(ctx, d.config.IP, d.config.Token, true)
}

func (d *Device) TurnOff(ctx context.Context) error {
	return d.client.setPower(ctx, d.config.IP, d.config.Token, false)
}

func (d *Device) SetBrightness(ctx context.Context, brightness int) error {
	if brightness < 0 || brightness > 100 {
		return fmt.Errorf("brightness must be between 0 and 100")
	}
	return d.client.setBrightness(ctx, d.config.IP, d.config.Token, brightness)
}

func (d *Device) SetEffect(ctx context.Context, effect string) error {
	if !slices.Contains(d.effects, effect) {
		return fmt.Errorf("device does not have effect %s", effect)
	}
	return d.client.setEffect(ctx, d.config.IP, d.config.Token, effect)
}

func (d *Device) GetDeviceIP() string {
	return d.config.IP
}

func (d *Device) createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
