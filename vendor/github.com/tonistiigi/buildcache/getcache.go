package buildcache

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/docker/distribution/digest"
	engineapi "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types/versions"
	"golang.org/x/net/context"
)

type buildCache struct {
	client *engineapi.Client
}

func New(client *engineapi.Client) *buildCache {
	return &buildCache{
		client: client,
	}
}

func (b *buildCache) Get(ctx context.Context, graphdir, image string) (io.ReadCloser, error) {
	id, err := b.getImageID(ctx, image)
	if err != nil {
		return nil, err
	}
	info, err := b.client.Info(ctx)
	if err != nil {
		return nil, err
	}
	if graphdir == "" {
		graphdir = info.DockerRootDir
	}
	imagedir := filepath.Join(graphdir, "image", info.Driver)

	if _, err := os.Stat(filepath.Join(imagedir, "imagedb/content/sha256", id.Hex())); err != nil {
		if os.IsNotExist(err) {
			return b.GetWithRemoteAPI(ctx, image)
		}
	}
	pc, err := b.getParentChain(ctx, imagedir, id)
	if err != nil {
		return nil, err
	}
	if err := validateParentChain(pc); err != nil {
		return nil, err
	}

	return b.writeCacheTar(ctx, pc), nil
}

func (b *buildCache) GetWithRemoteAPI(ctx context.Context, image string) (io.ReadCloser, error) {
	v, err := b.client.ServerVersion(ctx)
	if err != nil {
		return nil, err
	}

	if versions.LessThan(v.Version, "1.11.0") {
		return nil, fmt.Errorf("Buildcache needs at least Docker version v1.11")
	}

	if versions.LessThan(v.Version, "1.12.0") {
		logrus.Warnf("Docker versions before v1.12.0 have a bug causing extracting build cache through remote API to take very long time and use lots of disk space. Please consider upgrading before using this tool.")
	}

	id, err := b.getImageID(ctx, image)
	if err != nil {
		return nil, err
	}

	ids, err := b.getParentIDS(ctx, id)
	if err != nil {
		return nil, err
	}

	rc, err := b.client.ImageSave(ctx, ids)
	if err != nil {
		return nil, err
	}

	return b.filterSaveArchive(rc), nil
}

func (b *buildCache) getParentIDS(ctx context.Context, id digest.Digest) ([]string, error) {
	inspect, _, err := b.client.ImageInspectWithRaw(ctx, string(id))
	if err != nil {
		return nil, err
	}
	out := []string{string(id)}
	if inspect.Parent != "" {
		parent, err := digest.ParseDigest(inspect.Parent)
		if err != nil {
			return nil, err
		}
		rest, err := b.getParentIDS(ctx, parent)
		if err != nil {
			return nil, err
		}
		out = append(out, rest...)
	}
	return out, nil
}

func (b *buildCache) filterSaveArchive(in io.ReadCloser) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		gz := gzip.NewWriter(pw)
		tarReader := tar.NewReader(in)
		tarWriter := tar.NewWriter(gz)

		defer in.Close()

		blacklist := regexp.MustCompile("^[0-9a-f]{64}/")

		for {
			hdr, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				pw.CloseWithError(err)
				return
			}

			if blacklist.MatchString(hdr.Name) {
				_, err := io.Copy(ioutil.Discard, tarReader)
				if err != nil {
					pw.CloseWithError(err)
					return
				}
				continue
			}

			if err := tarWriter.WriteHeader(hdr); err != nil {
				pw.CloseWithError(err)
				return
			}

			if _, err := io.Copy(tarWriter, tarReader); err != nil {
				pw.CloseWithError(err)
				return
			}
		}

		if err := tarWriter.Close(); err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := gz.Close(); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()
	}()
	return pr
}

func (b *buildCache) writeCacheTar(ctx context.Context, imgs []image) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		gz := gzip.NewWriter(pw)
		archive := tar.NewWriter(gz)
		var mfst []manifestRow
		for _, img := range imgs {
			if ctx.Err() != nil {
				pw.CloseWithError(ctx.Err())
			}
			if err := archive.WriteHeader(&tar.Header{
				Name: img.id.Hex() + ".json",
				Size: int64(len(img.raw)),
				Mode: 0444,
			}); err != nil {
				pw.CloseWithError(err)
				return
			}
			if _, err := archive.Write(img.raw); err != nil {
				pw.CloseWithError(err)
				return
			}
			mfst = append(mfst, manifestRow{
				Config: img.id.Hex() + ".json",
				Parent: img.parent.String(),
				Layers: img.layers,
			})
		}
		mfstData, err := json.Marshal(mfst)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := archive.WriteHeader(&tar.Header{
			Name: "manifest.json",
			Size: int64(len(mfstData)),
			Mode: 0444,
		}); err != nil {
			pw.CloseWithError(err)
			return
		}
		if _, err := archive.Write(mfstData); err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := archive.Close(); err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := gz.Close(); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()
	}()
	go func() {
		<-ctx.Done()
		pw.CloseWithError(ctx.Err())
	}()
	return pr
}

func (b *buildCache) getImageID(ctx context.Context, ref string) (digest.Digest, error) {
	inspect, _, err := b.client.ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return "", err
	}
	return digest.ParseDigest(inspect.ID)
}

func (b *buildCache) getParentChain(ctx context.Context, dir string, id digest.Digest) ([]image, error) {
	config, err := ioutil.ReadFile(filepath.Join(dir, "imagedb/content/sha256", id.Hex()))
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	img, err := parseImage(config)
	if err != nil {
		return nil, err
	}
	if img.id != id {
		return nil, fmt.Errorf("invalid configuration for %v, got id %v", id, img.id)
	}
	parent, err := ioutil.ReadFile(filepath.Join(dir, "imagedb/metadata/sha256", id.Hex(), "parent"))
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err != nil {
		if os.IsNotExist(err) {
			return []image{*img}, nil
		}
		return nil, err
	}

	parentID, err := digest.ParseDigest(string(parent))
	if err != nil {
		return nil, err
	}
	img.parent = parentID

	pc, err := b.getParentChain(ctx, dir, parentID)
	if err != nil {
		return nil, err
	}
	return append([]image{*img}, pc...), nil
}

type image struct {
	raw    []byte
	id     digest.Digest
	parent digest.Digest
	layers []digest.Digest
}

type manifestRow struct {
	Config string
	Parent string `json:",omitempty"`
	Layers []digest.Digest
}

func parseImage(in []byte) (*image, error) {
	var conf struct {
		RootFS struct {
			DiffIDs []digest.Digest `json:"diff_ids"`
		} `json:"rootfs"`
	}
	if err := json.Unmarshal(in, &conf); err != nil {
		return nil, err
	}
	return &image{
		layers: conf.RootFS.DiffIDs,
		raw:    in,
		id:     digest.FromBytes(in),
	}, nil
}

func validateParentChain(imgs []image) error {
	if len(imgs) < 2 {
		return nil
	}
	if err := validateParentChain(imgs[1:]); err != nil {
		return err
	}
	for i, l := range imgs[1].layers {
		if l != imgs[0].layers[i] {
			return fmt.Errorf("invalid layers in parent chain")
		}
	}
	return nil
}
