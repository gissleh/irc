local Pipeline(version, mod) = {
  kind: "pipeline",
  name: version,
  workspace: if mod then {root: "/project", path: "irc/"} else {root: "/go", path: "src/github.com/gissleh/irc"},
  steps: [
    {
      name: "test",
      image: "goland:" + version,
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
  Pipeline("1.11", false),
  Pipeline("1.11", false),
  Pipeline("1.11", false),
]