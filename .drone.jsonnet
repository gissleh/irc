local Pipeline(version, mod) = {
  kind: "pipeline",
  name: version,
  workspace: if mod then {base: "/project", path: "irc/"} else {base: "/go", path: "src/github.com/gissleh/irc"},
  steps: [
    {
      name: "test",
      image: "golang:" + version,
      commands: [
        if mod then "go mod download" else "go get",
        "go test -v ./...",
        "go test -bench ./..."
      ]
    }
  ]
};

[
  Pipeline("1.14", true),
  Pipeline("1.13", true),
  Pipeline("1.12", true),
  Pipeline("1.11", true),
  Pipeline("1.10", false),
  Pipeline("1.9", false),
  Pipeline("1.8", false),
]