package main

import (
	"context"

	"go.dagger.io/dagger/sdk/go/dagger"
)

func main() {
	dagger.Serve(
		Alpine{},
	)
}

type Alpine struct {
}

func (a Alpine) Build(ctx dagger.Context, pkgs []string) (*dagger.Filesystem, error) {
	client, err := dagger.Connect(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// start with Alpine base
	fsid, err := image(ctx, client, "alpine:3.15")
	if err != nil {
		return nil, err
	}

	// install each of the requested packages
	for _, pkg := range pkgs {
		fsid, err = addPkg(ctx, client, fsid, pkg)
		if err != nil {
			return nil, err
		}
	}
	return &dagger.Filesystem{ID: fsid}, nil
}

func image(ctx context.Context, client *dagger.Client, ref string) (dagger.FSID, error) {
	req := &dagger.Request{
		Query: `
query Image ($ref: String!) {
	core {
		image(ref: $ref) {
			id
		}
	}
}
`,
		Variables: map[string]any{
			"ref": ref,
		},
	}
	resp := struct {
		Core struct {
			Image struct {
				ID dagger.FSID
			}
		}
	}{}
	err := client.Do(ctx, req, &dagger.Response{Data: &resp})
	if err != nil {
		return "", err
	}

	return resp.Core.Image.ID, nil
}

func addPkg(ctx context.Context, client *dagger.Client, root dagger.FSID, pkg string) (dagger.FSID, error) {
	req := &dagger.Request{
		Query: `
query AddPkg ($root: FSID!, $pkg: String!) {
	core {
		filesystem(id: $root) {
			exec(input: {
				args: ["apk", "add", "-U", "--no-cache", $pkg]
			}) {
				fs {
					id
				}
			}
		}
	}
}
`,
		Variables: map[string]any{
			"root": root,
			"pkg":  pkg,
		},
	}
	resp := struct {
		Core struct {
			Filesystem struct {
				Exec struct {
					FS struct {
						ID dagger.FSID
					}
				}
			}
		}
	}{}
	err := client.Do(ctx, req, &dagger.Response{Data: &resp})
	if err != nil {
		return "", err
	}

	return resp.Core.Filesystem.Exec.FS.ID, nil
}
