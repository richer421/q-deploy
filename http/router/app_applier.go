package router

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type kubectlApplicationApplier struct{}

func NewKubectlApplicationApplier() *kubectlApplicationApplier {
	return &kubectlApplicationApplier{}
}

func (a *kubectlApplicationApplier) Apply(ctx context.Context, manifest []byte) error {
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewReader(manifest)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, stderr.String())
		}
		return err
	}
	return nil
}
