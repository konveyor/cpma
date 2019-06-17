package oauth

import (
	"encoding/base64"

	"github.com/fusor/cpma/pkg/transform/configmaps"
	"github.com/fusor/cpma/pkg/transform/secrets"
	legacyconfigv1 "github.com/openshift/api/legacyconfig/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// IdentityProviderBasicAuth is a basic auth specific identity provider
type IdentityProviderBasicAuth struct {
	identityProviderCommon `json:",inline"`
	BasicAuth              BasicAuth `json:"basicAuth"`
}

// BasicAuth provider specific data
// BasicAuth provider specific data
type BasicAuth struct {
	URL           string         `json:"url"`
	CA            *CA            `json:"ca,omitempty"`
	TLSClientCert *TLSClientCert `json:"tlsClientCert,omitempty"`
	TLSClientKey  *TLSClientKey  `json:"tlsClientKey,omitempty"`
}

func buildBasicAuthIP(serializer *json.Serializer, p IdentityProvider) (*IdentityProviderBasicAuth, *secrets.Secret, *secrets.Secret, *configmaps.ConfigMap, error) {
	var (
		err                   error
		idP                   = &IdentityProviderBasicAuth{}
		certSecret, keySecret *secrets.Secret
		caConfigmap           *configmaps.ConfigMap
		basicAuth             legacyconfigv1.BasicAuthPasswordIdentityProvider
	)

	if _, _, err = serializer.Decode(p.Provider.Raw, nil, &basicAuth); err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "Failed to decode basic auth, see error")
	}

	idP.Type = "BasicAuth"
	idP.Name = p.Name
	idP.Challenge = p.UseAsChallenger
	idP.Login = p.UseAsLogin
	idP.MappingMethod = p.MappingMethod
	idP.BasicAuth.URL = basicAuth.URL

	if basicAuth.CA != "" {
		caConfigmap = configmaps.GenConfigMap("basicauth-configmap", OAuthNamespace, p.CAData)
		idP.BasicAuth.CA = &CA{Name: caConfigmap.Metadata.Name}
	}

	if basicAuth.CertFile != "" {
		certSecretName := p.Name + "-client-cert-secret"
		idP.BasicAuth.TLSClientCert = &TLSClientCert{Name: certSecretName}

		encoded := base64.StdEncoding.EncodeToString(p.CrtData)
		if certSecret, err = secrets.GenSecret(certSecretName, encoded, OAuthNamespace, secrets.BasicAuthSecretType); err != nil {
			return nil, nil, nil, nil, errors.Wrap(err, "Failed to generate cert secret for basic auth, see error")
		}

		keySecretName := p.Name + "-client-key-secret"
		idP.BasicAuth.TLSClientKey = &TLSClientKey{Name: keySecretName}

		encoded = base64.StdEncoding.EncodeToString(p.KeyData)
		if keySecret, err = secrets.GenSecret(keySecretName, encoded, OAuthNamespace, secrets.BasicAuthSecretType); err != nil {
			return nil, nil, nil, nil, errors.Wrap(err, "Failed to generate key secret for basic auth, see error")
		}
	}

	return idP, certSecret, keySecret, caConfigmap, nil
}

func validateBasicAuthProvider(serializer *json.Serializer, p IdentityProvider) error {
	var basicAuth legacyconfigv1.BasicAuthPasswordIdentityProvider

	if _, _, err := serializer.Decode(p.Provider.Raw, nil, &basicAuth); err != nil {
		return errors.Wrap(err, "Failed to decode basic auth, see error")
	}

	if p.Name == "" {
		return errors.New("Name can't be empty")
	}

	if err := validateMappingMethod(p.MappingMethod); err != nil {
		return err
	}

	if basicAuth.URL == "" {
		return errors.New("URL can't be empty")
	}

	if basicAuth.CertFile != "" && basicAuth.KeyFile == "" {
		return errors.New("Key file can't be empty if cert file is specified")
	}

	return nil
}
