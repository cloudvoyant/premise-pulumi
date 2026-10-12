package neondb

import (
	"fmt"
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Public API -----------------------------------------------------------------

func TestTwoDatabasesHaveIndependentResourcesAndSecretOutputs(t *testing.T) {
	mocks := &databaseMocks{}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		for _, name := range []string{"db-web", "db-admin"} {
			db, err := New(ctx, name, Args{ProjectID: "existing-project", ParentID: "existing-parent", Stack: "preview-test"})
			if err != nil {
				return err
			}
			ctx.Export(name, db.URL)
		}
		return nil
	}, pulumi.WithMocks("test", "preview-test", mocks))
	if err != nil {
		t.Fatal(err)
	}
	if len(mocks.names) != 8 {
		t.Fatalf("registered %d provider resources, want 8", len(mocks.names))
	}
	for _, name := range []string{"db-web-branch", "db-admin-branch"} {
		branch := mocks.names[name]
		if branch["projectId"].StringValue() != "existing-project" || branch["name"].StringValue() != "premise-preview-test-"+name[:len(name)-7] {
			t.Fatal("database identity lost")
		}
	}
}

// Private helpers ------------------------------------------------------------

type databaseMocks struct {
	mu    sync.Mutex
	names map[string]resource.PropertyMap
}

func (m *databaseMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.names == nil {
		m.names = map[string]resource.PropertyMap{}
	}
	state := args.Inputs.Copy()
	if args.Custom {
		if _, exists := m.names[args.Name]; exists {
			return "", nil, fmt.Errorf("duplicate resource %s", args.Name)
		}
		m.names[args.Name] = state
	}
	state["password"] = resource.MakeSecret(resource.NewStringProperty("mock:password@only"))
	state["host"] = resource.NewStringProperty("mock.neon.invalid")
	return args.Name + "-id", state, nil
}
func (*databaseMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}
