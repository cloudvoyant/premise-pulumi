package main

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Private helpers ------------------------------------------------------------

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Import the generated project components here and connect their outputs.
		ctx.Export("environment", pulumi.String(ctx.Stack()))
		return nil
	})
}
