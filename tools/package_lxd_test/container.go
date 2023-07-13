package main

import (
	"fmt"
	"math"
	"path/filepath"
	"time"
)

const influxDataRPMRepo = `[influxdata]
name = InfluxData Repository - Stable
baseurl = https://repos.influxdata.com/stable/x86_64/main
enabled = 1
gpgcheck = 1
gpgkey = https://repos.influxdata.com/influxdata-archive_compat.key
`

type Container struct {
	Name string

	client         LXDClient
	packageManager string
}

// create contianer with given name and image
func (c *Container) Create(image string) error {
	if c.Name == "" {
		return fmt.Errorf("unable to create container: no name given")
	}

	c.client = LXDClient{}
	err := c.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to lxd: %w", err)
	}

	err = c.client.Create(c.Name, "images", image)
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	// at this point the container is created, so on any error during setup
	// we want to delete it as well
	err = c.client.Start(c.Name)
	if err != nil {
		c.Delete()
		return fmt.Errorf("failed to start instance: %w", err)
	}

	if err := c.detectPackageManager(); err != nil {
		c.Delete()
		return err
	}

	if err := c.waitForNetwork(); err != nil {
		c.Delete()
		return err
	}

	if err := c.setupRepo(); err != nil {
		c.Delete()
		return err
	}

	return nil
}

// delete the container
func (c *Container) Delete() {
	_ = c.client.Stop(c.Name)
	_ = c.client.Delete(c.Name)
}

// installs the package from configured repos
func (c *Container) Install(packageName ...string) error {
	var cmd []string
	switch c.packageManager {
	case "apt":
		cmd = append([]string{"apt-get", "install", "--yes"}, packageName...)
	case "yum":
		cmd = append([]string{"yum", "install", "-y"}, packageName...)
	case "dnf":
		cmd = append([]string{"dnf", "install", "-y"}, packageName...)
	case "zypper":
		cmd = append([]string{"zypper", "install", "-y"}, packageName...)
	}

	err := c.client.Exec(c.Name, cmd...)
	if err != nil {
		return err
	}

	return nil
}

func (c *Container) CheckStatus(serviceName string) error {
	// push a valid config first, then start the service
	err := c.client.Exec(
		c.Name,
		"bash",
		"-c",
		"--",
		"echo '[[inputs.cpu]]\n[[outputs.file]]' | "+
			"tee /etc/telegraf/telegraf.conf",
	)
	if err != nil {
		return err
	}

	err = c.client.Exec(c.Name, "systemctl", "start", serviceName)
	if err != nil {
		_ = c.client.Exec(c.Name, "systemctl", "status", serviceName)
		_ = c.client.Exec(c.Name, "journalctl", "--no-pager", "--unit", serviceName)
		return err
	}

	err = c.client.Exec(c.Name, "systemctl", "status", serviceName)
	if err != nil {
		_ = c.client.Exec(c.Name, "journalctl", "--no-pager", "--unit", serviceName)
		return err
	}

	return nil
}

func (c *Container) UploadAndInstall(filename string) error {
	basename := filepath.Base(filename)
	destination := fmt.Sprintf("/root/%s", basename)

	if err := c.client.Push(c.Name, filename, destination); err != nil {
		return err
	}

	return c.Install(destination)
}

// Push key and config and update
func (c *Container) configureApt() error {
	err := c.client.Exec(c.Name, "apt-get", "update")
	if err != nil {
		return err
	}

	err = c.Install("ca-certificates", "gpg", "wget")
	if err != nil {
		return err
	}

	err = c.client.Exec(c.Name, "wget", "https://repos.influxdata.com/influxdata-archive_compat.key")
	if err != nil {
		return err
	}

	err = c.client.Exec(
		c.Name,
		"bash",
		"-c",
		"--",
		"echo '393e8779c89ac8d958f81f942f9ad7fb82a25e133faddaf92e15b16e6ac9ce4c influxdata-archive_compat.key' | "+
			"sha256sum -c && cat influxdata-archive_compat.key | gpg --dearmor | "+
			"sudo tee /etc/apt/trusted.gpg.d/influxdata-archive_compat.gpg > /dev/null",
	)
	if err != nil {
		return err
	}

	err = c.client.Exec(
		c.Name,
		"bash",
		"-c",
		"--",
		"echo 'deb [signed-by=/etc/apt/trusted.gpg.d/influxdata-archive_compat.gpg] https://repos.influxdata.com/debian stable main' | "+
			"tee /etc/apt/sources.list.d/influxdata.list",
	)
	if err != nil {
		return err
	}

	_ = c.client.Exec(
		c.Name,
		"bash", "-c", "--",
		"cat /etc/apt/sources.list.d/influxdata.list",
	)

	err = c.client.Exec(c.Name, "apt-get", "update")
	if err != nil {
		return err
	}

	return nil
}

// Create config and update yum
func (c *Container) configureYum() error {
	err := c.client.Exec(
		c.Name,
		"bash", "-c", "--",
		fmt.Sprintf("echo -e %q > /etc/yum.repos.d/influxdata.repo", influxDataRPMRepo),
	)
	if err != nil {
		return err
	}

	_ = c.client.Exec(
		c.Name,
		"bash", "-c", "--",
		"cat /etc/yum.repos.d/influxdata.repo",
	)

	// will return a non-zero return code if there are packages to update
	return c.client.Exec(c.Name, "bash", "-c", "yum check-update || true")
}

// Create config and update dnf
func (c *Container) configureDnf() error {
	err := c.client.Exec(
		c.Name,
		"bash", "-c", "--",
		fmt.Sprintf("echo -e %q > /etc/yum.repos.d/influxdata.repo", influxDataRPMRepo),
	)
	if err != nil {
		return err
	}

	_ = c.client.Exec(
		c.Name,
		"bash", "-c", "--",
		"cat /etc/yum.repos.d/influxdata.repo",
	)

	// will return a non-zero return code if there are packages to update
	return c.client.Exec(c.Name, "bash", "-c", "dnf check-update || true")
}

// Create config and update zypper
func (c *Container) configureZypper() error {
	err := c.client.Exec(
		c.Name,
		"echo", fmt.Sprintf("%q", influxDataRPMRepo), ">", "/etc/zypp/repos.d/influxdata.repo",
	)
	if err != nil {
		return err
	}

	_ = c.client.Exec(
		c.Name,
		"bash", "-c", "--",
		"cat /etc/zypp/repos.d/influxdata.repo",
	)

	return c.client.Exec(c.Name, "zypper", "refresh")
}

// Determine if the system uses yum or apt for software
func (c *Container) detectPackageManager() error {
	// Different options required across the distros as apt returns -1 when
	// run with no options. yum is listed last to prefer the newer
	// options first.
	err := c.client.Exec(c.Name, "which", "apt")
	if err == nil {
		c.packageManager = "apt"
		return nil
	}

	err = c.client.Exec(c.Name, "dnf")
	if err == nil {
		c.packageManager = "dnf"
		return nil
	}

	err = c.client.Exec(c.Name, "yum", "version")
	if err == nil {
		c.packageManager = "yum"
		return nil
	}

	return fmt.Errorf("unable to determine package manager")
}

// Configure the system with InfluxData repo
func (c *Container) setupRepo() error {
	if c.packageManager == "apt" {
		if err := c.configureApt(); err != nil {
			return err
		}
	} else if c.packageManager == "yum" {
		if err := c.configureYum(); err != nil {
			return err
		}
	} else if c.packageManager == "zypper" {
		if err := c.configureZypper(); err != nil {
			return err
		}
	} else if c.packageManager == "dnf" {
		if err := c.configureDnf(); err != nil {
			return err
		}
	}

	return nil
}

// Wait for the network to come up on a container
func (c *Container) waitForNetwork() error {
	var exponentialBackoffCeilingSecs int64 = 128

	attempts := 0
	for {
		if err := c.client.Exec(c.Name, "getent", "hosts", "influxdata.com"); err == nil {
			return nil
		}

		// uses exponetnial backoff to try after 1, 2, 4, 8, 16, etc. seconds
		delaySecs := int64(math.Pow(2, float64(attempts)))
		if delaySecs > exponentialBackoffCeilingSecs {
			break
		}

		fmt.Printf("waiting for network, sleeping for %d second(s)\n", delaySecs)
		time.Sleep(time.Duration(delaySecs) * time.Second)
		attempts++
	}

	return fmt.Errorf("timeout reached waiting for network on container")
}
