// Package registry handles the registry relation functionality such as pushing
// the container image to the provided registry, fetching registry credentials.
package registry

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type registryCredentials struct {
	provider string
	username string
	password string
}

func fetchCredentials() (*registryCredentials, error) {
	registryUsername := os.Getenv("REGISTRY_USERNAME")
	registryPassword := os.Getenv("REGISTRY_PASSWORD")
	registryProvider := os.Getenv("REGISTRY_PROVIDER")

	if len(registryUsername) == 0 || len(registryPassword) == 0 {
		return nil, fmt.Errorf("registry username or password cannot be empty")
	}

	return &registryCredentials{
		provider: registryProvider,
		username: registryUsername,
		password: registryPassword,
	}, nil
}

// GetDestinationImage returns the name of the destination image.
func GetDestinationImage(srcImage string) (string, error) {
	dstImage := ""

	creds, err := fetchCredentials()
	if err != nil {
		return dstImage, err
	}

	repo, tag := getRepoAndTagFromImage(srcImage)

	if len(creds.provider) != 0 {
		dstImage += fmt.Sprintf("%s/", creds.provider)
	}

	dstImage += fmt.Sprintf("%s/%s", creds.username, repo)

	if len(tag) != 0 {
		dstImage += fmt.Sprintf(":%s", tag)
	}

	return dstImage, nil
}

// Backup pushes the docker image to the provided repository.
func Backup(srcImage, dstImage string) error {
	srcRef, err := getReference(srcImage)
	if err != nil {
		return err
	}

	creds, err := fetchCredentials()
	if err != nil {
		return fmt.Errorf("failed to fetch credentials: %v", err)
	}

	authConfig := authn.AuthConfig{
		Username: creds.username,
		Password: creds.password,
	}

	auth := authn.FromConfig(authConfig)

	img, err := remote.Image(srcRef, remote.WithAuth(auth))
	if err != nil {
		return fmt.Errorf("failed to fetch image: %v", err)
	}

	dstRef, err := getReference(dstImage)
	if err != nil {
		return err
	}

	// Check if image:tag with latest digest already present. If yes then
	// there is no need to push the image.
	// If there is no error, it means atleast the image was present, may or not be
	// same as the source image.
	// If there is an error, it means either the image is not present in the backup repository
	// or network issue or some other issue. Point being we can go ahead and push the image.
	if dstimg, err := remote.Image(dstRef, remote.WithAuth(authn.Anonymous)); err == nil {
		dstHash, err := dstimg.Digest()
		if err != nil {
			return fmt.Errorf("destination image %q already present , failed to get digest: %v", dstImage, err)
		}

		srcHash, err := img.Digest()
		if err != nil {
			return fmt.Errorf("failed to get digest of source image %q: %v", srcImage, err)
		}

		if srcHash == dstHash {
			return nil
		}
	}

	if err = remote.Write(dstRef, img, remote.WithAuth(auth)); err != nil {
		return fmt.Errorf("failed to push image: %v", err)
	}

	return nil
}

func getRepoAndTagFromImage(image string) (repository, tag string) {
	if len(image) == 0 {
		return repository, tag
	}

	str := strings.Split(image, ":")

	imageWithoutTag := str[0]

	if len(str) == 1 {
		tag = ""
	} else {
		tag = str[len(str)-1]
	}

	str = strings.Split(imageWithoutTag, "/")
	repository = str[len(str)-1]

	return repository, tag
}

func getReference(image string) (name.Reference, error) {
	ref, err := name.ParseReference(image)
	if err != nil {
		return nil, fmt.Errorf("failed parsing image reference: %v", err)
	}

	return ref, nil
}
