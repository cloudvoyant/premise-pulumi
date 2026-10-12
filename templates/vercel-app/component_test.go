package vercelapp

import (
	"fmt"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strings"
	"sync"
	"testing"
)

// Public API -----------------------------------------------------------------

func TestTwoApplicationsKeepTheirOwnSourceAliasAndEnvironment(t *testing.T) {
	mocks := &applicationMocks{}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		for _, name := range []string{"web", "admin"} {
			app, err := New(ctx, name, Args{ProjectID: "existing-project", TeamID: "existing-team", Alias: name + ".test.invalid", Source: "../../apps/" + name, Release: "immutable-A", Environment: pulumi.StringMap{"DATABASE_URL": pulumi.String("test-only-connection")}})
			if err != nil {
				return err
			}
			ctx.Export(name, app.Deployment.ID())
		}
		return nil
	}, pulumi.WithMocks("test", "preview-test", mocks))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"web", "admin"} {
		deployment := mocks.resources[name+"-deployment"]
		if !deployment["environment"].IsSecret() {
			t.Fatal("runtime environment is not secret")
		}
		if deployment["pathPrefix"].StringValue() != "../../apps/"+name {
			t.Fatal("source binding crossed apps")
		}
		if mocks.resources[name+"-alias"]["alias"].StringValue() != name+".test.invalid" {
			t.Fatal("alias crossed apps")
		}
	}
}

func TestSourceGuard(t *testing.T) {
	for _, file := range []string{".env.local", "source/.npmrc", ".pulumi/state", "fnox.toml", "node_modules/module/index.js"} {
		t.Run(file, func(t *testing.T) {
			if validateFiles(map[string]string{file: "metadata"}) == nil {
				t.Fatal("unsafe source accepted")
			}
		})
	}
	if validateFiles(nil) == nil {
		t.Fatal("empty source accepted")
	}
	if err := validateFiles(map[string]string{"src/index.ts": "metadata"}); err != nil {
		t.Fatal(err)
	}
}

// Private helpers ------------------------------------------------------------

type applicationMocks struct {
	mu        sync.Mutex
	resources map[string]resource.PropertyMap
}

func (m *applicationMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.resources == nil {
		m.resources = map[string]resource.PropertyMap{}
	}
	if _, exists := m.resources[args.Name]; exists {
		return "", nil, fmt.Errorf("duplicate %s", args.Name)
	}
	m.resources[args.Name] = args.Inputs.Copy()
	return args.Name + "-id", args.Inputs, nil
}
func (*applicationMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	if strings.Contains(args.Token, "getProjectDirectory") {
		return resource.PropertyMap{"path": args.Args["path"], "files": resource.NewObjectProperty(resource.PropertyMap{"package.json": resource.NewStringProperty("fixture-metadata")})}, nil
	}
	return args.Args, nil
}
