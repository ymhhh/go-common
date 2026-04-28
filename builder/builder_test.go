package builder

import "testing"

func withBuildVars(t *testing.T, fn func()) {
	t.Helper()

	prevName := ProgramName
	prevVer := ProgramVersion
	prevBranch := ProgramBranch
	prevRev := ProgramRevision
	prevGo := CompilerVersion
	prevTime := BuildTime
	prevAuthor := Author

	t.Cleanup(func() {
		ProgramName = prevName
		ProgramVersion = prevVer
		ProgramBranch = prevBranch
		ProgramRevision = prevRev
		CompilerVersion = prevGo
		BuildTime = prevTime
		Author = prevAuthor
	})

	fn()
}

func TestOptions(t *testing.T) {
	o := &Options{}
	Color("c")(o)
	OnShow()(o)
	OnColor()(o)

	if o.Color != "c" {
		t.Fatalf("Color: got %q", o.Color)
	}
	if !o.OnShow {
		t.Fatalf("OnShow: expected true")
	}
	if !o.OnColor {
		t.Fatalf("OnColor: expected true")
	}
}

func TestVersion(t *testing.T) {
	withBuildVars(t, func() {
		ProgramName = "p"
		ProgramVersion = "1.2.3"
		ProgramBranch = "main"
		ProgramRevision = "abc"

		got := Version()
		want := "p, version: 1.2.3 (branch: main, revision: abc)"
		if got != want {
			t.Fatalf("Version(): got %q, want %q", got, want)
		}
	})
}

func TestBuildInfo(t *testing.T) {
	withBuildVars(t, func() {
		CompilerVersion = "go1.22.0"
		Author = "me"
		BuildTime = "2026-04-29"

		got := BuildInfo()
		want := "(go=go1.22.0, user=me, date=2026-04-29)"
		if got != want {
			t.Fatalf("BuildInfo(): got %q, want %q", got, want)
		}
	})
}

func TestShow_NoPanic(t *testing.T) {
	withBuildVars(t, func() {
		ProgramName = "p"
		ProgramVersion = "v"
		ProgramBranch = "b"
		ProgramRevision = "r"
		CompilerVersion = "go"
		BuildTime = "t"
		Author = "a"

		// banner.Init writes to stdout; we only assert it doesn't panic.
		Show()
		Show(Color("{{ .AnsiColor.Cyan }}"))
		Show(OnShow(), OnColor())
	})
}

