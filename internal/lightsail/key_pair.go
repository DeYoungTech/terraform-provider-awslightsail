package lightsail

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/aws/smithy-go"
	"github.com/deyoungtech/terraform-provider-awslightsail/internal/helper/encryption"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceKeyPair() *schema.Resource {
	return &schema.Resource{
		Create: resourceKeyPairCreate,
		Read:   resourceKeyPairRead,
		Delete: resourceKeyPairDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},

			// optional fields
			"pgp_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			// additional info returned from the API
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// fields returned from CreateKey
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// encrypted fields if pgp_key is given
			"encrypted_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceKeyPairCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)

	var kName string
	if v, ok := d.GetOk("name"); ok {
		kName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		kName = resource.PrefixedUniqueId(v.(string))
	} else {
		kName = resource.UniqueId()
	}

	var pubKey string
	var op *types.Operation
	if pubKeyInterface, ok := d.GetOk("public_key"); ok {
		pubKey = pubKeyInterface.(string)
	}

	if pubKey == "" {
		// creating new key
		resp, err := conn.CreateKeyPair(context.TODO(), &lightsail.CreateKeyPairInput{
			KeyPairName: aws.String(kName),
		})
		if err != nil {
			return err
		}
		if resp.Operation == nil {
			return fmt.Errorf("No operation found for CreateKeyPair response")
		}
		if resp.KeyPair == nil {
			return fmt.Errorf("No KeyPair information found for CreateKeyPair response")
		}
		d.SetId(kName)

		// private_key and public_key are only available in the response from
		// CreateKey pair. Here we set the public_key, and encrypt the private_key
		// if a pgp_key is given, else we store the private_key in state
		d.Set("public_key", resp.PublicKeyBase64)

		// encrypt private key if pgp_key is given
		pgpKey, err := encryption.RetrieveGPGKey(d.Get("pgp_key").(string))
		if err != nil {
			return err
		}
		if pgpKey != "" {
			fingerprint, encrypted, err := encryption.EncryptValue(pgpKey, *resp.PrivateKeyBase64, "Lightsail Private Key")
			if err != nil {
				return err
			}

			d.Set("encrypted_fingerprint", fingerprint)
			d.Set("encrypted_private_key", encrypted)
		} else {
			d.Set("private_key", resp.PrivateKeyBase64)
		}

		op = resp.Operation

	} else {
		// importing key
		resp, err := conn.ImportKeyPair(context.TODO(), &lightsail.ImportKeyPairInput{
			KeyPairName:     aws.String(kName),
			PublicKeyBase64: aws.String(pubKey),
		})

		if err != nil {
			log.Printf("[ERR] Error importing key: %s", err)
			return err
		}
		d.SetId(kName)

		op = resp.Operation
	}

	err := waitLightsailOperation(conn, op.Id)

	if err != nil {
		// We don't return an error here because the Create call succeeded
		log.Printf("[ERR] Error waiting for KeyPair (%s) to become ready: %s", d.Id(), err)
	}

	return resourceKeyPairRead(d, meta)
}

func resourceKeyPairRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)

	resp, err := conn.GetKeyPair(context.TODO(), &lightsail.GetKeyPairInput{
		KeyPairName: aws.String(d.Id()),
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
		}
		d.SetId("")
		return nil
	}

	d.Set("arn", resp.KeyPair.Arn)
	d.Set("name", resp.KeyPair.Name)
	d.Set("fingerprint", resp.KeyPair.Fingerprint)

	return nil
}

func resourceKeyPairDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*lightsail.Client)
	resp, err := conn.DeleteKeyPair(context.TODO(), &lightsail.DeleteKeyPairInput{
		KeyPairName: aws.String(d.Id()),
	})

	if err != nil {
		return err
	}

	op := resp.Operation

	err = waitLightsailOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf(
			"Error waiting for KeyPair (%s) to become destroyed: %s",
			d.Id(), err)
	}

	return nil
}
