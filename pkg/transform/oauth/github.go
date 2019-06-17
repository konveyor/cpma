package oauth

import (
	"encoding/base64"

	"github.com/pkg/errors"

	"github.com/fusor/cpma/pkg/io"
	"github.com/fusor/cpma/pkg/transform/configmaps"
	"github.com/fusor/cpma/pkg/transform/secrets"
	legacyconfigv1 "github.com/openshift/api/legacyconfig/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

//IdentityProviderGitHub is a Github specific identity provider
type IdentityProviderGitHub struct {
	identityProviderCommon `json:",inline"`
	GitHub                 GitHub `json:"github"`
}

// GitHub provider specific data
type GitHub struct {
	HostName      string       `json:"hostname,omitempty"`
	CA            *CA          `json:"ca,omitempty"`
	ClientID      string       `json:"clientID"`
	ClientSecret  ClientSecret `json:"clientSecret"`
	Organizations []string     `json:"organizations,omitempty"`
	Teams         []string     `json:"teams,omitempty"`
}

func buildGitHubIP(serializer *json.Serializer, p IdentityProvider) (*IdentityProviderGitHub, *secrets.Secret, *configmaps.ConfigMap, error) {
	var (
		err         error
		idP         = &IdentityProviderGitHub{}
		secret      *secrets.Secret
		caConfigmap *configmaps.ConfigMap
		github      legacyconfigv1.GitHubIdentityProvider
	)

	if _, _, err = serializer.Decode(p.Provider.Raw, nil, &github); err != nil {
		return nil, nil, nil, errors.Wrap(err, "Something is wrong in decoding github")
	}

	idP.Type = "GitHub"
	idP.Name = p.Name
	idP.Challenge = p.UseAsChallenger
	idP.Login = p.UseAsLogin
	idP.MappingMethod = p.MappingMethod
	idP.GitHub.HostName = github.Hostname
	idP.GitHub.ClientID = github.ClientID
	idP.GitHub.Organizations = github.Organizations
	idP.GitHub.Teams = github.Teams

	if github.CA != "" {
		caConfigmap = configmaps.GenConfigMap("github-configmap", OAuthNamespace, p.CAData)
		idP.GitHub.CA = &CA{Name: caConfigmap.Metadata.Name}
	}

	secretName := p.Name + "-secret"
	idP.GitHub.ClientSecret.Name = secretName
	secretContent, err := io.FetchStringSource(github.ClientSecret)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Something is wrong in fetching client secret for github")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(secretContent))
	if secret, err = secrets.GenSecret(secretName, encoded, OAuthNamespace, secrets.LiteralSecretType); err != nil {
		return nil, nil, nil, errors.Wrap(err, "Something is wrong in generating secret for github")
	}

	return idP, secret, caConfigmap, nil
}

func validateGithubProvider(serializer *json.Serializer, p IdentityProvider) error {
	var github legacyconfigv1.GitHubIdentityProvider

	if _, _, err := serializer.Decode(p.Provider.Raw, nil, &github); err != nil {
		return errors.Wrap(err, "Something is wrong in decoding github")
	}

	if p.Name == "" {
		return errors.New("Name can't be empty")
	}

	if err := validateMappingMethod(p.MappingMethod); err != nil {
		return err
	}

	if github.ClientSecret.KeyFile != "" {
		return errors.New("Usage of encrypted files as secret value is not supported")
	}

	if err := validateClientData(github.ClientID, github.ClientSecret); err != nil {
		return err
	}

	return nil
}
