package main

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	err := copySecrets(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func copySecrets(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("cannot get docker client: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(os.Getenv("RH_REG_USR") + ":" + os.Getenv("RH_REG_PWD")))

	config := map[string]any{
		"auths": map[string]any{
			"registry.redhat.io": map[string]string{"auth": auth},
		},
	}

	configData, err := json.Marshal(&config)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	var buff bytes.Buffer
	tw := tar.NewWriter(&buff)
	err = tw.WriteHeader(&tar.Header{
		Name:     "config.json",
		Typeflag: tar.TypeReg,
		Size:     int64(len(configData)),
		Mode:     0600,
	})
	if err != nil {
		return fmt.Errorf("cannot write tar header: %w", err)
	}
	_, err = tw.Write(configData)
	if err != nil {
		return fmt.Errorf("cannot write file to tar: %w", err)
	}
	err = tw.Close()
	if err != nil {
		return fmt.Errorf("cannot close tar: %w", err)
	}

	err = cli.CopyToContainer(ctx, "func-control-plane", "/var/lib/kubelet/", &buff, types.CopyToContainerOptions{})
	if err != nil {
		return fmt.Errorf("cannot copy to container: %w", err)
	}

	return nil
}
