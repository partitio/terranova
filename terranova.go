/*
Copyright The Terranova Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package terranova

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/plans"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
)

// Apply brings the platform to the desired state. It'll destroy the platform
// when `destroy` is `true`.
func (p *Platform) Apply(destroy bool) error {
	ctx, err := p.newContext(destroy)
	if err != nil {
		return err
	}

	// state := ctx.State()

	if _, diag := ctx.Refresh(); diag.HasErrors() {
		return diag.Err()
	}

	if _, diag := ctx.Plan(); diag.HasErrors() {
		return diag.Err()
	}

	sts, diag := ctx.Apply()
	p.State = sts
	// p.State = ctx.State()

	if diag.HasErrors() {
		return diag.Err()
	}
	return nil
}

// Plan returns execution plan for an existing configuration to apply to the
// platform.
func (p *Platform) Plan(destroy bool) (*plans.Plan, error) {
	ctx, err := p.newContext(destroy)
	if err != nil {
		return nil, err
	}

	if _, diag := ctx.Refresh(); diag.HasErrors() {
		return nil, diag.Err()
	}

	plan, diag := ctx.Plan()
	if diag.HasErrors() {
		return nil, diag.Err()
	}

	return plan, nil
}

// newContext creates the Terraform context or configuration
func (p *Platform) newContext(destroy bool) (*terraform.Context, error) {
	cfg, err := p.config()
	if err != nil {
		return nil, err
	}

	vars, err := p.variables(cfg.Module.Variables)
	if err != nil {
		return nil, err
	}

	// providerResolver := providers.ResolverFixed(p.Providers)
	// provisioners := p.Provisioners

	// Create ContextOpts with the current state and variables to apply
	ctxOpts := terraform.ContextOpts{
		Config:           cfg,
		Destroy:          destroy,
		State:            p.State,
		Variables:        vars,
		ProviderResolver: providers.ResolverFixed(p.Providers),
		Provisioners:     p.Provisioners,
	}

	ctx, diags := terraform.NewContext(&ctxOpts)
	if diags.HasErrors() {
		return nil, diags.Err()
	}

	// Validate the context
	if diags = ctx.Validate(); diags.HasErrors() {
		return nil, diags.Err()
	}

	return ctx, nil
}

func (p *Platform) config() (*configs.Config, error) {
	if len(p.Code) == 0 {
		return nil, fmt.Errorf("no code to apply")
	}

	// Get a temporal directory to save the infrastructure code
	cfgPath, err := ioutil.TempDir("", ".terranova")
	if err != nil {
		return nil, err
	}
	// This defer is executed second
	defer os.RemoveAll(cfgPath)

	// Save the infrastructure code
	cfgFileName := filepath.Join(cfgPath, "main.tf")
	cfgFile, err := os.Create(cfgFileName)
	if err != nil {
		return nil, err
	}
	// This defer is executed first
	defer cfgFile.Close()
	if _, err = io.Copy(cfgFile, strings.NewReader(p.Code)); err != nil {
		return nil, err
	}

	loader, err := configload.NewLoader(&configload.Config{
		ModulesDir: filepath.Join(cfgPath, "modules"),
	})
	if err != nil {
		return nil, err
	}

	config, diags := loader.LoadConfig(cfgPath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to load the configuration. %s", diags.Error())
	}

	return config, nil

	// mod, err := module.NewTreeModule("", cfgPath)
	// if err != nil {
	// 	return nil, err
	// }

	// s := module.NewStorage(filepath.Join(cfgPath, "modules"), nil)
	// s.Mode = module.GetModeNone // or module.GetModeGet?

	// if err := mod.Load(s); err != nil {
	// 	return nil, fmt.Errorf("failed to load the modules. %s", err)
	// }

	// if err := mod.Validate().Err(); err != nil {
	// 	return nil, fmt.Errorf("failed Terraform code validation. %s", err)
	// }

	// return mod, nil
}

func (p *Platform) variables(v map[string]*configs.Variable) (terraform.InputValues, error) {
	iv := make(terraform.InputValues)
	for name, value := range p.Vars {
		if _, declared := v[name]; !declared {
			return iv, fmt.Errorf("variable %q is not declared in the code", name)
		}

		val := &terraform.InputValue{
			Value:      value,
			SourceType: terraform.ValueFromCaller,
		}

		iv[name] = val
	}
	return iv, nil
}

// func (p *Platform) getProviderResolver() providers.Resolver {
// 	return providers.ResolverFixed(p.Providers)
// }

// func (p *Platform) getProvisioners() map[string]terraform.ResourceProvisionerFactory {
// 	provisioners := make(map[string]terraform.ResourceProvisionerFactory)

// 	for name, provisioner := range p.Provisioners {
// 		provisioners[name] = func() (terraform.ResourceProvisioner, error) {
// 			return provisioner, nil
// 		}
// 	}

// 	return provisioners
// }
