package builder

import (
	"fmt"
	"strings"

	"github.com/dimiro1/banner"
	"github.com/mattn/go-colorable"
)

// Build information
var (
	ProgramName     string
	ProgramVersion  string
	ProgramBranch   string
	ProgramRevision string
	CompilerVersion string
	BuildTime       string
	Author          string
)

const bannerLogo = `%s*******************************************************************
*******************************************************************
***                YMHHH GO COMMON BUILDER                      ***
*******************************************************************
******************** Compile Environment **************************
*** Program Name     : %s
*** Program Version  : %s
*** Program Branch   : %s
*** Program Revision : %s
*** Compiler Version : %s
*** Build Time       : %s
*** Author           : %s
*******************************************************************
******************** Running Environment **************************
*** GO ROOT            : {{ .GOROOT }}
*** Go running version : {{ .GoVersion }}
*** Go compiler        : {{ .Compiler }}
*** Go running OS      : {{ .GOOS }} {{ .GOARCH }}
*** Go CPU Numbers     : {{ .NumCPU }}
*** Startup time       : {{ .Now "2006-01-02 15:04:05 (Monday)" }}
*******************************************************************
*******************************************************************
`

type Option func(*Options)
type Options struct {
	Color   string
	OnShow  bool
	OnColor bool
}

func Color(c string) Option {
	return func(o *Options) {
		o.Color = c
	}
}

func OnShow() Option {
	return func(o *Options) {
		o.OnShow = true
	}
}

func OnColor() Option {
	return func(o *Options) {
		o.OnColor = true
	}
}

// Show displays project information
func Show(opts ...Option) {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.Color == "" {
		options.Color = "{{ .AnsiColor.Default }}"
	}

	newBanner := fmt.Sprintf(bannerLogo, options.Color,
		ProgramName, ProgramVersion,
		ProgramBranch, ProgramRevision,
		CompilerVersion, BuildTime, Author)

	banner.Init(colorable.NewColorableStdout(), options.OnShow, options.OnColor, strings.NewReader(newBanner))
}

// Version info of the program.
func Version() string {
	return fmt.Sprintf("%s, version: %s (branch: %s, revision: %s)",
		ProgramName, ProgramVersion, ProgramBranch, ProgramRevision,
	)
}

// BuildInfo returns build information.
func BuildInfo() string {
	return fmt.Sprintf("(go=%s, user=%s, date=%s)", CompilerVersion, Author, BuildTime)
}
