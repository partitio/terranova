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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/logutils"
	"github.com/hashicorp/terraform/helper/logging"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/provisioners"
	"github.com/hashicorp/terraform/states"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-null/null"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// Platform is the platform to be managed by Terraform
type Platform struct {
	Code         string
	Providers    map[string]providers.Factory
	Provisioners map[string]provisioners.Factory
	Vars         map[string]cty.Value
	State        *State
	Error        error
}

// State is an alias for terraform.State
type State = states.State

// NewPlatform return an instance of Platform with default values
func NewPlatform(code string) *Platform {
	platform := &Platform{
		Code: code,
	}
	platform.addDefaultProviders()
	platform.addDefaultProvisioners()

	platform.State = states.NewState()

	return platform
}

func (p *Platform) AddCode(code string) *Platform {
	p.appendCode(code)
	return p
}

func (p *Platform) AddCodeFiles(path ...string) *Platform {
	c, err := loadCodeFiles(path...)
	if err != nil {

		return p
	}
	p.appendCode(c)
	return p
}

func (p *Platform) appendCode(code string) {
	p.Code += "\n" + code
}

func (p *Platform) setLastError(err error) {
	p.Error = fmt.Errorf("last error: %v", err)
}

func loadCodeFiles(files ...string) (string, error) {
	parse := func(file string) ([]byte, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return ioutil.ReadAll(f)
	}
	var b []byte
	for _, file := range files {
		f, err := parse(file)
		if err != nil {
			return "", err
		}
		b = append(b, f...)
	}
	return string(b), nil
}

func (p *Platform) addDefaultProviders() {
	p.Providers = map[string]providers.Factory{}
	p.AddProvider("null", null.Provider())
}

// AddProvider adds a new provider to the providers list
func (p *Platform) AddProvider(name string, provider terraform.ResourceProvider) *Platform {
	p.Providers[name] = providersFactory(provider)
	return p
}

func (p *Platform) addDefaultProvisioners() {
	p.Provisioners = map[string]provisioners.Factory{}
}

// AddProvisioner adds a new provisioner to the provisioner list
func (p *Platform) AddProvisioner(name string, provisioner terraform.ResourceProvisioner) *Platform {
	p.Provisioners[name] = provisionersFactory(provisioner)
	return p
}

// BindVars binds the map of variables to the Platform variables, to be used
// by Terraform
func (p *Platform) BindVars(vars interface{}) *Platform {
	t, err := gocty.ImpliedType(vars)
	if err != nil {
		p.setLastError(err)
		return p
	}
	v, err := gocty.ToCtyValue(vars, t)
	if err != nil {
		p.setLastError(err)
		return p
	}
	p.Vars = v.AsValueMap()
	return p
}

// WriteState takes a io.Writer as input to write the Terraform state
func (p *Platform) WriteState(w io.Writer) (*Platform, error) {
	sf := statefile.New(p.State, "", 0)
	return p, statefile.Write(sf, w)
}

// ReadState takes a io.Reader as input to read from it the Terraform state
func (p *Platform) ReadState(r io.Reader) (*Platform, error) {
	sf, err := statefile.Read(r)
	if err != nil {
		return p, err
	}
	p.State = sf.State
	return p, nil
}

// WriteStateToFile save the state of the Terraform state to a file
func (p *Platform) WriteStateToFile(filename string) (*Platform, error) {
	var state bytes.Buffer
	if _, err := p.WriteState(&state); err != nil {
		return p, err
	}
	return p, ioutil.WriteFile(filename, state.Bytes(), 0644)
}

// ReadStateFromFile will load the Terraform state from a file and assign it to the
// Platform state.
func (p *Platform) ReadStateFromFile(filename string) (*Platform, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return p, err
	}
	return p.ReadState(file)
}

func (p *Platform) SetLogLevel(level string) *Platform {
	logLevel := logutils.LogLevel(strings.ToUpper(level))
	ll := logutils.LogLevel("INFO")
	for _, l := range logging.ValidLevels {
		if logLevel == l {
			ll = logLevel
			break
		}
	}
	log.SetOutput(&logutils.LevelFilter{
		Levels:   logging.ValidLevels,
		MinLevel: ll,
		Writer:   os.Stdout,
	})
	return p
}
