package builder

import (
	"fmt"
	"strings"
	"unicode/utf8"

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

const (
	bannerBorder      = "*******************************************************************"
	defaultBannerName = "GO COMMON BUILDER"
)

const bannerLogo = `%s` + bannerBorder + `
` + bannerBorder + `
%s
` + bannerBorder + `
******************** Compile Environment **************************
*** Program Name     : %s
*** Program Version  : %s
*** Program Branch   : %s
*** Program Revision : %s
*** Compiler Version : %s
*** Build Time       : %s
*** Author           : %s
` + bannerBorder + `
******************** Running Environment **************************
*** GO ROOT            : {{ .GOROOT }}
*** Go running version : {{ .GoVersion }}
*** Go compiler        : {{ .Compiler }}
*** Go running OS      : {{ .GOOS }} {{ .GOARCH }}
*** Go CPU Numbers     : {{ .NumCPU }}
*** Startup time       : {{ .Now "2006-01-02 15:04:05 (Monday)" }}
` + bannerBorder + `
` + bannerBorder + `
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

func OffShow() Option {
	return func(o *Options) {
		o.OnShow = false
	}
}

func OnColor() Option {
	return func(o *Options) {
		o.OnColor = true
	}
}

func formatBannerTitle(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultBannerName
	}

	inner := utf8.RuneCountInString(bannerBorder) - 6
	runes := []rune(name)
	if len(runes) > inner {
		runes = runes[:inner]
		name = string(runes)
	}
	left := (inner - len(runes)) / 2
	right := inner - len(runes) - left
	return "***" + strings.Repeat(" ", left) + name + strings.Repeat(" ", right) + "***"
}

// Show displays project information. Printing is on by default; pass OffShow to disable.
func Show(opts ...Option) {
	options := &Options{OnShow: true}
	for _, o := range opts {
		o(options)
	}

	if options.Color == "" {
		options.Color = "{{ .AnsiColor.Default }}"
	}

	newBanner := fmt.Sprintf(bannerLogo, options.Color, formatBannerTitle(ProgramName),
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
