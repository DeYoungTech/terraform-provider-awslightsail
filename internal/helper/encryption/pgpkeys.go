package encryption

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/keybase/go-crypto/openpgp"
	"github.com/keybase/go-crypto/openpgp/packet"
)

const (
	kbPrefix = "keybase:"
)

// FetchKeybasePubkeys fetches public keys from Keybase given a set of
// usernames, which are derived from correctly formatted input entries. It
// doesn't use their client code due to both the API and the fact that it is
// considered alpha and probably best not to rely on it.  The keys are returned
// as base64-encoded strings.
func FetchKeybasePubkeys(input []string) (map[string]string, error) {
	client := cleanhttp.DefaultClient()
	if client == nil {
		return nil, fmt.Errorf("unable to create an http client")
	}

	if len(input) == 0 {
		return nil, nil
	}

	usernames := make([]string, 0, len(input))
	for _, v := range input {
		if strings.HasPrefix(v, kbPrefix) {
			usernames = append(usernames, strings.TrimPrefix(v, kbPrefix))
		}
	}

	if len(usernames) == 0 {
		return nil, nil
	}

	ret := make(map[string]string, len(usernames))
	url := fmt.Sprintf("https://keybase.io/_/api/1.0/user/lookup.json?usernames=%s&fields=public_keys", strings.Join(usernames, ","))
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type PublicKeys struct {
		Primary struct {
			Bundle string
		}
	}

	type LThem struct {
		PublicKeys `json:"public_keys"`
	}

	type KbResp struct {
		Status struct {
			Name string
		}
		Them []LThem
	}

	out := &KbResp{
		Them: []LThem{},
	}

	if err := DecodeJSONFromReader(resp.Body, out); err != nil {
		return nil, err
	}

	if out.Status.Name != "OK" {
		return nil, fmt.Errorf("got non-OK response: %q", out.Status.Name)
	}

	missingNames := make([]string, 0, len(usernames))
	var keyReader *bytes.Reader
	serializedEntity := bytes.NewBuffer(nil)
	for i, themVal := range out.Them {
		if themVal.Primary.Bundle == "" {
			missingNames = append(missingNames, usernames[i])
			continue
		}
		keyReader = bytes.NewReader([]byte(themVal.Primary.Bundle))
		entityList, err := openpgp.ReadArmoredKeyRing(keyReader)
		if err != nil {
			return nil, err
		}
		if len(entityList) != 1 {
			return nil, fmt.Errorf("primary key could not be parsed for user %q", usernames[i])
		}
		if entityList[0] == nil {
			return nil, fmt.Errorf("primary key was nil for user %q", usernames[i])
		}

		serializedEntity.Reset()
		err = entityList[0].Serialize(serializedEntity)
		if err != nil {
			return nil, fmt.Errorf("error serializing entity for user %q: %w", usernames[i], err)
		}

		// The API returns values in the same ordering requested, so this should properly match
		ret[kbPrefix+usernames[i]] = base64.StdEncoding.EncodeToString(serializedEntity.Bytes())
	}

	if len(missingNames) > 0 {
		return nil, fmt.Errorf("unable to fetch keys for user(s) %q from keybase", strings.Join(missingNames, ","))
	}

	return ret, nil
}

// EncryptShares takes an ordered set of byte slices to encrypt and the
// corresponding base64-encoded public keys to encrypt them with, encrypts each
// byte slice with the corresponding public key.
//
// Note: There is no corresponding test function; this functionality is
// thoroughly tested in the init and rekey command unit tests
func EncryptShares(input [][]byte, pgpKeys []string) ([]string, [][]byte, error) {
	if len(input) != len(pgpKeys) {
		return nil, nil, fmt.Errorf("mismatch between number items to encrypt and number of PGP keys")
	}
	encryptedShares := make([][]byte, 0, len(pgpKeys))
	entities, err := GetEntities(pgpKeys)
	if err != nil {
		return nil, nil, err
	}
	for i, entity := range entities {
		ctBuf := bytes.NewBuffer(nil)
		pt, err := openpgp.Encrypt(ctBuf, []*openpgp.Entity{entity}, nil, nil, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("error setting up encryption for PGP message: %w", err)
		}
		_, err = pt.Write(input[i])
		if err != nil {
			return nil, nil, fmt.Errorf("error encrypting PGP message: %w", err)
		}
		pt.Close()
		encryptedShares = append(encryptedShares, ctBuf.Bytes())
	}

	fingerprints, err := GetFingerprints(nil, entities)
	if err != nil {
		return nil, nil, err
	}

	return fingerprints, encryptedShares, nil
}

// GetFingerprints takes in a list of openpgp Entities and returns the
// fingerprints. If entities is nil, it will instead parse both entities and
// fingerprints from the pgpKeys string slice.
func GetFingerprints(pgpKeys []string, entities []*openpgp.Entity) ([]string, error) {
	if entities == nil {
		var err error
		entities, err = GetEntities(pgpKeys)

		if err != nil {
			return nil, err
		}
	}
	ret := make([]string, 0, len(entities))
	for _, entity := range entities {
		ret = append(ret, fmt.Sprintf("%x", entity.PrimaryKey.Fingerprint))
	}
	return ret, nil
}

// GetEntities takes in a string array of base64-encoded PGP keys and returns
// the openpgp Entities
func GetEntities(pgpKeys []string) ([]*openpgp.Entity, error) {
	ret := make([]*openpgp.Entity, 0, len(pgpKeys))
	for _, keystring := range pgpKeys {
		data, err := base64.StdEncoding.DecodeString(keystring)
		if err != nil {
			return nil, fmt.Errorf("error decoding given PGP key: %w", err)
		}
		entity, err := openpgp.ReadEntity(packet.NewReader(bytes.NewBuffer(data)))
		if err != nil {
			return nil, fmt.Errorf("error parsing given PGP key: %w", err)
		}
		ret = append(ret, entity)
	}
	return ret, nil
}

// Decodes/Unmarshals the given io.Reader pointing to a JSON, into a desired object
func DecodeJSONFromReader(r io.Reader, out interface{}) error {
	if r == nil {
		return fmt.Errorf("'io.Reader' being decoded is nil")
	}
	if out == nil {
		return fmt.Errorf("output parameter 'out' is nil")
	}

	dec := json.NewDecoder(r)

	// While decoding JSON values, interpret the integer values as `json.Number`s instead of `float64`.
	dec.UseNumber()

	// Since 'out' is an interface representing a pointer, pass it to the decoder without an '&'
	return dec.Decode(out)
}
