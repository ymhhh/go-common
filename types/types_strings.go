package types

import (
	"flag"
	"fmt"

	"gopkg.in/yaml.v3"
)

var (
	_ flag.Value       = (*Secret)(nil)
	_ yaml.Marshaler   = (*Secret)(nil)
	_ yaml.Unmarshaler = (*Secret)(nil)

	_ flag.Value       = (*Strings)(nil)
	_ yaml.Marshaler   = (*Strings)(nil)
	_ yaml.Unmarshaler = (*Strings)(nil)
)

// Hidden is the hidden value for Secret.
const Hidden = "<hidden>"

// Secret is a secret value.
type Secret string

// String implements the fmt.Stringer interface for Secret.
func (p Secret) String() string {
	return Hidden
}

// Set implements flag.Value
func (p *Secret) Set(s string) error {
	*p = Secret(s)
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface for Secret.
func (p Secret) MarshalYAML() (any, error) {
	return Hidden, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Secret.
func (p *Secret) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Value == "" {
		*p = Secret("")
		return nil
	}
	*p = Secret(value.Value)
	return nil
}

// Strings array string
type Strings []string

// String implements flag.Value
func (x Strings) String() string {
	return fmt.Sprintf("%s", []string(x))
}

// Set implements flag.Value
func (x *Strings) Set(s string) error {
	*x = append(*x, s)
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface for Strings.
func (x Strings) MarshalYAML() (any, error) {
	return []string(x), nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Strings.
func (x *Strings) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		*x = nil
		return nil
	}
	switch value.Kind {
	case yaml.ScalarNode:
		if value.Value == "" {
			*x = Strings{}
			return nil
		}
		*x = Strings{value.Value}
		return nil
	case yaml.SequenceNode:
		var ss []string
		if err := value.Decode(&ss); err != nil {
			return err
		}
		*x = Strings(ss)
		return nil
	default:
		var ss []string
		if err := value.Decode(&ss); err != nil {
			return err
		}
		*x = Strings(ss)
		return nil
	}
}
